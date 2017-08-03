package azure

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/appscode/go/os/script"
	"github.com/appscode/krank/pkg/provider"
	"github.com/appscode/krank/pkg/provider/extpoints"
	"github.com/appscode/pharmer/pkg/contexts/env"
)

func init() {
	extpoints.KubeStarters.Register(new(debian), "azure-debian")
}

type debian struct {
}

func (*debian) Run(req *env.ClusterStartupConfig) error {
	script := &provider.KubeScript{
		Script: &script.DebianGeneric{},
		Req:    req,
	}
	script.HostData.APIServers = req.KubernetesMasterName

	script.EnsureBasicNetworking = func() {
		provider.Getent(script.CmdOut("_error_", "hostname", "-f"), "Waiting for functional DNS (trying to resolve my own FQDN)...")
		// provider.Getent(script.CmdOut("_error_", "hostname", "-i"), "Waiting for functional DNS (trying to resolve my own IP)...")
	}
	script.WriteCloudConfig = func() error {
		if script.Req.CloudConfigPath == "" || script.Req.AzureCloudConfig == nil {
			return nil
		}
		script.Mkdir(filepath.Dir(script.Req.CloudConfigPath))

		data, err := json.MarshalIndent(script.Req.AzureCloudConfig, "", "  ")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(script.Req.CloudConfigPath, data, 0644)
	}
	return script.Perform()
}
