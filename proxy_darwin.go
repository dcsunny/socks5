//go:build darwin
// +build darwin

package main

import (
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
				Host:      hostMatches[1],
				Port:      portMatches[1],
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
				Host:      hostMatches[1],
				Port:      portMatches[1],
				Enabled:   true,
			}, nil
		}
	}

	// No proxy enabled
	return nil, nil
}
