package cmd

import (
	"testing"

	"server/v3/pkg/common/config"

	"github.com/stretchr/testify/require"
)

func TestKubernetesDoesNotInitializeEtcd(t *testing.T) {
	root := NewRootCmd("server-api")
	require.NoError(t, root.initEtcdForEnvironment(config.KUBERNETES))
	require.Nil(t, root.etcdClient)
}
