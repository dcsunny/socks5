// Package main provides a SOCKS5 proxy server that can use system proxy settings
package main

import (
	"fmt"
	"runtime"
)

const (
	socks5Version = uint8(5)
)

// ProxyInfo stores proxy configuration
type ProxyInfo struct {
	ProxyType string // http, socks5, etc.
	Host      string
	Port      string
	Enabled   bool
}

// getSystemProxy retrieves the system proxy settings
func getSystemProxy() (*ProxyInfo, error) {
	goos := runtime.GOOS
	if goos == "darwin" || goos == "windows" || goos == "linux" {
		return getProxy()
	}
	return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}
