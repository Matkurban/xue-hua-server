//go:build mage
// +build mage

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openimsdk/gomake/mageutil"
)

func init() {
	Aliases["zip"] = Pack
}

func linuxAmd64Platform(name string) string {
	return filepath.Join("_output", "bin", "platforms", "linux", "amd64", name)
}

func linuxAmd64Tool(name string) string {
	return filepath.Join("_output", "bin", "tools", "linux", "amd64", name)
}

var packLinuxAmd64Entries = []string{
	linuxAmd64Platform("server-api"),
	linuxAmd64Platform("server-rpc-user"),
	linuxAmd64Platform("server-rpc-auth"),
	linuxAmd64Platform("server-rpc-friend"),
	linuxAmd64Platform("server-rpc-group"),
	linuxAmd64Platform("server-rpc-msg"),
	linuxAmd64Platform("server-rpc-conversation"),
	linuxAmd64Platform("server-rpc-third"),
	linuxAmd64Platform("server-msggateway"),
	linuxAmd64Platform("server-msgtransfer"),
	linuxAmd64Platform("server-push"),
	linuxAmd64Platform("server-crontask"),
	linuxAmd64Platform("server-cmdutils"),
	linuxAmd64Tool("check-component"),
	linuxAmd64Tool("check-free-memory"),
	linuxAmd64Tool("seq"),
	"config",
	"docker-compose.yml",
	".env",
	"start-config.yml",
	"magefile.go",
	"magefile_unix.go",
	"mage_pack.go",
	"go.mod",
	"go.sum",
	"version",
	"install.sh",
	"README.md",
	"README_zh_CN.md",
	"LICENSE",
}

// Pack zips linux/amd64 binaries and the files a single server needs.
//
// Cross-compile first: `CGO_ENABLED=0 PLATFORMS=linux_amd64 mage`
// Output: `_output/server-linux-amd64.zip`
func Pack() error {
	out := filepath.Join("_output", "server-linux-amd64.zip")
	launcher := filepath.Join("_output", "mage-linux-amd64")
	cmd := exec.Command("mage", "-compile", launcher, "-goos", "linux", "-goarch", "amd64")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compile release launcher: %w", err)
	}
	if err := mageutil.WithSpinnerR("Packing "+out+"...", func() error {
		return zipUpload(out, append(append([]string{}, packLinuxAmd64Entries...), launcher))
	}); err != nil {
		return err
	}
	fmt.Println("wrote", out)
	return nil
}

func zipUpload(out string, entries []string) error {
	for _, e := range entries {
		if _, err := os.Stat(e); err != nil {
			if os.IsNotExist(err) && strings.Contains(filepath.ToSlash(e), "linux/amd64") {
				return fmt.Errorf("%s missing; run: CGO_ENABLED=0 PLATFORMS=linux_amd64 mage", e)
			}
			return fmt.Errorf("%s: %w", e, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}

	tmp := out + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(f)
	addErr := error(nil)
	for _, e := range entries {
		if err := addZipPath(zw, e); err != nil {
			addErr = err
			break
		}
	}
	closeZW := zw.Close()
	closeF := f.Close()
	if addErr != nil || closeZW != nil || closeF != nil {
		_ = os.Remove(tmp)
		if addErr != nil {
			return addErr
		}
		if closeZW != nil {
			return closeZW
		}
		return closeF
	}
	return os.Rename(tmp, out)
}

func addZipPath(zw *zip.Writer, src string) error {
	if filepath.ToSlash(src) == "_output/mage-linux-amd64" {
		return addZipFile(zw, src, "mage")
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return addZipFile(zw, src, filepath.ToSlash(src))
	}
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Name() == ".DS_Store" {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		name := filepath.ToSlash(path)
		return addZipFile(zw, path, name)
	})
}

func addZipFile(zw *zip.Writer, src, name string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = zip.Deflate
	if strings.HasPrefix(name, "_output/bin/") || name == "mage" || name == "install.sh" {
		header.SetMode(0o755)
	}

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
