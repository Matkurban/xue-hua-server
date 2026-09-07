//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openimsdk/gomake/mageutil"
)

var Default = Build

var Aliases = map[string]any{
	"buildcc": BuildWithCustomConfig,
	"startcc": StartWithCustomConfig,
}

var (
	customRootDir   = "."
	customSrcDir    = "cmd"
	customOutputDir = "_output"
	customConfigDir = "config"
	customToolsDir  = "tools"
)

// Build compiles all server services and tools.
func Build() error {
	return mageutil.WithSpinnerR("Building binaries...", func() error {
		return mageutil.Build(nil, nil, nil, nil)
	})
}

// BuildOnly compiles space-separated targets, e.g. mage buildOnly "server-api seq".
func BuildOnly(targets string) error {
	services, tools, err := splitTargets(targets)
	if err != nil {
		return err
	}
	return mageutil.Build(services, tools, nil, nil)
}

func splitTargets(targets string) (services, tools []string, err error) {
	if strings.TrimSpace(targets) == "" {
		return nil, nil, fmt.Errorf("at least one target is required")
	}
	for _, name := range strings.Fields(targets) {
		if filepath.Base(name) != name || name == "." || name == ".." {
			return nil, nil, fmt.Errorf("invalid target %q", name)
		}
		if _, err := os.Stat(filepath.Join(customToolsDir, name, "main.go")); err == nil {
			tools = append(tools, name)
		} else if _, err := os.Stat(filepath.Join(customSrcDir, name, "main.go")); err == nil {
			services = append(services, name)
		} else if _, err := os.Stat(filepath.Join(customSrcDir, "server-rpc", name, "main.go")); err == nil {
			services = append(services, name)
		} else {
			return nil, nil, fmt.Errorf("unknown build target %q", name)
		}
	}
	return services, tools, nil
}

func BuildWithCustomConfig() error {
	return mageutil.Build(nil, nil, &mageutil.PathOptions{
		RootDir: &customRootDir, OutputDir: &customOutputDir,
		SrcDir: &customSrcDir, ToolsDir: &customToolsDir,
	}, nil)
}

func Start() error {
	return mageutil.WithSpinnerR("Starting...", func() error {
		return mageutil.StartToolsAndServices(nil, nil, nil)
	})
}

func StartWithCustomConfig() error {
	return mageutil.StartToolsAndServices(nil, nil, &mageutil.PathOptions{
		RootDir: &customRootDir, OutputDir: &customOutputDir, ConfigDir: &customConfigDir,
	})
}

func Stop() error {
	return mageutil.WithSpinnerR("Stopping...", mageutil.StopAndCheckServices)
}

func Check() error {
	return mageutil.WithSpinnerR("Checking...", mageutil.CheckAndReportServicesStatus)
}

// Export produces the same standalone Linux release as Pack.
func Export() error {
	return Pack()
}
