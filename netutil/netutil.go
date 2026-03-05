// Package netutil provides networking utility functions.
package netutil

import "net"

// MustGetIP returns the first non-loopback IPv4 address of the local machine.
// Returns an empty string if no suitable address is found or an error occurs.
//
// This is useful for logging, service registration, or health-check endpoints
// that need to identify the host.
func MustGetIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		// Skip down or loopback interfaces.
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return ""
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Only return IPv4 addresses.
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}

	return ""
}
