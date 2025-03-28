package querytracker

import (
	"encoding/base64"
	"errors"
	"net"
)

func decodeDNSIP(encodedStr string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return "", err
	}

	decodedStr := string(decodedBytes)
	ip := net.ParseIP(decodedStr)
	if ip == nil {
		return "", errors.New("invalid IP address")
	}
	return ip.String(), nil
}
