package push

import (
	"context"
	"gitee.com/matkurban/openimsdk-tools/mq"
	"math/rand"
	"strconv"

	"gitee.com/matkurban/openimsdk-tools/discovery"
	"gitee.com/matkurban/openimsdk-tools/log"
	"gitee.com/matkurban/openimsdk-tools/mcontext"
	pbpush "github.com/openimsdk/protocol/push"
	"google.golang.org/grpc"
	"server/v3/internal/push/offlinepush"
	"server/v3/pkg/authverify"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/storage/cache"
	"server/v3/pkg/common/storage/cache/mcache"
	"server/v3/pkg/common/storage/cache/redis"
	"server/v3/pkg/common/storage/controller"
	"server/v3/pkg/common/storage/database/mgo"
	"server/v3/pkg/dbbuild"
	"server/v3/pkg/mqbuild"
)

type pushServer struct {
	pbpush.UnimplementedPushMsgServiceServer
	database      controller.PushDatabase
	disCov        discovery.Conn
	offlinePusher offlinepush.OfflinePusher
}

type Config struct {
	RpcConfig          config.Push
	RedisConfig        config.Redis
	MongoConfig        config.Mongo
	KafkaConfig        config.Kafka
	NotificationConfig config.Notification
	Share              config.Share
	WebhooksConfig     config.Webhooks
	LocalCacheConfig   config.LocalCache
	Discovery          config.Discovery
	FcmConfigPath      config.Path
}

func (p pushServer) DelUserPushToken(ctx context.Context,
	req *pbpush.DelUserPushTokenReq) (resp *pbpush.DelUserPushTokenResp, err error) {
	if err = p.database.DelFcmToken(ctx, req.UserID, int(req.PlatformID)); err != nil {
		return nil, err
	}
	return &pbpush.DelUserPushTokenResp{}, nil
}

func Start(ctx context.Context, config *Config, client discovery.SvcDiscoveryRegistry, server grpc.ServiceRegistrar) error {
	dbb := dbbuild.NewBuilder(&config.MongoConfig, &config.RedisConfig)
	rdb, err := dbb.Redis(ctx)
	if err != nil {
		return err
	}
	var cacheModel cache.ThirdCache
	if rdb == nil {
		mdb, err := dbb.Mongo(ctx)
		if err != nil {
			return err
		}
		mc, err := mgo.NewCacheMgo(mdb.GetDB())
		if err != nil {
			return err
		}
		cacheModel = mcache.NewThirdCache(mc)
	} else {
		cacheModel = redis.NewThirdCache(rdb)
	}
	offlinePusher, err := offlinepush.NewOfflinePusher(&config.RpcConfig, cacheModel, string(config.FcmConfigPath))
	if err != nil {
		return err
	}
	builder := mqbuild.NewBuilder(&config.KafkaConfig)

	offlinePushProducer, err := builder.GetTopicProducer(ctx, config.KafkaConfig.ToOfflinePushTopic)
	if err != nil {
		return err
	}
	database := controller.NewPushDatabase(cacheModel, offlinePushProducer)

	pushConsumer, err := builder.GetTopicConsumer(ctx, config.KafkaConfig.ToPushTopic)
	if err != nil {
		return err
	}
	offlinePushConsumer, err := builder.GetTopicConsumer(ctx, config.KafkaConfig.ToOfflinePushTopic)
	if err != nil {
		return err
	}

	pushHandler, err := NewConsumerHandler(ctx, config, database, offlinePusher, rdb, client)
	if err != nil {
		return err
	}

	offlineHandler := NewOfflinePushConsumerHandler(offlinePusher)

	pbpush.RegisterPushMsgServiceServer(server, &pushServer{
		database:      database,
		disCov:        client,
		offlinePusher: offlinePusher,
	})

	go func() {
		pushHandler.WaitCache()
		fn := func(msg mq.Message) error {
			pushHandler.HandleMs2PsChat(authverify.WithTempAdmin(msg.Context()), msg.Value())
			return nil
		}
		consumerCtx := mcontext.SetOperationID(context.Background(), "push_"+strconv.Itoa(int(rand.Uint32())))
		log.ZInfo(consumerCtx, "begin consume messages")
		for {
			if err := pushConsumer.Subscribe(consumerCtx, fn); err != nil {
				log.ZError(consumerCtx, "subscribe err", err)
				return
			}
		}
	}()

	go func() {
		fn := func(msg mq.Message) error {
			offlineHandler.HandleMsg2OfflinePush(msg.Context(), msg.Value())
			return nil
		}
		consumerCtx := mcontext.SetOperationID(context.Background(), "push_"+strconv.Itoa(int(rand.Uint32())))
		log.ZInfo(consumerCtx, "begin consume messages")
		for {
			if err := offlinePushConsumer.Subscribe(consumerCtx, fn); err != nil {
				log.ZError(consumerCtx, "subscribe err", err)
				return
			}
		}
	}()

	return nil
}
