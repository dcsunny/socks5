package socks5

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
)

// ConnectViaProxy connects to the target through an proxy
func ConnectViaProxy(proxyAddr, targetHost, targetPort string) (net.Conn, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		log.Printf("Error parsing proxy URL:%s", err)
		return nil, err
	}
	remoteAddr := fmt.Sprintf("%s:%s", targetHost, targetPort)
	dialer := &ProxyDialer{ProxyUrl: proxyURL}
	var conn net.Conn
	conn, err = dialer.Dial("tcp", remoteAddr)
	if err != nil {
		log.Print(err)
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
