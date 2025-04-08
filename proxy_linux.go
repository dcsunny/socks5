//go:build linux
// +build linux

package socks5

import (
	"os"
)

// getProxy retrieves proxy settings from Linux
func getProxy() (*ProxyInfo, error) {
	// Check environment variables
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy == "" {
		httpProxy = os.Getenv("http_proxy")
	}

	socksProxy := os.Getenv("SOCKS_PROXY")
	if socksProxy == "" {
		socksProxy = os.Getenv("socks_proxy")
	}

	if httpProxy != "" {
		return &ProxyInfo{
			ProxyType: "http",
			Addr:      httpProxy,
			Enabled:   true,
		}, nil
	}

	if socksProxy != "" {
		return &ProxyInfo{
			ProxyType: "socks5",
			Addr:      socksProxy,
			Enabled:   true,
		}, nil
	}

	return nil, nil
}
