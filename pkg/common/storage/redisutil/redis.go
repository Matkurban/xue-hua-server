package redisutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/mw/specialerror"
)

func init() {
	if err := specialerror.AddReplace(redis.Nil, errs.ErrRecordNotFound); err != nil {
		panic(err)
	}
}

// RedisMode constants.
const (
	ClusterMode    = "cluster"
	SentinelMode   = "sentinel"
	StandaloneMode = "standalone"
)

var ErrFailedToLoadCACert = errors.New("failed to load CACertificate")

// TLSConfig configures optional TLS client authentication for Redis.
type TLSConfig struct {
	Insecure       bool   `json:"insecure" yaml:"insecure" config:"allowempty"`
	ServerName     string `json:"serverName" yaml:"server_name" config:"allowempty"`
	CACertFile     string `json:"caCertFile" yaml:"ca_cert_file" config:"allowempty"`
	ClientCertFile string `json:"clientCertFile" yaml:"client_cert_file" config:"allowempty"`
	ClientKeyFile  string `json:"clientKeyFile" yaml:"client_key_file" config:"allowempty"`
}

func (c *TLSConfig) ClientTLSConfig() (*tls.Config, error) {
	tlsConf := tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if c.ClientCertFile != "" {
		certificate, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %w", err)
		}
		tlsConf.Certificates = []tls.Certificate{certificate}
	}
	if c.Insecure {
		tlsConf.InsecureSkipVerify = true
	}
	if c.ServerName != "" {
		tlsConf.ServerName = c.ServerName
	}
	if c.CACertFile != "" {
		certPool := x509.NewCertPool()
		ca, err := os.ReadFile(c.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %w", err)
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, ErrFailedToLoadCACert
		}
		tlsConf.RootCAs = certPool
	}
	return &tlsConf, nil
}

// Config defines the configuration parameters for a Redis client, including
// options for standalone, sentinel, and cluster deployments.
type Config struct {
	RedisMode string     // "cluster", "sentinel", or "standalone".
	Address   []string   // List of Redis server addresses (host:port).
	Username  string     // Username for Redis authentication (Redis 6 ACL).
	Password  string     // Password for Redis authentication.
	MaxRetry  int        // Maximum number of retries for a command.
	DB        int        // Database number to connect to, for non-cluster mode.
	PoolSize  int        // Number of connections to pool.
	TLS       *TLSConfig // Optional TLS configuration for secure connections.
	Sentinel  *Sentinel  // Sentinel configuration for high availability.
}

type Sentinel struct {
	MasterName     string   `json:"masterName" yaml:"masterName" config:"allowempty"`
	SentinelAddrs  []string `json:"sentinelAddrs" yaml:"sentinelAddrs" config:"allowempty"`
	RouteByLatency bool     `json:"routeByLatency" yaml:"routeByLatency" config:"allowempty"`
	RouteRandomly  bool     `json:"routeRandomly" yaml:"routeRandomly" config:"allowempty"`
}

func NewRedisClient(ctx context.Context, config *Config) (redis.UniversalClient, error) {
	if len(config.Address) == 0 {
		return nil, errs.New("redis address is empty").Wrap()
	}

	if config.RedisMode == SentinelMode && config.Sentinel != nil {
		if config.Sentinel.MasterName == "" {
			return nil, errs.New("sentinel master name is required").Wrap()
		}
		if len(config.Sentinel.SentinelAddrs) == 0 {
			return nil, errs.New("sentinel addresses are required").Wrap()
		}
	}

	var tlsConf *tls.Config
	if config.TLS != nil {
		var err error
		tlsConf, err = config.TLS.ClientTLSConfig()
		if err != nil {
			return nil, errs.WrapMsg(err, "failed to get TLS config")
		}
	}

	var cli redis.UniversalClient
	switch {
	case config.Sentinel != nil && config.RedisMode == SentinelMode:
		opt := &redis.FailoverOptions{
			MasterName:     config.Sentinel.MasterName,
			SentinelAddrs:  config.Sentinel.SentinelAddrs,
			Username:       config.Username,
			Password:       config.Password,
			DB:             config.DB,
			PoolSize:       config.PoolSize,
			MaxRetries:     config.MaxRetry,
			RouteByLatency: config.Sentinel.RouteByLatency,
			RouteRandomly:  config.Sentinel.RouteRandomly,
			TLSConfig:      tlsConf,
		}
		if opt.RouteByLatency || opt.RouteRandomly {
			cli = redis.NewFailoverClusterClient(opt)
		} else {
			cli = redis.NewFailoverClient(opt)
		}
	case config.RedisMode == ClusterMode:
		cli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:      config.Address,
			Username:   config.Username,
			Password:   config.Password,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
			TLSConfig:  tlsConf,
		})
	default:
		cli = redis.NewClient(&redis.Options{
			Addr:       config.Address[0],
			Username:   config.Username,
			Password:   config.Password,
			DB:         config.DB,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
			TLSConfig:  tlsConf,
		})
	}

	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, errs.WrapMsg(err, "Redis Ping failed", "Address", config.Address, "Username", config.Username, "RedisMode", config.RedisMode)
	}
	return cli, nil
}
