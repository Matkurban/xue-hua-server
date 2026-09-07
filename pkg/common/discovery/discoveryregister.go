package discovery

import (
	"time"

	"gitee.com/matkurban/openimsdk-tools/discovery"
	"gitee.com/matkurban/openimsdk-tools/discovery/standalone"
	"gitee.com/matkurban/openimsdk-tools/utils/runtimeenv"
	"google.golang.org/grpc"
	"server/v3/pkg/common/config"

	"gitee.com/matkurban/openimsdk-tools/discovery/kubernetes"

	"gitee.com/matkurban/openimsdk-tools/discovery/etcd"
	"gitee.com/matkurban/openimsdk-tools/errs"
)

// NewDiscoveryRegister creates a new service discovery and registry client based on the provided environment type.
func NewDiscoveryRegister(discovery *config.Discovery, watchNames []string) (discovery.SvcDiscoveryRegistry, error) {
	if config.Standalone() {
		return standalone.GetSvcDiscoveryRegistry(), nil
	}
	if runtimeenv.RuntimeEnvironment() == config.KUBERNETES {
		return kubernetes.NewConnManager(discovery.Kubernetes.Namespace, nil,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(1024*1024*20),
			),
		)
	}

	switch discovery.Enable {
	case config.ETCD:
		return etcd.NewSvcDiscoveryRegistry(
			discovery.Etcd.RootDirectory,
			discovery.Etcd.Address,
			watchNames,
			etcd.WithDialTimeout(10*time.Second),
			etcd.WithMaxCallSendMsgSize(20*1024*1024),
			etcd.WithUsernameAndPassword(discovery.Etcd.Username, discovery.Etcd.Password))
	default:
		return nil, errs.New("unsupported discovery type", "type", discovery.Enable).Wrap()
	}
}
