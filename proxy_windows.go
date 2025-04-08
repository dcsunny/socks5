//go:build windows
// +build windows

package main

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// getProxy retrieves proxy settings from Windows registry
func getProxy() (*ProxyInfo, error) {
	// Open the registry key for Internet Settings
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %v", err)
	}
	defer key.Close()

	// Check if proxy is enabled
	proxyEnable, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil {
		return nil, fmt.Errorf("failed to get ProxyEnable value: %v", err)
	}

	// If proxy is not enabled, return nil
	if proxyEnable == 0 {
		return nil, nil
	}

	// Get proxy server address
	proxyServer, _, err := key.GetStringValue("ProxyServer")
	if err != nil {
		return nil, fmt.Errorf("failed to get ProxyServer value: %v", err)
	}

	// Check if the proxy server is empty
	if proxyServer == "" {
		return nil, nil
	}

	// Parse the proxy server string
	// It can be in the format "server:port" or "http=server:port;https=server:port;ftp=server:port;socks=server:port"
	var host, port, proxyType string

	// Check if it's a protocol-specific proxy configuration
	if strings.Contains(proxyServer, "=") {
		// Try to find SOCKS proxy first
		socksRegex := regexp.MustCompile(`socks=([^;:]+)(?::([0-9]+))?`)
		matches := socksRegex.FindStringSubmatch(proxyServer)
		if len(matches) > 1 {
			host = matches[1]
			if len(matches) > 2 && matches[2] != "" {
				port = matches[2]
			} else {
				port = "1080" // Default SOCKS port
			}
			proxyType = "socks5"
		} else {
			// Try to find HTTP proxy
			httpRegex := regexp.MustCompile(`http=([^;:]+)(?::([0-9]+))?`)
			matches = httpRegex.FindStringSubmatch(proxyServer)
			if len(matches) > 1 {
				host = matches[1]
				if len(matches) > 2 && matches[2] != "" {
					port = matches[2]
				} else {
					port = "80" // Default HTTP port
				}
				proxyType = "http"
			} else {
				// If no specific proxy found, return nil
				return nil, nil
			}
		}
	} else {
		// It's a simple "server:port" format
		parts := strings.Split(proxyServer, ":")
		host = parts[0]
		if len(parts) > 1 {
			port = parts[1]
		} else {
			port = "80" // Default to HTTP port
		}
		proxyType = "http" // Default to HTTP proxy type
	}

	// Return the proxy information
	return &ProxyInfo{
		ProxyType: proxyType,
		Host:      host,
		Port:      port,
		Enabled:   true,
	}, nil
}
