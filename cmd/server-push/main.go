package main

import (
	"server/v3/pkg/common/program"
	"server/v3/pkg/common/cmd"
)

func main() {
	if err := cmd.NewPushRpcCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
