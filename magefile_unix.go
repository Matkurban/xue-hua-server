//go:build mage && !windows
// +build mage,!windows

package main

import (
	"syscall"
)

func setMaxOpenFiles() error {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	rLimit.Cur = min(rLimit.Max, 10000)
	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}
