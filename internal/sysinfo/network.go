package sysinfo

import "net"

// InterfaceInfo holds the name and IPv4 address of a network interface.
type InterfaceInfo struct {
	Name string
	Addr string
}

// NetworkInterfaces returns all active, non-loopback network interfaces
// with at least one IPv4 address. An interface must be both administratively
// up (FlagUp) and have an active link (FlagRunning) to be included.
// This filters out bridges and virtual interfaces with no carrier
// (e.g., docker0, incusbr0 when no containers are connected).
func NetworkInterfaces() []InterfaceInfo {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var result []InterfaceInfo
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagRunning == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			// Only IPv4 addresses.
			if ipNet.IP.To4() == nil {
				continue
			}
			result = append(result, InterfaceInfo{
				Name: iface.Name,
				Addr: ipNet.IP.String(),
			})
		}
	}
	return result
}
