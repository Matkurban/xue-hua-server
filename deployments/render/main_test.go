package main

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigMapUsesCurrentRuntimeConfiguration(t *testing.T) {
	config, err := configMap("../..", "server")
	if err != nil {
		t.Fatal(err)
	}
	data := config["data"].(map[string]string)
	if len(data) != len(services)+10 {
		t.Fatalf("unexpected number of runtime configurations: %d", len(data))
	}
	for _, service := range services {
		var settings object
		if err := yaml.Unmarshal([]byte(data[service.name+".yml"]), &settings); err != nil {
			t.Fatal(err)
		}
		if service.rpc != 0 {
			rpc := settings["rpc"].(map[string]any)
			if rpc["autoSetPorts"] != false || rpc["ports"].([]any)[0] != service.rpc {
				t.Errorf("%s has inconsistent RPC ports: %v", service.name, rpc)
			}
		}
		if service.metrics != 0 {
			metrics := settings["prometheus"].(map[string]any)
			if metrics["autoSetPorts"] != false || metrics["ports"].([]any)[0] != service.metrics {
				t.Errorf("%s has inconsistent metrics ports: %v", service.name, metrics)
			}
		}
	}
	var discovery object
	if err := yaml.Unmarshal([]byte(data["discovery.yml"]), &discovery); err != nil {
		t.Fatal(err)
	}
	for _, service := range services {
		if service.discovery != "" && discovery["rpcService"].(map[string]any)[service.discovery] != service.name {
			t.Errorf("discovery missing %s", service.name)
		}
	}
}

func TestDeploymentAndServiceContracts(t *testing.T) {
	for _, service := range services {
		t.Run(service.name, func(t *testing.T) {
			deployment, svc := manifests(service, "server", "server:local")
			spec := deployment["spec"].(object)
			template := spec["template"].(object)
			label := template["metadata"].(object)["labels"].(object)["app"]
			if label != service.name || spec["selector"].(object)["matchLabels"].(object)["app"] != label {
				t.Fatal("deployment selector does not match pod")
			}
			pod := template["spec"].(object)
			if pod["automountServiceAccountToken"] != true || pod["serviceAccountName"] != "server" {
				t.Fatal("Kubernetes detection requires the server service account mount")
			}
			container := pod["containers"].([]object)[0]
			if container["image"] != "server:local" || container["command"].([]string)[0] != "/server/bin/"+service.name {
				t.Fatal("deployment does not execute the locally built service")
			}
			if svc == nil {
				if service.rpc != 0 || service.metrics != 0 || service.public != 0 {
					t.Fatal("missing service")
				}
				return
			}
			svcSpec := svc["spec"].(object)
			if svcSpec["selector"].(object)["app"] != label {
				t.Fatal("service selector does not match pod")
			}
			ports := container["ports"].([]object)
			svcPorts := svcSpec["ports"].([]object)
			for i, port := range ports {
				if port["name"] != svcPorts[i]["targetPort"] || port["containerPort"] != svcPorts[i]["port"] {
					t.Fatal("service port does not match container")
				}
			}
			if service.rpc != 0 && (ports[0]["name"] != "grpc" || svcPorts[0]["name"] != "grpc") {
				t.Fatal("discovery requires grpc as the first named port")
			}
		})
	}
}
