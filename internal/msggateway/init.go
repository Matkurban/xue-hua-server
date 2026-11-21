package msggateway

import (
	"context"
	"time"

	"server/v3/pkg/discovery"
	"server/v3/pkg/util/datautil"
	"server/v3/pkg/util/runtimeenv"
	"google.golang.org/grpc"
	"server/v3/pkg/common/config"
	"server/v3/pkg/dbbuild"
	"server/v3/pkg/rpccache"

	"server/v3/pkg/common/log"
)

type Config struct {
	MsgGateway     config.MsgGateway
	Share          config.Share
	RedisConfig    config.Redis
	WebhooksConfig config.Webhooks
	Discovery      config.Discovery
	Index          config.Index
}

// Start run ws server.
func Start(ctx context.Context, conf *Config, client discovery.SvcDiscoveryRegistry, server grpc.ServiceRegistrar) error {
	log.CInfo(ctx, "MSG-GATEWAY server is initializing", "runtimeEnv", runtimeenv.RuntimeEnvironment(),
		"rpcPorts", conf.MsgGateway.RPC.Ports,
		"wsPort", conf.MsgGateway.LongConnSvr.Ports, "prometheusPorts", conf.MsgGateway.Prometheus.Ports)
	wsPort, err := datautil.GetElemByIndex(conf.MsgGateway.LongConnSvr.Ports, int(conf.Index))
	if err != nil {
		return err
	}

	dbb := dbbuild.NewBuilder(nil, &conf.RedisConfig)
	rdb, err := dbb.Redis(ctx)
	if err != nil {
		return err
	}

	longServer := NewWsServer(
		conf,
		WithPort(wsPort),
		WithMaxConnNum(int64(conf.MsgGateway.LongConnSvr.WebsocketMaxConnNum)),
		WithHandshakeTimeout(time.Duration(conf.MsgGateway.LongConnSvr.WebsocketTimeout)*time.Second),
		WithMessageMaxMsgLength(conf.MsgGateway.LongConnSvr.WebsocketMaxMsgLen),
	)

	hubServer := NewServer(longServer, conf, func(srv *Server) error {
		var err error
		longServer.online, err = rpccache.NewOnlineCache(srv.userClient, nil, rdb, false, longServer.subscriberUserOnlineStatusChanges)
		return err
	})

	if err := hubServer.InitServer(ctx, conf, client, server); err != nil {
		return err
	}

	go longServer.ChangeOnlineStatus(4)

	return hubServer.LongConnServer.Run(ctx)
}

//
//// Start run ws server.
//func Start(ctx context.Context, index int, conf *Config) error {
//	log.CInfo(ctx, "MSG-GATEWAY server is initializing", "runtimeEnv", runtimeenv.RuntimeEnvironment(),
//		"rpcPorts", conf.MsgGateway.RPC.Ports,
//		"wsPort", conf.MsgGateway.LongConnSvr.Ports, "prometheusPorts", conf.MsgGateway.Prometheus.Ports)
//	wsPort, err := datautil.GetElemByIndex(conf.MsgGateway.LongConnSvr.Ports, index)
//	if err != nil {
//		return err
//	}
//
//	rdb, err := redisutil.NewRedisClient(ctx, conf.RedisConfig.Build())
//	if err != nil {
//		return err
//	}
//	longServer := NewWsServer(
//		conf,
//		WithPort(wsPort),
//		WithMaxConnNum(int64(conf.MsgGateway.LongConnSvr.WebsocketMaxConnNum)),
//		WithHandshakeTimeout(time.Duration(conf.MsgGateway.LongConnSvr.WebsocketTimeout)*time.Second),
//		WithMessageMaxMsgLength(conf.MsgGateway.LongConnSvr.WebsocketMaxMsgLen),
//	)
//
//	hubServer := NewServer(longServer, conf, func(srv *Server) error {
//		var err error
//		longServer.online, err = rpccache.NewOnlineCache(srv.userClient, nil, rdb, false, longServer.subscriberUserOnlineStatusChanges)
//		return err
//	})
//
//	go longServer.ChangeOnlineStatus(4)
//
//	netDone := make(chan error)
//	go func() {
//		err = hubServer.Start(ctx, index, conf)
//		netDone <- err
//	}()
//	return hubServer.LongConnServer.Run(netDone)
//}
