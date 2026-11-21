package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"server/v3/pkg/common/storage/mongoutil"
	"gitee.com/matkurban/openimsdk-tools/db/redisutil"
	"gitee.com/matkurban/openimsdk-tools/discovery/etcd"
	"gitee.com/matkurban/openimsdk-tools/mq/kafka"
	"gitee.com/matkurban/openimsdk-tools/s3/minio"
	"gitee.com/matkurban/openimsdk-tools/system/program"
	"server/v3/pkg/common/config"
)

const maxRetry = 180

func CheckEtcd(ctx context.Context, config *config.Etcd) error {
	return etcd.Check(ctx, config.Address, "/check_server_component",
		true,
		etcd.WithDialTimeout(10*time.Second),
		etcd.WithMaxCallSendMsgSize(20*1024*1024),
		etcd.WithUsernameAndPassword(config.Username, config.Password))
}

func CheckMongo(ctx context.Context, config *config.Mongo) error {
	return mongoutil.Check(ctx, config.Build())
}

func CheckRedis(ctx context.Context, config *config.Redis) error {
	return redisutil.Check(ctx, config.Build())
}

func CheckMinIO(ctx context.Context, config *config.Minio) error {
	return minio.Check(ctx, config.Build())
}

func CheckKafka(ctx context.Context, conf *config.Kafka) error {
	return kafka.CheckHealth(ctx, conf.Build())
}

func initConfig(configDir string) (*config.Mongo, *config.Redis, *config.Kafka, *config.Minio, *config.Discovery, error) {
	var (
		mongoConfig = &config.Mongo{}
		redisConfig = &config.Redis{}
		kafkaConfig = &config.Kafka{}
		minioConfig = &config.Minio{}
		discovery   = &config.Discovery{}
		thirdConfig = &config.Third{}
	)

	err := config.Load(configDir, config.MongodbConfigFileName, config.EnvPrefixMap[config.MongodbConfigFileName], mongoConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = config.Load(configDir, config.RedisConfigFileName, config.EnvPrefixMap[config.RedisConfigFileName], redisConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = config.Load(configDir, config.KafkaConfigFileName, config.EnvPrefixMap[config.KafkaConfigFileName], kafkaConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	err = config.Load(configDir, config.ServerRPCThirdCfgFileName, config.EnvPrefixMap[config.ServerRPCThirdCfgFileName], thirdConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if thirdConfig.Object.Enable == "minio" {
		err = config.Load(configDir, config.MinioConfigFileName, config.EnvPrefixMap[config.MinioConfigFileName], minioConfig)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	} else {
		minioConfig = nil
	}
	err = config.Load(configDir, config.DiscoveryConfigFilename, config.EnvPrefixMap[config.DiscoveryConfigFilename], discovery)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return mongoConfig, redisConfig, kafkaConfig, minioConfig, discovery, nil
}

func main() {
	var index int
	var configDir string
	flag.IntVar(&index, "i", 0, "Index number")
	defaultConfigDir := filepath.Join("..", "..", "..", "..", "..", "config")
	flag.StringVar(&configDir, "c", defaultConfigDir, "Configuration dir")
	flag.Parse()

	fmt.Printf("%s Index: %d, Config Path: %s\n", filepath.Base(os.Args[0]), index, configDir)

	mongoConfig, redisConfig, kafkaConfig, minioConfig, zookeeperConfig, err := initConfig(configDir)
	if err != nil {
		program.ExitWithError(err)
	}

	ctx := context.Background()
	err = performChecks(ctx, mongoConfig, redisConfig, kafkaConfig, minioConfig, zookeeperConfig, maxRetry)
	if err != nil {
		// Assume program.ExitWithError logs the error and exits.
		// Replace with your error handling logic as necessary.
		program.ExitWithError(err)
	}
}

func performChecks(ctx context.Context, mongoConfig *config.Mongo, redisConfig *config.Redis, kafkaConfig *config.Kafka, minioConfig *config.Minio, discovery *config.Discovery, maxRetry int) error {
	checksDone := make(map[string]bool)

	checks := map[string]func(ctx context.Context) error{
		"Mongo": func(ctx context.Context) error {
			return CheckMongo(ctx, mongoConfig)
		},
		"Redis": func(ctx context.Context) error {
			return CheckRedis(ctx, redisConfig)
		},
		"Kafka": func(ctx context.Context) error {
			return CheckKafka(ctx, kafkaConfig)
		},
	}
	if minioConfig != nil {
		checks["MinIO"] = func(ctx context.Context) error {
			return CheckMinIO(ctx, minioConfig)
		}
	}
	if discovery.Enable == "etcd" {
		checks["Etcd"] = func(ctx context.Context) error {
			return CheckEtcd(ctx, &discovery.Etcd)
		}
	}

	for range maxRetry {
		allSuccess := true
		for name, check := range checks {
			if !checksDone[name] {
				if err := check(ctx); err != nil {
					fmt.Printf("%s check failed: %v\n", name, err)
					allSuccess = false
				} else {
					fmt.Printf("%s check succeeded.\n", name)
					checksDone[name] = true
				}
			}
		}

		if allSuccess {
			fmt.Println("All components checks passed successfully.")
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("not all components checks passed successfully after %d attempts", maxRetry)
}
