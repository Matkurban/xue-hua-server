package main

import (
	"server/v3/pkg/common/program"
	_ "net/http/pprof"
	"server/v3/pkg/common/cmd"
)

func main() {
	if err := cmd.NewApiCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
