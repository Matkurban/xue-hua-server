package cmd

import (
	"context"

	"server/v3/pkg/common/program"
	"github.com/spf13/cobra"
	"server/v3/internal/rpc/third"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/startrpc"
	"server/v3/version"
)

type ThirdRpcCmd struct {
	*RootCmd
	ctx         context.Context
	configMap   map[string]any
	thirdConfig *third.Config
}

func NewThirdRpcCmd() *ThirdRpcCmd {
	var thirdConfig third.Config
	ret := &ThirdRpcCmd{thirdConfig: &thirdConfig}
	ret.configMap = map[string]any{
		config.ServerRPCThirdCfgFileName: &thirdConfig.RpcConfig,
		config.RedisConfigFileName:       &thirdConfig.RedisConfig,
		config.MongodbConfigFileName:     &thirdConfig.MongodbConfig,
		config.ShareFileName:             &thirdConfig.Share,
		config.NotificationFileName:      &thirdConfig.NotificationConfig,
		config.RustFSConfigFileName:      &thirdConfig.RustFSConfig,
		config.LocalCacheConfigFileName:  &thirdConfig.LocalCacheConfig,
		config.DiscoveryConfigFilename:   &thirdConfig.Discovery,
	}
	ret.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(ret.configMap))
	ret.ctx = context.WithValue(context.Background(), "version", version.Version)
	ret.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return ret.runE()
	}
	return ret
}

func (a *ThirdRpcCmd) Exec() error {
	return a.Execute()
}

func (a *ThirdRpcCmd) runE() error {
	return startrpc.Start(a.ctx, &a.thirdConfig.Discovery, &a.thirdConfig.RpcConfig.CircuitBreaker, &a.thirdConfig.RpcConfig.RateLimiter, &a.thirdConfig.RpcConfig.Prometheus, a.thirdConfig.RpcConfig.RPC.ListenIP,
		a.thirdConfig.RpcConfig.RPC.RegisterIP, a.thirdConfig.RpcConfig.RPC.AutoSetPorts, a.thirdConfig.RpcConfig.RPC.Ports,
		a.Index(), a.thirdConfig.Discovery.RpcService.Third, &a.thirdConfig.NotificationConfig, a.thirdConfig,
		[]string{
			a.thirdConfig.RpcConfig.GetConfigFileName(),
			a.thirdConfig.RedisConfig.GetConfigFileName(),
			a.thirdConfig.MongodbConfig.GetConfigFileName(),
			a.thirdConfig.NotificationConfig.GetConfigFileName(),
			a.thirdConfig.Share.GetConfigFileName(),
			a.thirdConfig.RustFSConfig.GetConfigFileName(),
			a.thirdConfig.LocalCacheConfig.GetConfigFileName(),
			a.thirdConfig.Discovery.GetConfigFileName(),
		}, nil,
		third.Start)
}
