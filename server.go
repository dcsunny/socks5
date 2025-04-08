package socks5

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// DownstreamProxyInfo stores downstream proxy configuration
type DownstreamProxyInfo struct {
	ProxyType string // http, socks5, etc.
	Addr      string
	Enabled   bool
}

// Global variable to store downstream proxy configuration
var downstreamProxy *DownstreamProxyInfo

func Server(listenAddr, downProxy string) {
	// Configure downstream proxy if provided
	if downProxy != "" {
		downstreamProxy = &DownstreamProxyInfo{
			Addr:    downProxy,
			Enabled: true,
		}
		fmt.Println(downProxy)
		if strings.HasPrefix(downProxy, "socks5") {
			downstreamProxy.ProxyType = "socks5"
		} else if strings.HasPrefix(downProxy, "https") {
			downstreamProxy.ProxyType = "https"
		} else if strings.HasPrefix(downProxy, "http") {
			downstreamProxy.ProxyType = "http"
		} else {
			log.Fatalf("Unsupported downstream proxy type: %s", downProxy)
		}
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}
	defer listener.Close()

	log.Printf("SOCKS5 proxy server started on %s", listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	bufConn := bufio.NewReader(conn)

	// Read the version and number of authentication methods
	versionByte, err := bufConn.ReadByte()
	if err != nil {
		log.Printf("Failed to read version byte: %v", err)
		return
	}
	version := uint8(versionByte)
	if version != Socks5Version {
		log.Printf("Unsupported SOCKS version: %d (0x%02X)", version, version)
		return
	}

	nMethodsByte, err := bufConn.ReadByte()
	if err != nil {
		log.Printf("Failed to read number of methods: %v", err)
		return
	}
	nMethods := uint8(nMethodsByte)

	// Read the authentication methods
	methods := make([]byte, nMethods)
	_, err = bufConn.Read(methods)
	if err != nil {
		log.Printf("Failed to read methods: %v", err)
		return
	}

	// We only support no authentication
	if !Contains(methods, 0) {
		log.Printf("No supported authentication methods")
		return
	}

	// Send the response
	_, err = conn.Write([]byte{Socks5Version, 0})
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

	// Read the request
	request := make([]byte, 4)
	_, err = bufConn.Read(request)
	if err != nil {
		log.Printf("Failed to read request: %v", err)
		return
	}

	cmd := request[1]
	if cmd != 1 { // 1 is for TCP/IP stream connection
		log.Printf("Unsupported command: %d", cmd)
		return
	}

	// Read the address type
	addrType := request[3]
	var targetHost string
	var targetPort string

	switch addrType {
	case 1: // IPv4
		ip := make([]byte, 4)
		_, err = io.ReadFull(bufConn, ip)
		if err != nil {
			log.Printf("Failed to read IPv4 address: %v", err)
			return
		}
		targetHost = net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
	case 3: // Domain name
		length, err := bufConn.ReadByte()
		if err != nil {
			log.Printf("Failed to read domain length: %v", err)
			return
		}
		domain := make([]byte, length)
		_, err = io.ReadFull(bufConn, domain)
		if err != nil {
			log.Printf("Failed to read domain: %v", err)
			return
		}
		targetHost = string(domain)
	case 4: // IPv6
		ip := make([]byte, 16)
		_, err = io.ReadFull(bufConn, ip)
		if err != nil {
			log.Printf("Failed to read IPv6 address: %v", err)
			return
		}
		targetHost = net.IP(ip).String()
	default:
		log.Printf("Unsupported address type: %d", addrType)
		return
	}

	// Read the port
	portBytes := make([]byte, 2)
	_, err = io.ReadFull(bufConn, portBytes)
	if err != nil {
		log.Printf("Failed to read port: %v", err)
		return
	}
	targetPort = fmt.Sprintf("%d", (int(portBytes[0])<<8)|int(portBytes[1]))

	var targetConn net.Conn

	if downstreamProxy != nil && downstreamProxy.Enabled {
		log.Printf("Using downstream proxy: %s", downstreamProxy.Addr)
		switch downstreamProxy.ProxyType {
		case "http", "https":
			targetConn, err = ConnectViaHttpProxy(downstreamProxy.Addr, targetHost, targetPort)
		case "socks5":
			targetConn, err = ConnectViaSocks5Proxy(downstreamProxy.Addr, targetHost, targetPort)
		default:
			err = fmt.Errorf("unsupported downstream proxy type: %s", downstreamProxy.ProxyType)
		}
	} else {
		sysProxy, err := GetSystemProxy()
		if err != nil {
			log.Printf("Failed to get system proxy: %v", err)
		}
		if sysProxy != nil && sysProxy.Enabled {
			log.Printf("Using system proxy: %s", sysProxy.Addr)
			switch sysProxy.ProxyType {
			case "http", "https":
				targetConn, err = ConnectViaHttpProxy(sysProxy.Addr, targetHost, targetPort)
			case "socks5":
				targetConn, err = ConnectViaSocks5Proxy(sysProxy.Addr, targetHost, targetPort)
			default:
				log.Printf("Unsupported proxy type: %s, connecting directly", sysProxy.ProxyType)
				targetConn, err = net.Dial("tcp", net.JoinHostPort(targetHost, targetPort))
			}
		} else {
			// Connect directly if no system proxy is enabled
			targetConn, err = net.Dial("tcp", net.JoinHostPort(targetHost, targetPort))
		}
	}

	if err != nil {
		log.Printf("Failed to connect to target: %v", err)
		return
	}
	if targetConn == nil {
		return
	}
	// Send the response - BND.ADDR and BND.PORT should be the address and port that the proxy is using
	// Get the local address that the proxy is using for this connection
	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	ip := localAddr.IP.To4()
	port := uint16(localAddr.Port)

	// Prepare the response
	response := make([]byte, 10)
	response[0] = Socks5Version // VER
	response[1] = 0             // REP - 0 for success
	response[2] = 0             // RSV - reserved, must be 0

	if ip != nil {
		// IPv4
		response[3] = 1         // ATYP - 1 for IPv4
		copy(response[4:8], ip) // BND.ADDR
	} else {
		// If not IPv4, use 0.0.0.0 as placeholder
		response[3] = 1 // ATYP - 1 for IPv4
		response[4] = 0 // BND.ADDR
		response[5] = 0
		response[6] = 0
		response[7] = 0
	}

	// BND.PORT in network byte order
	response[8] = byte(port >> 8)
	response[9] = byte(port & 0xff)

	_, err = conn.Write(response)
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

	// Start proxying data with proper synchronization
	var wg sync.WaitGroup
	wg.Add(2)

	// Copy from client to target
	go func() {
		defer wg.Done()
		defer targetConn.Close() // Close target connection when done
		_, err := io.Copy(targetConn, conn)
		if err != nil && !IsConnectionClosed(err) {
			log.Printf("Failed to copy data from client to target: %v", err)
		}
	}()

	// Copy from target to client
	go func() {
		defer wg.Done()
		defer conn.Close() // Close client connection when done
		_, err := io.Copy(conn, targetConn)
		if err != nil && !IsConnectionClosed(err) {
			log.Printf("Failed to copy data from target to client: %v", err)
		}
	}()

	// Wait for both copy operations to complete
	wg.Wait()
}
