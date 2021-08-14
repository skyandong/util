package consul

import (
	"errors"
	"net"
)

// ErrNotFound an IP address
var ErrNotFound = errors.New("ip address not found")

// LocalIP address
func LocalIP() (ip string, err error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, a := range as {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				return
			}
		}
	}
	err = ErrNotFound
	return
}
