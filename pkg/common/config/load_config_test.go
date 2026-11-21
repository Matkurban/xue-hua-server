package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadServerConfigs(t *testing.T) {
	t.Setenv("DEPLOYMENT_TYPE", "")
	tests := []struct {
		name   string
		config any
	}{
		{ServerAPICfgFileName, &API{}},
		{ServerCronTaskCfgFileName, &CronTask{}},
		{ServerMsgGatewayCfgFileName, &MsgGateway{}},
		{ServerMsgTransferCfgFileName, &MsgTransfer{}},
		{ServerPushCfgFileName, &Push{}},
		{ServerRPCAuthCfgFileName, &Auth{}},
		{ServerRPCConversationCfgFileName, &Conversation{}},
		{ServerRPCFriendCfgFileName, &Friend{}},
		{ServerRPCGroupCfgFileName, &Group{}},
		{ServerRPCMsgCfgFileName, &Msg{}},
		{ServerRPCThirdCfgFileName, &Third{}},
		{ServerRPCUserCfgFileName, &User{}},
		{LogConfigFileName, &Log{}},
		{MongodbConfigFileName, &Mongo{}},
		{RustFSConfigFileName, &RustFS{}},
		{NotificationFileName, &Notification{}},
		{WebhooksConfigFileName, &Webhooks{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := "IMENV_" + strings.ToUpper(strings.ReplaceAll(strings.TrimSuffix(tt.name, ".yml"), "-", "_"))
			require.Equal(t, prefix, EnvPrefixMap[tt.name])
			require.NoError(t, Load("../../../config", tt.name, prefix, tt.config))
		})
	}
}

func TestLoadServerConfigEnvironment(t *testing.T) {
	t.Setenv("DEPLOYMENT_TYPE", "")
	t.Setenv("IMENV_SERVER_RPC_USER_RPC_LISTENIP", "127.0.0.2")
	t.Setenv("IMENV_SERVER_RPC_USER_RPC_PORTS", "20110,20111")
	var user User
	require.NoError(t, Load("../../../config", ServerRPCUserCfgFileName, EnvPrefixMap[ServerRPCUserCfgFileName], &user))
	require.Equal(t, "127.0.0.2", user.RPC.ListenIP)
	require.Equal(t, []int{20110, 20111}, user.RPC.Ports)
}

func TestLoadStorageEnvironment(t *testing.T) {
	t.Setenv("DEPLOYMENT_TYPE", "")
	t.Setenv("IMENV_MONGODB_DATABASE", "server_config_test")
	t.Setenv("IMENV_MONGODB_PASSWORD", "test-only-password")
	var database Mongo
	require.NoError(t, Load("../../../config", MongodbConfigFileName, EnvPrefixMap[MongodbConfigFileName], &database))
	require.Equal(t, "server_config_test", database.Database)
	require.Equal(t, "test-only-password", database.Password)
}
