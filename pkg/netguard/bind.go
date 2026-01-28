package netguard

import (
	"fmt"
	"net"
	"strings"
)

// EnsureLocalOnly rejects non-loopback bind addresses.
func EnsureLocalOnly(addr string) error {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("refusing to bind non-loopback address %q; Autarch is local-only by default", addr)
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("refusing to bind non-loopback address %q; Autarch is local-only by default", host)
}
