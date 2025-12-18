package model

import (
	"fmt"
	"net"
)

// IPMode represents preferred IP family selection.
type IPMode int

const (
	// IPAny allows both IPv4 and IPv6.
	IPAny IPMode = iota
	// IPv4Only restricts operations to IPv4.
	IPv4Only
	// IPv6Only restricts operations to IPv6.
	IPv6Only
)

// NewIPMode derives the mode from CLI flags and validates exclusivity.
func NewIPMode(forceIPv4, forceIPv6 bool) (IPMode, error) {
	switch {
	case forceIPv4 && forceIPv6:
		return IPAny, fmt.Errorf("cannot combine -4 and -6")
	case forceIPv4:
		return IPv4Only, nil
	case forceIPv6:
		return IPv6Only, nil
	default:
		return IPAny, nil
	}
}

// Network returns the appropriate network string for net.Dial/Listen.
func (m IPMode) Network(udp bool) string {
	switch m {
	case IPv4Only:
		if udp {
			return "udp4"
		}
		return "tcp4"
	case IPv6Only:
		if udp {
			return "udp6"
		}
		return "tcp6"
	default:
		if udp {
			return "udp"
		}
		return "tcp"
	}
}

// ResolveNetwork returns the network hint for address resolution.
func (m IPMode) ResolveNetwork() string {
	switch m {
	case IPv4Only:
		return "ip4"
	case IPv6Only:
		return "ip6"
	default:
		return "ip"
	}
}

// ValidateHost ensures literal addresses comply with the selected family and numeric-only expectations.
func (m IPMode) ValidateHost(host string, numericOnly bool) error {
	ip := net.ParseIP(host)
	if ip == nil {
		if numericOnly {
			return fmt.Errorf("numeric host required when -n is set")
		}
		return nil
	}

	switch m {
	case IPv4Only:
		if ip.To4() == nil {
			return fmt.Errorf("IPv4 address required when -4 is set")
		}
	case IPv6Only:
		if ip.To4() != nil {
			return fmt.Errorf("IPv6 address required when -6 is set")
		}
	}

	return nil
}
