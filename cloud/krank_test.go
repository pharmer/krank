package cloud

import (
	"fmt"
	"net"
	"testing"
)

func TestIP(t *testing.T) {
	ifaces, err := net.Interfaces()
	fmt.Println(err)
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		fmt.Println(err)
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			fmt.Println(ip)
			// process IP address
		}
	}
}
