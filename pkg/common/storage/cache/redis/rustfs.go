package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"server/v3/pkg/common/s3/rustfs"
	"server/v3/pkg/common/storage/cache"
	"server/v3/pkg/common/storage/cache/cachekey"
)

func NewRustFSCache(rdb redis.UniversalClient) rustfs.Cache {
	rc := newRocksCacheClient(rdb)
	return &rustfsCacheRedis{
		BatchDeleter: rc.GetBatchDeleter(),
		rcClient:     rc,
		expireTime:   time.Hour * 24 * 7,
	}
}

type rustfsCacheRedis struct {
	cache.BatchDeleter
	rcClient   *rocksCacheClient
	expireTime time.Duration
}

func (g *rustfsCacheRedis) getObjectImageInfoKey(key string) string {
	return cachekey.GetObjectImageInfoKey(key)
}

func (g *rustfsCacheRedis) getImageThumbnailKey(key string, format string, width int, height int) string {
	return cachekey.GetObjectImageThumbnailKey(key, format, width, height)
}

func (g *rustfsCacheRedis) DelObjectImageInfoKey(ctx context.Context, keys ...string) error {
	ks := make([]string, 0, len(keys))
	for _, key := range keys {
		ks = append(ks, g.getObjectImageInfoKey(key))
	}
	return g.BatchDeleter.ExecDelWithKeys(ctx, ks)
}

func (g *rustfsCacheRedis) DelImageThumbnailKey(ctx context.Context, key string, format string, width int, height int) error {
	return g.BatchDeleter.ExecDelWithKeys(ctx, []string{g.getImageThumbnailKey(key, format, width, height)})
}

func (g *rustfsCacheRedis) GetImageObjectKeyInfo(ctx context.Context, key string, fn func(ctx context.Context) (*rustfs.ImageInfo, error)) (*rustfs.ImageInfo, error) {
	info, err := getCache(ctx, g.rcClient, g.getObjectImageInfoKey(key), g.expireTime, fn)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (g *rustfsCacheRedis) GetThumbnailKey(ctx context.Context, key string, format string, width int, height int, fn func(ctx context.Context) (string, error)) (string, error) {
	return getCache(ctx, g.rcClient, g.getImageThumbnailKey(key, format, width, height), g.expireTime, fn)
}
