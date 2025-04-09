//go:build linux
// +build linux

package socks5

import (
	"os"
)

// getProxy retrieves proxy settings from Linux
func getProxy() (*ProxyInfo, error) {
	// Check environment variables
	socksProxy := os.Getenv("SOCKS_PROXY")
	if socksProxy == "" {
		socksProxy = os.Getenv("socks_proxy")
	}
	if socksProxy != "" {
		return &ProxyInfo{
			ProxyType: "socks5",
			Addr:      socksProxy,
			Enabled:   true,
		}, nil
	}

	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy == "" {
		httpProxy = os.Getenv("http_proxy")
	}

	if httpProxy != "" {
		return &ProxyInfo{
			ProxyType: "http",
			Addr:      httpProxy,
			Enabled:   true,
		}, nil
	}

	httpsProxy := os.Getenv("HTTPS_PROXY")
	if httpsProxy == "" {
		httpsProxy = os.Getenv("https_proxy")
	}

	if httpsProxy != "" {
		return &ProxyInfo{
			ProxyType: "https",
			Addr:      httpsProxy,
			Enabled:   true,
		}, nil
	}

	return nil, nil
}
