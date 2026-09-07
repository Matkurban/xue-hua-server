// Command render generates Kubernetes application manifests from the current config.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type object = map[string]any

type service struct {
	name      string
	discovery string
	rpc       int
	public    int
	metrics   int
}

var services = []service{
	{name: "server-api", public: 10002, metrics: 12002},
	{name: "server-crontask"},
	{name: "server-msggateway", discovery: "messageGateway", rpc: 10140, public: 10001, metrics: 12140},
	{name: "server-msgtransfer", metrics: 12020},
	{name: "server-push", discovery: "push", rpc: 10170, metrics: 12170},
	{name: "server-rpc-auth", discovery: "auth", rpc: 10200, metrics: 12200},
	{name: "server-rpc-conversation", discovery: "conversation", rpc: 10220, metrics: 12220},
	{name: "server-rpc-friend", discovery: "friend", rpc: 10240, metrics: 12240},
	{name: "server-rpc-group", discovery: "group", rpc: 10260, metrics: 12260},
	{name: "server-rpc-msg", discovery: "msg", rpc: 10280, metrics: 12280},
	{name: "server-rpc-third", discovery: "third", rpc: 10300, metrics: 12300},
	{name: "server-rpc-user", discovery: "user", rpc: 10320, metrics: 12320},
}

func set(config object, path string, value any) {
	keys := strings.Split(path, ".")
	for _, key := range keys[:len(keys)-1] {
		child, ok := config[key].(map[string]any)
		if !ok {
			child = object{}
			config[key] = child
		}
		config = child
	}
	config[keys[len(keys)-1]] = value
}

func encode(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), encoder.Close()
}

func configMap(root, namespace string) (object, error) {
	names := []string{
		"discovery.yml", "kafka.yml", "local-cache.yml", "log.yml", "minio.yml",
		"mongodb.yml", "notification.yml", "redis.yml", "share.yml", "webhooks.yml",
	}
	for _, service := range services {
		names = append(names, service.name+".yml")
	}
	data := map[string]string{}
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(root, "config", name))
		if err != nil {
			return nil, err
		}
		var config object
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		switch name {
		case "discovery.yml":
			set(config, "enable", "kubernetes")
			set(config, "kubernetes.namespace", namespace)
			for _, service := range services {
				if service.discovery != "" {
					set(config, "rpcService."+service.discovery, service.name)
				}
			}
		case "mongodb.yml":
			set(config, "uri", "")
			set(config, "address", []string{"mongo-service:37017"})
			set(config, "mongoMode", "standalone")
			set(config, "username", "")
			set(config, "password", "")
		case "redis.yml":
			set(config, "address", []string{"redis-service:16379"})
			set(config, "redisMode", "standalone")
			set(config, "username", "")
			set(config, "password", "")
		case "kafka.yml":
			set(config, "address", []string{"kafka-service:9092"})
			set(config, "username", "")
			set(config, "password", "")
		case "minio.yml":
			set(config, "internalAddress", "minio-service:10005")
			set(config, "accessKeyID", "")
			set(config, "secretAccessKey", "")
		}
		for _, service := range services {
			if name != service.name+".yml" {
				continue
			}
			if service.rpc != 0 {
				set(config, "rpc.autoSetPorts", false)
				set(config, "rpc.ports", []int{service.rpc})
			}
			if service.metrics != 0 {
				set(config, "prometheus.autoSetPorts", false)
				set(config, "prometheus.ports", []int{service.metrics})
			}
		}
		encoded, err := encode(config)
		if err != nil {
			return nil, err
		}
		data[name] = string(encoded)
	}
	return object{
		"apiVersion": "v1", "kind": "ConfigMap",
		"metadata": object{"name": "server-config", "namespace": namespace},
		"data":     data,
	}, nil
}

func secretEnv(name, secret, key string) object {
	return object{
		"name": name,
		"valueFrom": object{
			"secretKeyRef": object{"name": secret, "key": key},
		},
	}
}

func manifests(service service, namespace, image string) (object, object) {
	labels := object{"app": service.name}
	containerPorts := []object{}
	servicePorts := []object{}
	addPort := func(name string, port int) {
		if port == 0 {
			return
		}
		containerPorts = append(containerPorts, object{"name": name, "containerPort": port})
		servicePorts = append(servicePorts, object{"name": name, "port": port, "targetPort": name})
	}
	// Discovery reads the first service port and requires the exact name "grpc".
	addPort("grpc", service.rpc)
	addPort("http", service.public)
	addPort("metrics", service.metrics)
	container := object{
		"name": service.name, "image": image, "imagePullPolicy": "IfNotPresent",
		"command": []string{"/server/bin/" + service.name},
		"args":    []string{"-c", "/config"},
		"env": []object{
			{"name": "CONFIG_PATH", "value": "/config"},
			secretEnv("IMENV_REDIS_PASSWORD", "server-redis-secret", "redis-password"),
			secretEnv("IMENV_MONGODB_USERNAME", "server-mongo-secret", "mongo_server_username"),
			secretEnv("IMENV_MONGODB_PASSWORD", "server-mongo-secret", "mongo_server_password"),
			secretEnv("IMENV_MINIO_ACCESSKEYID", "server-minio-secret", "minio-root-user"),
			secretEnv("IMENV_MINIO_SECRETACCESSKEY", "server-minio-secret", "minio-root-password"),
			secretEnv("IMENV_KAFKA_PASSWORD", "server-kafka-secret", "kafka-password"),
		},
		"volumeMounts": []object{{"name": "server-config", "mountPath": "/config", "readOnly": true}},
	}
	if len(containerPorts) > 0 {
		container["ports"] = containerPorts
	}
	deployment := object{
		"apiVersion": "apps/v1", "kind": "Deployment",
		"metadata": object{"name": service.name, "namespace": namespace},
		"spec": object{
			"replicas": 1, "selector": object{"matchLabels": labels},
			"template": object{
				"metadata": object{"labels": labels},
				"spec": object{
					"serviceAccountName": "server", "automountServiceAccountToken": true,
					"containers": []object{container},
					"volumes": []object{{
						"name": "server-config", "configMap": object{"name": "server-config"},
					}},
				},
			},
		},
	}
	var svc object
	if len(servicePorts) > 0 {
		svc = object{
			"apiVersion": "v1", "kind": "Service",
			"metadata": object{"name": service.name, "namespace": namespace},
			"spec":     object{"selector": labels, "ports": servicePorts, "type": "ClusterIP"},
		}
	}
	return deployment, svc
}

func run(root, namespace, image string, check bool) error {
	config, err := configMap(root, namespace)
	if err != nil {
		return err
	}
	files := map[string]object{"server-config.yml": config}
	for _, service := range services {
		deployment, svc := manifests(service, namespace, image)
		files[service.name+"-deployment.yml"] = deployment
		if svc != nil {
			files[service.name+"-service.yml"] = svc
		}
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		content, err := encode(files[name])
		if err != nil {
			return err
		}
		content = append([]byte("# Generated by go run ./deployments/render; edit config/ or the renderer.\n"), content...)
		path := filepath.Join(root, "deployments", "deploy", name)
		if check {
			current, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(current, content) {
				return fmt.Errorf("%s is stale; run go run ./deployments/render", path)
			}
		} else if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	root := flag.String("root", ".", "project root")
	image := flag.String("image", "server:local", "locally built image available on each cluster node")
	check := flag.Bool("check", false, "verify generated manifests without changing files")
	flag.Parse()
	if err := run(*root, "server", *image, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
