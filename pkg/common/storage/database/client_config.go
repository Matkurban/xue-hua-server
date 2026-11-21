package database

import (
	"context"

	"server/v3/pkg/common/storage/pagination"
	"server/v3/pkg/common/storage/model"
)

type ClientConfig interface {
	Set(ctx context.Context, userID string, config map[string]string) error
	Get(ctx context.Context, userID string) (map[string]string, error)
	Del(ctx context.Context, userID string, keys []string) error
	GetPage(ctx context.Context, userID string, key string, pagination pagination.Pagination) (int64, []*model.ClientConfig, error)
}
