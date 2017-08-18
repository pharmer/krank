//go:generate go-extpoints ../../pkg/provider/extpoints
package main

import (
	"os"

	_ "github.com/appscode/krank/cloud/providers"
	"github.com/appscode/krank/cmds"
	logs "github.com/appscode/log/golog"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewCmdStartup().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
