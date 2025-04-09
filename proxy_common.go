package socks5

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// ConnectViaHttpProxy connects to the target through an HTTP proxy
func ConnectViaHttpProxy(proxyAddr, targetHost, targetPort string) (net.Conn, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return nil, err
	}
	remoteAddr := fmt.Sprintf("%s:%s", targetHost, targetPort)
	dialer := &HttpProxyDialer{ProxyUrl: proxyURL}
	var conn net.Conn
	fmt.Println(remoteAddr)
	conn, err = dialer.Dial("tcp", remoteAddr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return conn, nil
}

// ConnectViaSocks5Proxy connects to the target through a SOCKS5 proxy
func ConnectViaSocks5Proxy(proxyAddr, targetHost, targetPort string) (net.Conn, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return nil, err
	}

	auth := &proxy.Auth{
		User: proxyURL.User.Username(),
	}
	auth.Password, _ = proxyURL.User.Password()
	var dialer proxy.Dialer
	dialer, err = proxy.SOCKS5("tcp", proxyURL.Host,
		auth,
		proxy.Direct,
	)
	if err != nil {
		return nil, err
	}
	remoteAddr := fmt.Sprintf("%s:%s", targetHost, targetPort)
	var conn net.Conn
	conn, err = dialer.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// IsConnectionClosed checks if an error is related to a closed connection
func IsConnectionClosed(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "i/o timeout")
}

// Contains checks if a byte array contains a specific value
func Contains(arr []byte, val byte) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
