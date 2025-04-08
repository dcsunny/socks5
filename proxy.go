// Package main provides a SOCKS5 proxy server that can use system proxy settings
package socks5

import (
	"fmt"
	"runtime"
	"time"
)

const (
	Socks5Version = uint8(5)
)

// ProxyInfo stores proxy configuration
type ProxyInfo struct {
	ProxyType string // http, socks5, etc.
	Addr      string
	Enabled   bool
}

// GetSystemProxy retrieves the system proxy settings
func GetSystemProxy() (*ProxyInfo, error) {
	cacheKey := "systemProxy"
	val := proxyCache.Get(cacheKey)
	if val != nil {
		return val.(*ProxyInfo), nil
	}
	proxyInfo, err := getSystemProxyNotCache()
	if err != nil {
		return nil, err
	}
	if proxyInfo == nil {
		proxyInfo = &ProxyInfo{
			Enabled: false,
		}
	}
	proxyCache.Put(cacheKey, proxyInfo, time.Second*5)
	return proxyInfo, nil
}

func getSystemProxyNotCache() (*ProxyInfo, error) {

	goos := runtime.GOOS
	if goos == "darwin" || goos == "windows" || goos == "linux" {
		return getProxy()
	}
	return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}

//func initCheckSystemCron() {
//	for {
//		time.Sleep(time.Second * 5)
//	}
//}
