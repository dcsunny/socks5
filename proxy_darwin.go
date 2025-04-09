//go:build darwin
// +build darwin

package socks5

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// getProxy retrieves proxy settings from macOS
func getProxy() (*ProxyInfo, error) {
	// Check HTTP proxy first
	cmd := exec.Command("networksetup", "-getwebproxy", "Wi-Fi")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "Enabled: Yes") {
		// Parse the output to get proxy host and port
		hostRegex := regexp.MustCompile(`Server: (.+)`)
		portRegex := regexp.MustCompile(`Port: (\d+)`)

		hostMatches := hostRegex.FindStringSubmatch(string(output))
		portMatches := portRegex.FindStringSubmatch(string(output))

		if len(hostMatches) > 1 && len(portMatches) > 1 {
			return &ProxyInfo{
				ProxyType: "http",
				Addr:      fmt.Sprintf("http://%s:%s", hostMatches[1], portMatches[1]),
				Enabled:   true,
			}, nil
		}
	}

	// Check SOCKS proxy
	cmd = exec.Command("networksetup", "-getsocksfirewallproxy", "Wi-Fi")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "Enabled: Yes") {
		// Parse the output to get proxy host and port
		hostRegex := regexp.MustCompile(`Server: (.+)`)
		portRegex := regexp.MustCompile(`Port: (\d+)`)

		hostMatches := hostRegex.FindStringSubmatch(string(output))
		portMatches := portRegex.FindStringSubmatch(string(output))

		if len(hostMatches) > 1 && len(portMatches) > 1 {
			return &ProxyInfo{
				ProxyType: "socks5",
				Addr:      fmt.Sprintf("socks5://%s:%s", hostMatches[1], portMatches[1]),
				Enabled:   true,
			}, nil
		}
	}

	// Check environment variables
	return getEnvProxy()
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
