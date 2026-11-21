package rustfs

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	aws3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/s3"
)

const (
	defaultRegion = "us-east-1"

	minPartSize int64 = 1024 * 1024 * 5        // 5MB
	maxPartSize int64 = 1024 * 1024 * 1024 * 5 // 5GB
	maxNumSize  int64 = 10000

	maxImageWidth      = 1024
	maxImageHeight     = 1024
	maxImageSize       = 1024 * 1024 * 50
	imageThumbnailPath = "openim/thumbnail"

	successCode = http.StatusOK
)

var _ s3.Interface = (*RustFS)(nil)

// Config holds connection and signing settings for a RustFS S3-compatible server.
type Config struct {
	Bucket          string
	Region          string
	Endpoint        string
	SignEndpoint    string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	PublicRead      bool
}

type RustFS struct {
	conf         Config
	bucket       string
	region       string
	signEndpoint string
	signPrefix   string
	client       *aws3.Client
	presign      *aws3.PresignClient
	cache        Cache

	initMu sync.Mutex
	inited bool
}

// NewRustFS initializes a RustFS client using the official AWS SDK for Go v2
// with path-style addressing as recommended by RustFS.
func NewRustFS(ctx context.Context, cache Cache, conf Config) (*RustFS, error) {
	r, err := newClient(cache, conf)
	if err != nil {
		return nil, err
	}
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

// Check verifies connectivity and credentials against the configured RustFS server.
func Check(ctx context.Context, conf *Config) error {
	if conf == nil {
		return errs.New("rustfs config is nil").Wrap()
	}
	r, err := newClient(nil, *conf)
	if err != nil {
		return err
	}
	if _, err := r.client.HeadBucket(ctx, &aws3.HeadBucketInput{Bucket: aws.String(r.bucket)}); err != nil {
		if r.IsNotFound(err) {
			// Server is reachable and authenticated; bucket will be created on service startup.
			return nil
		}
		return errs.WrapMsg(err, "rustfs health check failed", "endpoint", conf.Endpoint, "bucket", conf.Bucket)
	}
	return nil
}

func newClient(cache Cache, conf Config) (*RustFS, error) {
	if conf.Bucket == "" {
		return nil, errs.New("rustfs bucket is empty").Wrap()
	}
	if conf.Endpoint == "" {
		return nil, errs.New("rustfs endpoint is empty").Wrap()
	}
	region := conf.Region
	if region == "" {
		region = defaultRegion
	}

	internalBase, _, err := splitEndpoint(conf.Endpoint)
	if err != nil {
		return nil, errs.WrapMsg(err, "invalid rustfs endpoint", "endpoint", conf.Endpoint)
	}

	signRaw := conf.SignEndpoint
	if signRaw == "" {
		signRaw = conf.Endpoint
	}
	signBase, signPrefix, err := splitEndpoint(signRaw)
	if err != nil {
		return nil, errs.WrapMsg(err, "invalid rustfs signEndpoint", "signEndpoint", signRaw)
	}

	creds := credentials.NewStaticCredentialsProvider(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken)
	awsCfg := aws.Config{
		Region:      region,
		Credentials: creds,
	}

	internalClient := aws3.NewFromConfig(awsCfg, func(o *aws3.Options) {
		o.BaseEndpoint = aws.String(internalBase)
		o.UsePathStyle = true
	})

	var signClient *aws3.Client
	if signBase == internalBase {
		signClient = internalClient
	} else {
		signClient = aws3.NewFromConfig(awsCfg, func(o *aws3.Options) {
			o.BaseEndpoint = aws.String(signBase)
			o.UsePathStyle = true
		})
	}

	return &RustFS{
		conf:         conf,
		bucket:       conf.Bucket,
		region:       region,
		signEndpoint: signBase,
		signPrefix:   signPrefix,
		client:       internalClient,
		presign:      aws3.NewPresignClient(signClient),
		cache:        cache,
	}, nil
}

func splitEndpoint(raw string) (base string, prefix string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("endpoint %q must include scheme and host", raw)
	}
	prefix = strings.TrimRight(u.Path, "/")
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), prefix, nil
}

func (r *RustFS) ensureBucket(ctx context.Context) error {
	r.initMu.Lock()
	defer r.initMu.Unlock()
	if r.inited {
		return nil
	}

	_, err := r.client.HeadBucket(ctx, &aws3.HeadBucketInput{Bucket: aws.String(r.bucket)})
	if err != nil {
		if !r.IsNotFound(err) {
			return fmt.Errorf("check bucket exists error: %w", err)
		}
		if _, err = r.client.CreateBucket(ctx, &aws3.CreateBucketInput{Bucket: aws.String(r.bucket)}); err != nil {
			return fmt.Errorf("create bucket error: %w", err)
		}
	}

	if r.conf.PublicRead {
		policy := fmt.Sprintf(
			`{"Version":"2012-10-17","Statement":[{"Sid":"PublicRead","Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
			r.bucket,
		)
		if _, err = r.client.PutBucketPolicy(ctx, &aws3.PutBucketPolicyInput{
			Bucket: aws.String(r.bucket),
			Policy: aws.String(policy),
		}); err != nil {
			return fmt.Errorf("set bucket public-read policy error: %w", err)
		}
	}

	r.inited = true
	return nil
}

func (r *RustFS) Engine() string {
	return "rustfs"
}

func (r *RustFS) PartLimit() (*s3.PartLimit, error) {
	return &s3.PartLimit{
		MinPartSize: minPartSize,
		MaxPartSize: maxPartSize,
		MaxNumSize:  maxNumSize,
	}, nil
}

func (r *RustFS) PartSize(_ context.Context, size int64) (int64, error) {
	if size <= 0 {
		return 0, errors.New("size must be greater than 0")
	}
	if size > maxPartSize*maxNumSize {
		return 0, errors.New("rustfs object size exceeds maximum allowed limit")
	}
	if size <= minPartSize*maxNumSize {
		return minPartSize, nil
	}
	partSize := size / maxNumSize
	if size%maxNumSize != 0 {
		partSize++
	}
	return partSize, nil
}

func formatETag(etag string) string {
	return strings.ToLower(strings.Trim(etag, `"`))
}

func (r *RustFS) applyPrefix(rawURL string) (string, error) {
	if r.signPrefix == "" {
		return rawURL, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(r.signPrefix, u.Path)
	return u.String(), nil
}

func (r *RustFS) InitiateMultipartUpload(ctx context.Context, name string, opt *s3.PutOption) (*s3.InitiateMultipartUploadResult, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	res, err := r.client.CreateMultipartUpload(ctx, &aws3.CreateMultipartUploadInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	if res.UploadId == nil || *res.UploadId == "" {
		return nil, errors.New("CreateMultipartUpload returned empty upload id")
	}
	return &s3.InitiateMultipartUploadResult{
		Bucket:   r.bucket,
		Key:      name,
		UploadID: *res.UploadId,
	}, nil
}

func (r *RustFS) CompleteMultipartUpload(ctx context.Context, uploadID string, name string, parts []s3.Part) (*s3.CompleteMultipartUploadResult, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	completedParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(part.PartNumber)),
			ETag:       aws.String(formatETag(part.ETag)),
		}
	}
	res, err := r.client.CompleteMultipartUpload(ctx, &aws3.CompleteMultipartUploadInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(name),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return nil, err
	}
	if res.ETag == nil || *res.ETag == "" {
		return nil, errors.New("CompleteMultipartUpload returned empty etag")
	}
	r.delObjectImageInfoKey(ctx, name, 0)
	info := &s3.CompleteMultipartUploadResult{
		Bucket: r.bucket,
		Key:    name,
		ETag:   formatETag(*res.ETag),
	}
	if res.Location != nil {
		info.Location = *res.Location
	}
	return info, nil
}

func (r *RustFS) AbortMultipartUpload(ctx context.Context, uploadID string, name string) error {
	if err := r.ensureBucket(ctx); err != nil {
		return err
	}
	_, err := r.client.AbortMultipartUpload(ctx, &aws3.AbortMultipartUploadInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(name),
		UploadId: aws.String(uploadID),
	})
	return err
}

func (r *RustFS) ListUploadedParts(ctx context.Context, uploadID string, name string, partNumberMarker int, maxParts int) (*s3.ListUploadedPartsResult, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	input := &aws3.ListPartsInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(name),
		UploadId: aws.String(uploadID),
	}
	if partNumberMarker > 0 {
		input.PartNumberMarker = aws.String(strconv.Itoa(partNumberMarker))
	}
	if maxParts > 0 {
		input.MaxParts = aws.Int32(int32(maxParts))
	}
	res, err := r.client.ListParts(ctx, input)
	if err != nil {
		return nil, err
	}
	out := &s3.ListUploadedPartsResult{
		Key:           name,
		UploadID:      uploadID,
		UploadedParts: make([]s3.UploadedPart, 0, len(res.Parts)),
	}
	if res.MaxParts != nil {
		out.MaxParts = int(*res.MaxParts)
	}
	if res.NextPartNumberMarker != nil {
		out.NextPartNumberMarker, _ = strconv.Atoi(*res.NextPartNumberMarker)
	}
	for _, part := range res.Parts {
		var item s3.UploadedPart
		if part.PartNumber != nil {
			item.PartNumber = int(*part.PartNumber)
		}
		if part.LastModified != nil {
			item.LastModified = *part.LastModified
		}
		if part.ETag != nil {
			item.ETag = formatETag(*part.ETag)
		}
		if part.Size != nil {
			item.Size = *part.Size
		}
		out.UploadedParts = append(out.UploadedParts, item)
	}
	return out, nil
}

func (r *RustFS) AuthSign(ctx context.Context, uploadID string, name string, expire time.Duration, partNumbers []int) (*s3.AuthSignResult, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	res := &s3.AuthSignResult{
		Parts: make([]s3.SignPart, 0, len(partNumbers)),
	}
	params := &aws3.UploadPartInput{
		Bucket:   aws.String(r.bucket),
		Key:      aws.String(name),
		UploadId: aws.String(uploadID),
	}
	presignOpt := aws3.WithPresignExpires(expire)
	for _, number := range partNumbers {
		params.PartNumber = aws.Int32(int32(number))
		val, err := r.presign.PresignUploadPart(ctx, params, presignOpt)
		if err != nil {
			return nil, err
		}
		u, err := url.Parse(val.URL)
		if err != nil {
			return nil, err
		}
		if r.signPrefix != "" {
			u.Path = path.Join(r.signPrefix, u.Path)
		}
		query := u.Query()
		u.RawQuery = ""
		baseURL := u.String()
		if res.URL == "" {
			res.URL = baseURL
		}
		partURL := baseURL
		if res.URL == baseURL {
			partURL = ""
		}
		res.Parts = append(res.Parts, s3.SignPart{
			PartNumber: number,
			URL:        partURL,
			Query:      query,
			Header:     val.SignedHeader,
		})
	}
	return res, nil
}

func (r *RustFS) PresignedPutObject(ctx context.Context, name string, expire time.Duration, opt *s3.PutOption) (*s3.PresignedPutResult, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	res, err := r.presign.PresignPutObject(
		ctx,
		&aws3.PutObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(name)},
		aws3.WithPresignExpires(expire),
		withPresignQuery(nil),
	)
	if err != nil {
		return nil, err
	}
	signedURL, err := r.applyPrefix(res.URL)
	if err != nil {
		return nil, err
	}
	return &s3.PresignedPutResult{URL: signedURL}, nil
}

func (r *RustFS) DeleteObject(ctx context.Context, name string) error {
	if err := r.ensureBucket(ctx); err != nil {
		return err
	}
	info, err := r.StatObject(ctx, name)
	if err != nil {
		if r.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.clearObjectImageInfoKey(ctx, name, info.ETag); err != nil {
		return err
	}
	_, err = r.client.DeleteObject(ctx, &aws3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (r *RustFS) StatObject(ctx context.Context, name string) (*s3.ObjectInfo, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	res, err := r.client.HeadObject(ctx, &aws3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	if res.ETag == nil || *res.ETag == "" {
		return nil, errors.New("HeadObject returned empty etag")
	}
	if res.ContentLength == nil {
		return nil, errors.New("HeadObject returned nil content length")
	}
	info := &s3.ObjectInfo{
		ETag: formatETag(*res.ETag),
		Key:  name,
		Size: *res.ContentLength,
	}
	if res.LastModified != nil {
		info.LastModified = *res.LastModified
	} else {
		info.LastModified = time.Unix(0, 0)
	}
	return info, nil
}

func (r *RustFS) CopyObject(ctx context.Context, src string, dst string) (*s3.CopyObjectInfo, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	res, err := r.client.CopyObject(ctx, &aws3.CopyObjectInput{
		Bucket:     aws.String(r.bucket),
		CopySource: aws.String(r.bucket + "/" + url.PathEscape(src)),
		Key:        aws.String(dst),
	})
	if err != nil {
		return nil, err
	}
	if res.CopyObjectResult == nil || res.CopyObjectResult.ETag == nil || *res.CopyObjectResult.ETag == "" {
		return nil, errors.New("CopyObject returned empty etag")
	}
	return &s3.CopyObjectInfo{
		Key:  dst,
		ETag: formatETag(*res.CopyObjectResult.ETag),
	}, nil
}

func (r *RustFS) IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var respErr *awshttp.ResponseError
	if errors.As(err, &respErr) && respErr != nil && respErr.Response != nil {
		return respErr.Response.StatusCode == http.StatusNotFound
	}
	return false
}

func (r *RustFS) PresignedGetObject(ctx context.Context, name string, expire time.Duration, query url.Values) (string, error) {
	if r.conf.PublicRead {
		u, err := url.Parse(r.signEndpoint)
		if err != nil {
			return "", err
		}
		u.Path = path.Join(r.signPrefix, r.bucket, name)
		if len(query) > 0 {
			u.RawQuery = query.Encode()
		}
		return u.String(), nil
	}

	if expire <= 0 {
		expire = time.Hour * 24 * 7 // SigV4 max presign expiration is 7 days
	} else if expire < time.Second {
		expire = time.Second
	} else if expire > time.Hour*24*7 {
		expire = time.Hour * 24 * 7
	}

	res, err := r.presign.PresignGetObject(
		ctx,
		&aws3.GetObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(name)},
		aws3.WithPresignExpires(expire),
		withPresignQuery(query),
	)
	if err != nil {
		return "", err
	}
	return r.applyPrefix(res.URL)
}

func (r *RustFS) AccessURL(ctx context.Context, name string, expire time.Duration, opt *s3.AccessURLOption) (string, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return "", err
	}
	if opt == nil {
		return r.PresignedGetObject(ctx, name, expire, nil)
	}
	if opt.Image != nil {
		opt.Filename = ""
		opt.ContentType = ""
	}
	reqParams := make(url.Values)
	if opt.ContentType != "" {
		reqParams.Set("response-content-type", opt.ContentType)
	}
	if opt.Filename != "" {
		reqParams.Set("response-content-disposition", `attachment; filename*=UTF-8''`+url.PathEscape(opt.Filename))
	}
	if opt.Image == nil ||
		(opt.Image.Width < 0 && opt.Image.Height < 0 && opt.Image.Format == "") ||
		opt.Image.Width > maxImageWidth || opt.Image.Height > maxImageHeight {
		return r.PresignedGetObject(ctx, name, expire, reqParams)
	}
	return r.getImageThumbnailURL(ctx, name, expire, opt.Image)
}

func (r *RustFS) getObjectData(ctx context.Context, name string, limit int64) ([]byte, error) {
	out, err := r.client.GetObject(ctx, &aws3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	if limit < 0 {
		return io.ReadAll(out.Body)
	}
	return io.ReadAll(io.LimitReader(out.Body, limit))
}

// FormData generates a standard AWS S3 SigV4 browser POST upload policy for RustFS.
func (r *RustFS) FormData(ctx context.Context, name string, size int64, contentType string, duration time.Duration) (*s3.FormData, error) {
	if err := r.ensureBucket(ctx); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	expires := now.Add(duration)
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", r.conf.AccessKeyID, dateStamp, r.region)

	conditions := []any{
		map[string]string{"bucket": r.bucket},
		map[string]string{"key": name},
		map[string]string{"success_action_status": strconv.Itoa(successCode)},
		map[string]string{"x-amz-algorithm": "AWS4-HMAC-SHA256"},
		map[string]string{"x-amz-credential": credential},
		map[string]string{"x-amz-date": amzDate},
	}
	if size > 0 {
		conditions = append(conditions, []any{"content-length-range", 0, size})
	}
	if contentType != "" {
		conditions = append(conditions, map[string]string{"Content-Type": contentType})
	}
	if r.conf.SessionToken != "" {
		conditions = append(conditions, map[string]string{"x-amz-security-token": r.conf.SessionToken})
	}

	policyDoc := map[string]any{
		"expiration": expires.Format("2006-01-02T15:04:05.000Z"),
		"conditions": conditions,
	}
	policyJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return nil, err
	}
	policyB64 := base64.StdEncoding.EncodeToString(policyJSON)
	signature := signPolicyV4(r.conf.SecretAccessKey, dateStamp, r.region, policyB64)

	fd := map[string]string{
		"key":                   name,
		"bucket":                r.bucket,
		"success_action_status": strconv.Itoa(successCode),
		"x-amz-algorithm":       "AWS4-HMAC-SHA256",
		"x-amz-credential":      credential,
		"x-amz-date":            amzDate,
		"policy":                policyB64,
		"x-amz-signature":       signature,
	}
	if contentType != "" {
		fd["Content-Type"] = contentType
	}
	if r.conf.SessionToken != "" {
		fd["x-amz-security-token"] = r.conf.SessionToken
	}

	u, err := url.Parse(r.signEndpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(r.signPrefix, r.bucket)

	return &s3.FormData{
		URL:          u.String(),
		File:         "file",
		Header:       nil,
		FormData:     fd,
		Expires:      expires,
		SuccessCodes: []int{successCode},
	}, nil
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func signPolicyV4(secretKey, dateStamp, region, policyB64 string) string {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	return hex.EncodeToString(hmacSHA256(kSigning, policyB64))
}

func (r *RustFS) GetImageThumbnailKey(ctx context.Context, name string) (string, error) {
	info, img, err := r.getObjectImageInfo(ctx, name)
	if err != nil {
		return "", errs.Wrap(err)
	}
	if !info.IsImg {
		return "", errs.New("object not image").Wrap()
	}
	thumbnailWidth, thumbnailHeight := getThumbnailSize(img)
	return path.Join(imageThumbnailPath, info.Etag, fmt.Sprintf("image_w%d_h%d.%s", thumbnailWidth, thumbnailHeight, info.Format)), nil
}

func (r *RustFS) clearObjectImageInfoKey(ctx context.Context, name string, etag string) error {
	if err := r.ensureBucket(ctx); err != nil {
		return err
	}
	prefix := path.Join(imageThumbnailPath, etag) + "/"
	paginator := aws3.NewListObjectsV2Paginator(r.client, &aws3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})
	var clearedInfo bool
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			if !clearedInfo && r.cache != nil {
				clearedInfo = true
				if err := r.cache.DelObjectImageInfoKey(ctx, etag); err != nil {
					return err
				}
			}
			if r.cache != nil {
				if err := r.deleteThumbnailCache(ctx, name, *obj.Key); err != nil {
					return err
				}
			}
			if _, err := r.client.DeleteObject(ctx, &aws3.DeleteObjectInput{
				Bucket: aws.String(r.bucket),
				Key:    obj.Key,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RustFS) deleteThumbnailCache(ctx context.Context, name string, key string) error {
	filename := path.Base(key)
	ext := path.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	filename = strings.TrimPrefix(filename, "image_w")
	var width, height int
	if arr := strings.Split(filename, "_h"); len(arr) == 2 {
		width, _ = strconv.Atoi(arr[0])
		height, _ = strconv.Atoi(arr[1])
	}
	format := strings.Trim(ext, ".")
	return r.cache.DelImageThumbnailKey(ctx, name, format, width, height)
}

func withPresignQuery(query url.Values) func(*aws3.PresignOptions) {
	return func(options *aws3.PresignOptions) {
		options.Presigner = &presignQueryWrapper{
			query:     query,
			presigner: options.Presigner,
		}
	}
}

type presignQueryWrapper struct {
	query     url.Values
	presigner aws3.HTTPPresignerV4
}

func (w *presignQueryWrapper) PresignHTTP(
	ctx context.Context,
	creds aws.Credentials,
	req *http.Request,
	payloadHash string,
	service string,
	region string,
	signingTime time.Time,
	optFns ...func(*v4.SignerOptions),
) (string, http.Header, error) {
	optFns = append(optFns, func(o *v4.SignerOptions) {
		o.DisableHeaderHoisting = true
	})
	req.Header.Del("Amz-Sdk-Request")
	if len(w.query) > 0 {
		q := req.URL.Query()
		for k, vals := range w.query {
			for _, v := range vals {
				q.Set(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	return w.presigner.PresignHTTP(ctx, creds, req, payloadHash, service, region, signingTime, optFns...)
}
