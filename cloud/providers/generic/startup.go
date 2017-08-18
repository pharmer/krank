package generic

import (
	"io"

	netutil "github.com/appscode/go/net"
	"github.com/appscode/go/os/script"
	"github.com/appscode/krank/cloud"
	"github.com/appscode/pharmer/api"
)

func init() {
	cloud.RegisterCloudProvider("generic-debian", func(config io.Reader) (cloud.KubeStarter, error) {
		return new(debian), nil
	})
}

type debian struct {
}

func (*debian) Run(req *api.ClusterStartupConfig) error {
	iface, nodeIP, err := netutil.NodeIP()
	if err != nil {
		return err
	}

	script := &cloud.KubeScript{
		Script: &script.DebianGeneric{},
		Req:    req,
		HostData: cloud.HostData{
			HostnameOverride: nodeIP.String(),
			InternalIP:       nodeIP.String(),
			Iface:            iface,
		},
	}
	if !req.KubernetesMaster && req.ClusterInternalDomain != "" {
		script.HostData.APIServers = req.ClusterInternalDomain
	} else {
		script.HostData.APIServers = req.KubernetesMasterName
	}

	script.EnsureBasicNetworking = func() {
		cloud.Getent(script.CmdOut("_error_", "hostname", "-f"), "Waiting for functional DNS (trying to resolve my own FQDN)...")
		cloud.Getent(script.CmdOut("_error_", "hostname", "-i"), "Waiting for functional DNS (trying to resolve my own IP)...")
	}
	return script.Perform()
}
