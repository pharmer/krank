package packet

import (
	"io"

	netutil "github.com/appscode/go/net"
	"github.com/appscode/go/os/script"
	"github.com/appscode/krank/cloud"
	"github.com/appscode/pharmer/api"
)

func init() {
	cloud.RegisterCloudProvider("packet-debian", func(config io.Reader) (cloud.KubeStarter, error) {
		return new(debian), nil
	})
}

type debian struct {
}

func (*debian) Run(req *api.ClusterStartupConfig) error {
	iface, nodeIP, err := netutil.NodeIP("bond0")
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
		cloud.Getent("", "Waiting for functional DNS (trying to resolve my own IP)...")
	}
	return script.Perform()
}
