package cache

import (
	"context"
	"server/v3/pkg/common/s3"
	relationtb "server/v3/pkg/common/storage/model"
)

type ObjectCache interface {
	BatchDeleter
	CloneObjectCache() ObjectCache
	GetName(ctx context.Context, engine string, name string) (*relationtb.Object, error)
	DelObjectName(engine string, names ...string) ObjectCache
}

type S3Cache interface {
	BatchDeleter
	GetKey(ctx context.Context, engine string, key string) (*s3.ObjectInfo, error)
	DelS3Key(engine string, keys ...string) S3Cache
}

type RustFSCache interface {
	BatchDeleter
	GetImageObjectKeyInfo(ctx context.Context, key string, fn func(ctx context.Context) (*ObjectImageInfo, error)) (*ObjectImageInfo, error)
	GetThumbnailKey(ctx context.Context, key string, format string, width int, height int, fn func(ctx context.Context) (string, error)) (string, error)
	DelObjectImageInfoKey(keys ...string) RustFSCache
	DelImageThumbnailKey(key string, format string, width int, height int) RustFSCache
}

type ObjectImageInfo struct {
	IsImg  bool   `json:"isImg"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
	Etag   string `json:"etag"`
}
