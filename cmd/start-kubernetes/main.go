//go:generate go-extpoints ../../pkg/provider/extpoints
package main

import (
	"os"

	"appscode.com/krank/pkg/cmd"
	_ "appscode.com/krank/pkg/provider/aws"
	_ "appscode.com/krank/pkg/provider/azure"
	_ "appscode.com/krank/pkg/provider/digitalocean"
	_ "appscode.com/krank/pkg/provider/gce"
	_ "appscode.com/krank/pkg/provider/generic"
	_ "appscode.com/krank/pkg/provider/linode"
	_ "appscode.com/krank/pkg/provider/packet"
	v "github.com/appscode/go/version"
	logs "github.com/appscode/log/golog"
)

var (
	Version         string
	VersionStrategy string
	Os              string
	Arch            string
	CommitHash      string
	GitBranch       string
	GitTag          string
	CommitTimestamp string
	BuildTimestamp  string
	BuildHost       string
	BuildHostOs     string
	BuildHostArch   string
)

func init() {
	v.Version.Version = Version
	v.Version.VersionStrategy = VersionStrategy
	v.Version.Os = Os
	v.Version.Arch = Arch
	v.Version.CommitHash = CommitHash
	v.Version.GitBranch = GitBranch
	v.Version.GitTag = GitTag
	v.Version.CommitTimestamp = CommitTimestamp
	v.Version.BuildTimestamp = BuildTimestamp
	v.Version.BuildHost = BuildHost
	v.Version.BuildHostOs = BuildHostOs
	v.Version.BuildHostArch = BuildHostArch
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmd.NewCmdStartup().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
