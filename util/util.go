/*
Provide Utility functions
*/
package util

import (
	"errors"
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
