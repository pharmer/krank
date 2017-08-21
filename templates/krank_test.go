package templates

import (
	"fmt"
	"net"
	"os"
	"strings"
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
			fmt.Println(ip, addr.Network())
			// process IP address
		}
	}
}

func TestLp(t *testing.T) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.To16())
				os.Stdout.WriteString(ipnet.IP.String() + "\n")
			}

		}
	}
}

func TestUDP(t *testing.T) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	fmt.Println(err)
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	fmt.Println(localAddr[0:idx])
}
