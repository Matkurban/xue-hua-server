package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
)

func ExecuteCommand(cmdName string, args ...string) (string, error) {
	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing %s: %v, stderr: %s", cmdName, err, stderr.String())
	}
	return out.String(), nil
}

func getGoVersion() string {
	version := runtime.Version()
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return fmt.Sprintf("Go Version: %s\nOS: %s\nArchitecture: %s", version, goos, goarch)
}

func getDockerVersion() string {
	version, err := ExecuteCommand("docker", "--version")
	if err != nil {
		return "Docker is not installed. Please install it to get the version."
	}
	return version
}

func getKubernetesVersion() string {
	version, err := ExecuteCommand("kubectl", "version", "--client", "--short")
	if err != nil {
		return "Kubernetes is not installed. Please install it to get the version."
	}
	return version
}

func getGitVersion() string {
	version, err := ExecuteCommand("git", "branch", "--show-current")
	if err != nil {
		return "Git is not installed. Please install it to get the version."
	}
	return version
}

// // NOTE: You'll need to provide appropriate commands for Server versions.
// func getserverVersion() string {
// 	// Placeholder
// 	serverVersion := version.GetSingleVersion()
// 	return "Server Server: " + serverVersion + "\n"
// }

// func getServerClientVersion() (string, error) {
// 	serverClientVersion, err := version.GetClientVersion()
// 	if err != nil {
// 		return "", err
// 	}
// 	return "Server Client: " + serverClientVersion.ClientVersion + "\n", nil
// }

func main() {
	// red := color.New(color.FgRed).SprintFunc()
	//	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	//	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Println(blue("## Go Version"))
	fmt.Println(getGoVersion())
	fmt.Println(blue("## Branch Type"))
	fmt.Println(getGitVersion())
	fmt.Println(blue("## Docker Version"))
	fmt.Println(getDockerVersion())
	fmt.Println(blue("## Kubernetes Version"))
	fmt.Println(getKubernetesVersion())
	// fmt.Println(blue("## Server Versions"))
	// fmt.Println(getserverVersion())
	// clientVersion, err := getServerClientVersion()
	// if err != nil {
	// 	fmt.Println(red("Error getting Server Client Version: "), err)
	// } else {
	// 	fmt.Println(clientVersion)
	// }
}
