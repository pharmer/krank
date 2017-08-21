package cloud

import (
	"net"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/api/v1"
)

func (k *KrankCloudProvider) NodeAddresses(name types.NodeName) ([]v1.NodeAddress, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []v1.NodeAddress{}, err
	}
	ip := make([]string, 0)
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = append(ip, ipnet.IP.String())
			}

		}
	}
	return []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: ip[0]},
		{Type: v1.NodeExternalIP, Address: ip[1]},
	}, nil
}
