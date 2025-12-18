/*
Provide Utility functions
*/
package util

import (
	"errors"
	"net"
	"strconv"
)

func PortCheck(portStr string) (string, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", errors.New("invalid port syntax")
	}
	if port <= 0 || port > 65535 {
		return "", errors.New("invalid port range")
	}
	return strconv.Itoa(port), nil
}

func DNSLookUp(host string) ([]net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	return ips, nil
}
