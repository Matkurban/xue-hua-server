package cmd

import (
	"context"

	"server/v3/pkg/common/program"
	"github.com/spf13/cobra"
	"server/v3/internal/msgtransfer"
	"server/v3/pkg/common/config"
	"server/v3/pkg/common/prommetrics"
	"server/v3/pkg/common/startrpc"
	"server/v3/version"
)

type MsgTransferCmd struct {
	*RootCmd
	ctx               context.Context
	configMap         map[string]any
	msgTransferConfig *msgtransfer.Config
}

func NewMsgTransferCmd() *MsgTransferCmd {
	var msgTransferConfig msgtransfer.Config
	ret := &MsgTransferCmd{msgTransferConfig: &msgTransferConfig}
	ret.configMap = map[string]any{
		config.ServerMsgTransferCfgFileName: &msgTransferConfig.MsgTransfer,
		config.RedisConfigFileName:          &msgTransferConfig.RedisConfig,
		config.MongodbConfigFileName:        &msgTransferConfig.MongodbConfig,
		config.KafkaConfigFileName:          &msgTransferConfig.KafkaConfig,
		config.ShareFileName:                &msgTransferConfig.Share,
		config.WebhooksConfigFileName:       &msgTransferConfig.WebhooksConfig,
		config.DiscoveryConfigFilename:      &msgTransferConfig.Discovery,
	}
	ret.RootCmd = NewRootCmd(program.GetProcessName(), WithConfigMap(ret.configMap))
	ret.ctx = context.WithValue(context.Background(), "version", version.Version)
	ret.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return ret.runE()
	}
	return ret
}

func (m *MsgTransferCmd) Exec() error {
	return m.Execute()
}

func (m *MsgTransferCmd) runE() error {
	m.msgTransferConfig.Index = config.Index(m.Index())
	var prometheus config.Prometheus
	return startrpc.Start(
		m.ctx, &m.msgTransferConfig.Discovery,
		&m.msgTransferConfig.MsgTransfer.CircuitBreaker,
		&m.msgTransferConfig.MsgTransfer.RateLimiter,
		&prometheus,
		"", "",
		true,
		nil, int(m.msgTransferConfig.Index),
		prommetrics.MessageTransferKeyName,
		nil,
		m.msgTransferConfig,
		[]string{},
		[]string{},
		msgtransfer.Start,
	)
}
