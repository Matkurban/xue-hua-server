package rustfs

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/log"
	"server/v3/pkg/common/s3"
)

func (r *RustFS) getImageThumbnailURL(ctx context.Context, name string, expire time.Duration, opt *s3.Image) (string, error) {
	var img image.Image
	info, err := r.cache.GetImageObjectKeyInfo(ctx, name, func(ctx context.Context) (info *ImageInfo, err error) {
		info, img, err = r.getObjectImageInfo(ctx, name)
		return
	})
	if err != nil {
		return "", err
	}
	if !info.IsImg {
		return "", errs.New("object not image").Wrap()
	}

	width := opt.Width
	if width > info.Width || width <= 0 {
		width = info.Width
	}
	height := opt.Height
	if height > info.Height || height <= 0 {
		height = info.Height
	}

	format := strings.ToLower(opt.Format)
	if format == formatJpg {
		format = formatJpeg
	}
	switch format {
	case formatPng, formatJpeg, formatGif:
	default:
		format = formatPng
	}

	reqParams := make(url.Values)
	if width == info.Width && height == info.Height && (format == info.Format || format == "") {
		reqParams.Set("response-content-type", "image/"+info.Format)
		return r.PresignedGetObject(ctx, name, expire, reqParams)
	}

	key, err := r.cache.GetThumbnailKey(ctx, name, format, width, height, func(ctx context.Context) (string, error) {
		if img == nil {
			out, err := r.client.GetObject(ctx, &aws3.GetObjectInput{
				Bucket: aws.String(r.bucket),
				Key:    aws.String(name),
			})
			if err != nil {
				return "", err
			}
			defer out.Body.Close()
			img, _, err = ImageStat(out.Body)
			if err != nil {
				return "", err
			}
		}

		thumbnail := resizeImage(img, width, height)
		var buf bytes.Buffer
		switch format {
		case formatPng:
			err = png.Encode(&buf, thumbnail)
		case formatJpeg:
			err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 40})
		case formatGif:
			err = gif.Encode(&buf, thumbnail, nil)
		}
		if err != nil {
			return "", errs.WrapMsg(err, "encode failed", "type", format)
		}

		cacheKey := path.Join(imageThumbnailPath, info.Etag, fmt.Sprintf("image_w%d_h%d.%s", width, height, format))
		contentType := "image/" + format
		if _, err = r.client.PutObject(ctx, &aws3.PutObjectInput{
			Bucket:        aws.String(r.bucket),
			Key:           aws.String(cacheKey),
			Body:          bytes.NewReader(buf.Bytes()),
			ContentLength: aws.Int64(int64(buf.Len())),
			ContentType:   aws.String(contentType),
		}); err != nil {
			return "", err
		}
		return cacheKey, nil
	})
	if err != nil {
		return "", err
	}

	reqParams.Set("response-content-type", "image/"+format)
	return r.PresignedGetObject(ctx, key, expire, reqParams)
}

func (r *RustFS) getObjectImageInfo(ctx context.Context, name string) (*ImageInfo, image.Image, error) {
	fileInfo, err := r.StatObject(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	if fileInfo.Size > maxImageSize {
		return nil, nil, errs.New("file size too large").Wrap()
	}
	imageData, err := r.getObjectData(ctx, name, fileInfo.Size)
	if err != nil {
		return nil, nil, err
	}
	var info ImageInfo
	imageInfo, format, err := ImageStat(bytes.NewReader(imageData))
	if err == nil {
		info.IsImg = true
		info.Format = format
		info.Width, info.Height = ImageWidthHeight(imageInfo)
	} else {
		info.IsImg = false
	}
	info.Etag = fileInfo.ETag
	return &info, imageInfo, nil
}

func (r *RustFS) delObjectImageInfoKey(ctx context.Context, key string, size int64) {
	if r.cache == nil {
		return
	}
	if size > 0 && size > maxImageSize {
		return
	}
	if err := r.cache.DelObjectImageInfoKey(ctx, key); err != nil {
		log.ZError(ctx, "DelObjectImageInfoKey failed", err, "key", key)
	}
}
