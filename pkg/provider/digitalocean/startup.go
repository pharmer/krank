package digitalocean

import (
	"github.com/appscode/pharmer/pkg/contexts/env"
	"github.com/appscode/krank/pkg/provider"
	"github.com/appscode/krank/pkg/provider/extpoints"
	netutil "github.com/appscode/go/net"
	"github.com/appscode/go/os/script"
)

func init() {
	extpoints.KubeStarters.Register(new(debian), "digitalocean-debian")
}

type debian struct {
}

func (*debian) Run(req *env.ClusterStartupConfig) error {
	iface, nodeIP, err := netutil.NodeIP("eth1")
	if err != nil {
		return err
	}

	script := &provider.KubeScript{
		Script: &script.DebianGeneric{},
		Req:    req,
		HostData: provider.HostData{
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
		provider.Getent(script.CmdOut("_error_", "hostname", "-f"), "Waiting for functional DNS (trying to resolve my own FQDN)...")
		provider.Getent(script.CmdOut("_error_", "hostname", "-i"), "Waiting for functional DNS (trying to resolve my own IP)...")
	}
	return script.Perform()
}
