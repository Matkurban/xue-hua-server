//go:build mage

package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerArchiveLayout(t *testing.T) {
	t.Chdir(t.TempDir())
	paths := []string{
		linuxAmd64Platform("server-api"),
		"config/server-api.yml",
		"start-config.yml",
		"_output/mage-linux-amd64",
	}
	for _, path := range paths {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("fixture"), 0o600))
	}
	require.NoError(t, zipUpload("server-linux-amd64.zip", paths))
	archive, err := zip.OpenReader("server-linux-amd64.zip")
	require.NoError(t, err)
	defer archive.Close()
	var names []string
	for _, file := range archive.File {
		names = append(names, file.Name)
		require.NotContains(t, file.Name, "elephant")
		require.NotContains(t, file.Name, "openim")
		if file.Name == "mage" || file.Name == linuxAmd64Platform("server-api") {
			require.Equal(t, os.FileMode(0o755), file.Mode().Perm())
		}
	}
	require.ElementsMatch(t, []string{
		linuxAmd64Platform("server-api"), "config/server-api.yml", "start-config.yml", "mage",
	}, names)
}

func TestArchiveMissingInputPreservesPreviousRelease(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("release.zip", []byte("previous-release"), 0o600))
	err := zipUpload("release.zip", []string{linuxAmd64Platform("server-api")})
	require.ErrorContains(t, err, "missing")
	content, err := os.ReadFile("release.zip")
	require.NoError(t, err)
	require.Equal(t, "previous-release", string(content))
}

func TestBuildOnlyRejectsUnknownTarget(t *testing.T) {
	_, _, err := splitTargets("does-not-exist")
	require.ErrorContains(t, err, "unknown build target")
	_, _, err = splitTargets("../server-api")
	require.ErrorContains(t, err, "invalid target")
}
