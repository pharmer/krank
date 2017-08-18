package azure

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/appscode/go/os/script"
	"github.com/appscode/krank/cloud"
	"github.com/appscode/pharmer/api"
)

func init() {
	cloud.RegisterCloudProvider("azure-debian", func(config io.Reader) (cloud.KubeStarter, error) {
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
		cloud.Getent(script.CmdOut("_error_", "hostname", "-f"), "Waiting for functional DNS (trying to resolve my own FQDN)...")
		// cloud.Getent(script.CmdOut("_error_", "hostname", "-i"), "Waiting for functional DNS (trying to resolve my own IP)...")
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
