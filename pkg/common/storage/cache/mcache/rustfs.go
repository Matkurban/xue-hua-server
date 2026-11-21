package mcache

import (
	"context"
	"time"

	"server/v3/pkg/common/s3/rustfs"
	"server/v3/pkg/common/storage/cache/cachekey"
	"server/v3/pkg/common/storage/database"
)

func NewRustFSCache(cache database.Cache) rustfs.Cache {
	return &rustfsCache{
		cache:      cache,
		expireTime: time.Hour * 24 * 7,
	}
}

type rustfsCache struct {
	cache      database.Cache
	expireTime time.Duration
}

func (g *rustfsCache) getObjectImageInfoKey(key string) string {
	return cachekey.GetObjectImageInfoKey(key)
}

func (g *rustfsCache) getImageThumbnailKey(key string, format string, width int, height int) string {
	return cachekey.GetObjectImageThumbnailKey(key, format, width, height)
}

func (g *rustfsCache) DelObjectImageInfoKey(ctx context.Context, keys ...string) error {
	ks := make([]string, 0, len(keys))
	for _, key := range keys {
		ks = append(ks, g.getObjectImageInfoKey(key))
	}
	return g.cache.Del(ctx, ks)
}

func (g *rustfsCache) DelImageThumbnailKey(ctx context.Context, key string, format string, width int, height int) error {
	return g.cache.Del(ctx, []string{g.getImageThumbnailKey(key, format, width, height)})
}

func (g *rustfsCache) GetImageObjectKeyInfo(ctx context.Context, key string, fn func(ctx context.Context) (*rustfs.ImageInfo, error)) (*rustfs.ImageInfo, error) {
	return getCache[*rustfs.ImageInfo](ctx, g.cache, g.getObjectImageInfoKey(key), g.expireTime, fn)
}

func (g *rustfsCache) GetThumbnailKey(ctx context.Context, key string, format string, width int, height int, fn func(ctx context.Context) (string, error)) (string, error) {
	return getCache[string](ctx, g.cache, g.getImageThumbnailKey(key, format, width, height), g.expireTime, fn)
}
