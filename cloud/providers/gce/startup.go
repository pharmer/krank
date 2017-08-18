package gce

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/appscode/go/os/script"
	"github.com/appscode/krank/cloud"
	"github.com/appscode/pharmer/api"
	"github.com/go-ini/ini"
)

func init() {
	cloud.RegisterCloudProvider("gce-debian", func(config io.Reader) (cloud.KubeStarter, error) {
		return new(debian), nil
	})
}

type debian struct {
}

func (*debian) Run(req *api.ClusterStartupConfig) error {
	script := &cloud.KubeScript{
		Script: &script.DebianGeneric{},
		Req:    req,
	}
	script.HostData.APIServers = req.KubernetesMasterName
	script.EnsureBasicNetworking = func() {
		cloud.Getent("metadata.google.internal", "Waiting for functional DNS (trying to resolve metadata.google.internal)...")
		cloud.Getent(script.CmdOut("_error_", "hostname", "-f"), "Waiting for functional DNS (trying to resolve my own FQDN)...")
		cloud.Getent(script.CmdOut("_error_", "hostname", "-i"), "Waiting for functional DNS (trying to resolve my own IP)...")
	}
	script.FindMasterPd = func() (string, bool) {
		// ref: https://github.com/kubernetes/kubernetes/blob/c406665b2b1fdec98cd321c427896f6e4b959530/cluster/gce/configure-vm.sh#L264
		devicepath := "/dev/disk/by-id/google-master-pd"
		if _, err := os.Stat(devicepath); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, devicepath+" does not exist")
			// path does not exist
			return "", false
		}

		outBytes, _ := script.Shell.Command("ls", "-l", devicepath).Output()
		out := string(outBytes)
		out = strings.TrimSpace(out)
		relativePath := "/dev/disk/by-id/" + out[strings.LastIndex(out, " ")+1:]
		path, err := filepath.Abs(relativePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine absolute path for %v:%v\n", relativePath, err.Error())
			return "", false
		}
		fmt.Println("Found master disk:", path)
		return path, true
	}
	script.WriteCloudConfig = func() error {
		if script.Req.CloudConfigPath == "" || script.Req.GCECloudConfig == nil {
			return nil
		}
		script.Mkdir(filepath.Dir(script.Req.CloudConfigPath))

		// ref: https://github.com/kubernetes/kubernetes/blob/release-1.5/cluster/gce/configure-vm.sh#L846
		cfg := ini.Empty()
		err := cfg.Section("global").ReflectFrom(script.Req.GCECloudConfig)
		if err != nil {
			return err
		}
		return cfg.SaveTo(script.Req.CloudConfigPath)
	}
	return script.Perform()
}
