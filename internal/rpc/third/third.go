package third

import (
	"context"
	"fmt"
	"time"

	"server/v3/pkg/common/s3/disable"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/storage/cache"
	"server/v3/pkg/common/storage/cache/mcache"
	"server/v3/pkg/dbbuild"
	"server/v3/pkg/rpcli"

	"server/v3/pkg/common/s3/aws"
	"server/v3/pkg/common/s3/kodo"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/storage/cache/redis"
	"server/v3/pkg/common/storage/database/mgo"
	"server/v3/pkg/localcache"

	"server/v3/pkg/discovery"
	"server/v3/pkg/common/s3"
	"server/v3/pkg/common/s3/cos"
	"server/v3/pkg/common/s3/minio"
	"server/v3/pkg/common/s3/oss"
	"github.com/openimsdk/protocol/third"
	"google.golang.org/grpc"
	"server/v3/pkg/common/storage/controller"
)

type thirdServer struct {
	third.UnimplementedThirdServer
	thirdDatabase controller.ThirdDatabase
	s3dataBase    controller.S3Database
	defaultExpire time.Duration
	config        *Config
	s3            s3.Interface
	userClient    *rpcli.UserClient
}

type Config struct {
	RpcConfig          config.Third
	RedisConfig        config.Redis
	MongodbConfig      config.Mongo
	NotificationConfig config.Notification
	Share              config.Share
	MinioConfig        config.Minio
	LocalCacheConfig   config.LocalCache
	Discovery          config.Discovery
}

func Start(ctx context.Context, config *Config, client discovery.SvcDiscoveryRegistry, server grpc.ServiceRegistrar) error {
	dbb := dbbuild.NewBuilder(&config.MongodbConfig, &config.RedisConfig)
	mgocli, err := dbb.Mongo(ctx)
	if err != nil {
		return err
	}
	rdb, err := dbb.Redis(ctx)
	if err != nil {
		return err
	}

	logdb, err := mgo.NewLogMongo(mgocli.GetDB())
	if err != nil {
		return err
	}
	s3db, err := mgo.NewS3Mongo(mgocli.GetDB())
	if err != nil {
		return err
	}
	var thirdCache cache.ThirdCache
	if rdb == nil {
		tc, err := mgo.NewCacheMgo(mgocli.GetDB())
		if err != nil {
			return err
		}
		thirdCache = mcache.NewThirdCache(tc)
	} else {
		thirdCache = redis.NewThirdCache(rdb)
	}
	// Select the oss method according to the profile policy
	var o s3.Interface
	switch enable := config.RpcConfig.Object.Enable; enable {
	case "minio":
		var minioCache minio.Cache
		if rdb == nil {
			mc, err := mgo.NewCacheMgo(mgocli.GetDB())
			if err != nil {
				return err
			}
			minioCache = mcache.NewMinioCache(mc)
		} else {
			minioCache = redis.NewMinioCache(rdb)
		}
		o, err = minio.NewMinio(ctx, minioCache, *config.MinioConfig.Build())
	case "cos":
		o, err = cos.NewCos(*config.RpcConfig.Object.Cos.Build())
	case "oss":
		o, err = oss.NewOSS(*config.RpcConfig.Object.Oss.Build())
	case "kodo":
		o, err = kodo.NewKodo(*config.RpcConfig.Object.Kodo.Build())
	case "aws":
		o, err = aws.NewAws(*config.RpcConfig.Object.Aws.Build())
	case "":
		o = disable.NewDisable()
	default:
		err = fmt.Errorf("invalid object enable: %s", enable)
	}
	if err != nil {
		return err
	}
	userConn, err := client.GetConn(ctx, config.Discovery.RpcService.User)
	if err != nil {
		return err
	}
	localcache.InitLocalCache(&config.LocalCacheConfig)
	third.RegisterThirdServer(server, &thirdServer{
		thirdDatabase: controller.NewThirdDatabase(thirdCache, logdb),
		s3dataBase:    controller.NewS3Database(rdb, o, s3db),
		defaultExpire: time.Hour * 24 * 7,
		config:        config,
		s3:            o,
		userClient:    rpcli.NewUserClient(userConn),
	})
	return nil
}

func (t *thirdServer) FcmUpdateToken(ctx context.Context, req *third.FcmUpdateTokenReq) (resp *third.FcmUpdateTokenResp, err error) {
	err = t.thirdDatabase.FcmUpdateToken(ctx, req.Account, int(req.PlatformID), req.FcmToken, req.ExpireTime)
	if err != nil {
		return nil, err
	}
	return &third.FcmUpdateTokenResp{}, nil
}

func (t *thirdServer) SetAppBadge(ctx context.Context, req *third.SetAppBadgeReq) (resp *third.SetAppBadgeResp, err error) {
	if err := authverify.CheckAccess(ctx, req.UserID); err != nil {
		return nil, err
	}
	err = t.thirdDatabase.SetAppBadge(ctx, req.UserID, int(req.AppUnreadCount))
	if err != nil {
		return nil, err
	}
	return &third.SetAppBadgeResp{}, nil
}
