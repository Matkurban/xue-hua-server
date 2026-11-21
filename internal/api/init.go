package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"gitee.com/matkurban/openimsdk-tools/discovery"
	"gitee.com/matkurban/openimsdk-tools/log"
	"gitee.com/matkurban/openimsdk-tools/utils/datautil"
	"gitee.com/matkurban/openimsdk-tools/utils/network"
	"gitee.com/matkurban/openimsdk-tools/utils/runtimeenv"
	"google.golang.org/grpc"
	conf "server/v3/pkg/common/config"
)

type Config struct {
	conf.AllConfig

	ConfigPath conf.Path
	Index      conf.Index
}

func Start(ctx context.Context, config *Config, client discovery.SvcDiscoveryRegistry, service grpc.ServiceRegistrar) error {
	apiPort, err := datautil.GetElemByIndex(config.API.Api.Ports, int(config.Index))
	if err != nil {
		return err
	}

	router, err := newGinRouter(ctx, client, config)
	if err != nil {
		return err
	}

	apiCtx, apiCancel := context.WithCancelCause(context.Background())
	done := make(chan struct{})
	go func() {
		httpServer := &http.Server{
			Handler:           router,
			Addr:              net.JoinHostPort(network.GetListenIP(config.API.Api.ListenIP), strconv.Itoa(apiPort)),
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			defer close(done)
			select {
			case <-ctx.Done():
				apiCancel(fmt.Errorf("recv ctx %w", context.Cause(ctx)))
			case <-apiCtx.Done():
			}
			log.ZDebug(ctx, "api server is shutting down")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				log.ZWarn(ctx, "api server shutdown err", err)
			}
		}()
		log.CInfo(ctx, "api server is init", "runtimeEnv", runtimeenv.RuntimeEnvironment(), "address", httpServer.Addr, "apiPort", apiPort)
		err := httpServer.ListenAndServe()
		if err == nil {
			err = errors.New("api done")
		}
		apiCancel(err)
	}()

	//if config.Discovery.Enable == conf.ETCD {
	//	cm := disetcd.NewConfigManager(client.(*etcd.SvcDiscoveryRegistryImpl).GetClient(), config.GetConfigNames())
	//	cm.Watch(ctx)
	//}
	//sigs := make(chan os.Signal, 1)
	//signal.Notify(sigs, syscall.SIGTERM)
	//select {
	//case val := <-sigs:
	//	log.ZDebug(ctx, "recv exit", "signal", val.String())
	//	cancel(fmt.Errorf("signal %s", val.String()))
	//case <-ctx.Done():
	//}
	<-apiCtx.Done()
	exitCause := context.Cause(apiCtx)
	log.ZWarn(ctx, "api server exit", exitCause)
	timer := time.NewTimer(time.Second * 15)
	defer timer.Stop()
	select {
	case <-timer.C:
		log.ZWarn(ctx, "api server graceful stop timeout", nil)
	case <-done:
		log.ZDebug(ctx, "api server graceful stop done")
	}
	return exitCause
}
