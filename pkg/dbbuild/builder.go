package dbbuild

import (
	"context"

	"server/v3/pkg/common/storage/mongoutil"
	"github.com/redis/go-redis/v9"
	"server/v3/pkg/common/config"
)

type Builder interface {
	Mongo(ctx context.Context) (*mongoutil.Client, error)
	Redis(ctx context.Context) (redis.UniversalClient, error)
}

func NewBuilder(mongoConf *config.Mongo, redisConf *config.Redis) Builder {
	if config.Standalone() {
		globalStandalone.setConfig(mongoConf, redisConf)
		return globalStandalone
	}
	return &microservices{
		mongo: mongoConf,
		redis: redisConf,
	}
}
