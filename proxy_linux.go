//go:build linux
// +build linux

package main

import (
	"net/url"
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
		proxyURL, err := url.Parse(httpProxy)
		if err != nil {
			return nil, err
		}

		host := proxyURL.Hostname()
		port := proxyURL.Port()
		if port == "" {
			port = "80"
		}

		return &ProxyInfo{
			ProxyType: "http",
			Host:      host,
			Port:      port,
			Enabled:   true,
		}, nil
	}

	if socksProxy != "" {
		proxyURL, err := url.Parse(socksProxy)
		if err != nil {
			return nil, err
		}

		host := proxyURL.Hostname()
		port := proxyURL.Port()
		if port == "" {
			port = "1080"
		}

		return &ProxyInfo{
			ProxyType: "socks5",
			Host:      host,
			Port:      port,
			Enabled:   true,
		}, nil
	}

	return nil, nil
}
