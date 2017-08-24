package main

import (
	"os"

	"github.com/appscode/krank/cmds"
	logs "github.com/appscode/log/golog"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
