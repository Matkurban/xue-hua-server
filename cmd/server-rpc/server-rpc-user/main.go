package main

import (
	"gitee.com/matkurban/openimsdk-tools/system/program"
	"server/v3/pkg/common/cmd"
)

func main() {
	if err := cmd.NewUserRpcCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
