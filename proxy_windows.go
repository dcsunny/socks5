//go:build windows
// +build windows

package socks5

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// getProxy retrieves proxy settings from Windows registry and environment variables
func getProxy() (*ProxyInfo, error) {
	// First try to get proxy from Windows registry
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

	// If proxy is not enabled, try environment variables
	if proxyEnable == 0 {
		return getEnvProxy()
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
		Addr:      fmt.Sprintf("%s://%s:%s", proxyType, host, port),
		Enabled:   true,
	}, nil
}

// getEnvProxy retrieves proxy settings from environment variables
func getEnvProxy() (*ProxyInfo, error) {
	// Check SOCKS proxy first
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

	// Then check HTTP proxy
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

	// Finally check HTTPS proxy
	httpsProxy := os.Getenv("HTTPS_PROXY")
	if httpsProxy == "" {
		httpsProxy = os.Getenv("https_proxy")
	}
	if httpsProxy != "" {
		return &ProxyInfo{
			ProxyType: "http",
			Addr:      httpsProxy,
			Enabled:   true,
		}, nil
	}

	return nil, nil
}
