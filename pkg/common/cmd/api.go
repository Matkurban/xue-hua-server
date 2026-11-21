package cmd

import (
	"context"

	"server/v3/pkg/common/program"
	"github.com/spf13/cobra"
	"server/v3/internal/api"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/prommetrics"
	"server/v3/pkg/common/startrpc"
	"server/v3/version"
)

type ApiCmd struct {
	*RootCmd
	ctx       context.Context
	configMap map[string]any
	apiConfig *api.Config
}

func NewApiCmd() *ApiCmd {
	var apiConfig api.Config
	ret := &ApiCmd{apiConfig: &apiConfig}
	ret.configMap = map[string]any{
		config.DiscoveryConfigFilename:          &apiConfig.Discovery,
		config.KafkaConfigFileName:              &apiConfig.Kafka,
		config.LocalCacheConfigFileName:         &apiConfig.LocalCache,
		config.LogConfigFileName:                &apiConfig.Log,
		config.RustFSConfigFileName:             &apiConfig.RustFS,
		config.MongodbConfigFileName:            &apiConfig.Mongo,
		config.NotificationFileName:             &apiConfig.Notification,
		config.ServerAPICfgFileName:             &apiConfig.API,
		config.ServerCronTaskCfgFileName:        &apiConfig.CronTask,
		config.ServerMsgGatewayCfgFileName:      &apiConfig.MsgGateway,
		config.ServerMsgTransferCfgFileName:     &apiConfig.MsgTransfer,
		config.ServerPushCfgFileName:            &apiConfig.Push,
		config.ServerRPCAuthCfgFileName:         &apiConfig.Auth,
		config.ServerRPCConversationCfgFileName: &apiConfig.Conversation,
		config.ServerRPCFriendCfgFileName:       &apiConfig.Friend,
		config.ServerRPCGroupCfgFileName:        &apiConfig.Group,
		config.ServerRPCMsgCfgFileName:          &apiConfig.Msg,
		config.ServerRPCThirdCfgFileName:        &apiConfig.Third,
		config.ServerRPCUserCfgFileName:         &apiConfig.User,
		config.RedisConfigFileName:              &apiConfig.Redis,
		config.ShareFileName:                    &apiConfig.Share,
		config.WebhooksConfigFileName:           &apiConfig.Webhooks,
	}
	ret.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(ret.configMap))
	ret.ctx = context.WithValue(context.Background(), "version", version.Version)
	ret.Command.RunE = func(cmd *cobra.Command, args []string) error {
		apiConfig.ConfigPath = config.Path(ret.configPath)
		return ret.runE()
	}
	return ret
}

func (a *ApiCmd) Exec() error {
	return a.Execute()
}

func (a *ApiCmd) runE() error {
	a.apiConfig.Index = config.Index(a.Index())
	prometheus := config.Prometheus{
		Enable: a.apiConfig.API.Prometheus.Enable,
		Ports:  a.apiConfig.API.Prometheus.Ports,
	}
	return startrpc.Start(
		a.ctx, &a.apiConfig.Discovery,
		nil,
		nil,
		// &a.apiConfig.API.RateLimiter,
		&prometheus,
		a.apiConfig.API.Api.ListenIP, "",
		a.apiConfig.API.Prometheus.AutoSetPorts,
		nil, int(a.apiConfig.Index),
		prommetrics.APIKeyName,
		&a.apiConfig.Notification,
		a.apiConfig,
		[]string{},
		[]string{},
		api.Start,
	)
}
