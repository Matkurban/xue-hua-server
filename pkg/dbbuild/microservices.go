package dbbuild

import (
	"context"

	"server/v3/pkg/common/storage/mongoutil"
	"server/v3/pkg/common/storage/redisutil"
	"github.com/redis/go-redis/v9"
	"server/v3/pkg/common/config"
)

type microservices struct {
	mongo *config.Mongo
	redis *config.Redis
}

func (x *microservices) Mongo(ctx context.Context) (*mongoutil.Client, error) {
	return mongoutil.NewMongoDB(ctx, x.mongo.Build())
}

func (x *microservices) Redis(ctx context.Context) (redis.UniversalClient, error) {
	if x.redis.Disable {
		return nil, nil
	}
	return redisutil.NewRedisClient(ctx, x.redis.Build())
}
