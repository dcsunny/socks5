package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// connectViaHttpProxy connects to the target through an HTTP proxy
func connectViaHttpProxy(proxyHost, proxyPort, targetHost, targetPort string) (net.Conn, error) {
	// Connect to the HTTP proxy
	proxyConn, err := net.Dial("tcp", net.JoinHostPort(proxyHost, proxyPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP proxy: %v", err)
	}

	// Send HTTP CONNECT request
	connectReq := fmt.Sprintf(
		"CONNECT %s:%s HTTP/1.1\r\n"+
			"Host: %s:%s\r\n"+
			"User-Agent: Go-SOCKS5-Proxy\r\n"+
			"Proxy-Connection: keep-alive\r\n\r\n",
		targetHost, targetPort, targetHost, targetPort)

	_, err = proxyConn.Write([]byte(connectReq))
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %v", err)
	}

	// Read the response
	bufConn := bufio.NewReader(proxyConn)
	respLine, err := bufConn.ReadString('\n')
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %v", err)
	}

	// Check if the connection was established
	if !strings.Contains(respLine, "200") {
		proxyConn.Close()
		return nil, fmt.Errorf("proxy connection failed: %s", strings.TrimSpace(respLine))
	}

	// Skip the rest of the headers
	for {
		line, err := bufConn.ReadString('\n')
		if err != nil {
			proxyConn.Close()
			return nil, fmt.Errorf("failed to read proxy response headers: %v", err)
		}

		if line == "\r\n" || line == "\n" {
			break
		}
	}

	return proxyConn, nil
}

// connectViaSocks5Proxy connects to the target through a SOCKS5 proxy
func connectViaSocks5Proxy(proxyHost, proxyPort, targetHost, targetPort string) (net.Conn, error) {
	// Connect to the SOCKS5 proxy
	proxyConn, err := net.Dial("tcp", net.JoinHostPort(proxyHost, proxyPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SOCKS5 proxy: %v", err)
	}

	// SOCKS5 handshake
	// Send version and authentication methods
	_, err = proxyConn.Write([]byte{socks5Version, 1, 0}) // Version 5, 1 method, no authentication
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to send SOCKS5 handshake: %v", err)
	}

	// Read server's response
	resp := make([]byte, 2)
	_, err = io.ReadFull(proxyConn, resp)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read SOCKS5 handshake response: %v", err)
	}

	if resp[0] != socks5Version || resp[1] != 0 {
		proxyConn.Close()
		return nil, fmt.Errorf("SOCKS5 handshake failed: %v", resp)
	}

	// Send connection request
	request := make([]byte, 0, 10)
	request = append(request, socks5Version) // Version
	request = append(request, 1)             // Command: connect
	request = append(request, 0)             // Reserved

	// Add address type and address
	ip := net.ParseIP(targetHost)
	if ip == nil {
		// Domain name
		request = append(request, 3)                     // Address type: domain name
		request = append(request, byte(len(targetHost))) // Domain length
		request = append(request, []byte(targetHost)...) // Domain
	} else if ip.To4() != nil {
		// IPv4
		request = append(request, 1)           // Address type: IPv4
		request = append(request, ip.To4()...) // IPv4 address
	} else {
		// IPv6
		request = append(request, 4)            // Address type: IPv6
		request = append(request, ip.To16()...) // IPv6 address
	}

	// Add port in network byte order
	port, _ := strconv.Atoi(targetPort)
	request = append(request, byte(port>>8), byte(port&0xff))

	// Send the request
	_, err = proxyConn.Write(request)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to send SOCKS5 connect request: %v", err)
	}

	// Read the response
	response := make([]byte, 4)
	_, err = io.ReadFull(proxyConn, response)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read SOCKS5 connect response: %v", err)
	}

	if response[0] != socks5Version {
		proxyConn.Close()
		return nil, fmt.Errorf("unexpected SOCKS5 version: %v", response[0])
	}

	if response[1] != 0 {
		proxyConn.Close()
		return nil, fmt.Errorf("SOCKS5 connect failed with code: %v", response[1])
	}

	// Skip the bound address and port
	switch response[3] {
	case 1: // IPv4
		_, err = io.ReadFull(proxyConn, make([]byte, 4+2)) // 4 bytes for IPv4 + 2 bytes for port
	case 3: // Domain name
		length := make([]byte, 1)
		_, err = io.ReadFull(proxyConn, length)
		if err == nil {
			_, err = io.ReadFull(proxyConn, make([]byte, int(length[0])+2)) // Domain length + 2 bytes for port
		}
	case 4: // IPv6
		_, err = io.ReadFull(proxyConn, make([]byte, 16+2)) // 16 bytes for IPv6 + 2 bytes for port
	}

	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read SOCKS5 bound address: %v", err)
	}

	return proxyConn, nil
}

// isConnectionClosed checks if an error is related to a closed connection
func isConnectionClosed(err error) bool {
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

// contains checks if a byte array contains a specific value
func contains(arr []byte, val byte) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
