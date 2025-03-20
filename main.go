package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	socks5Version = uint8(5)
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:21080")
	if err != nil {
		log.Fatalf("Failed to listen on port 21080: %v", err)
	}
	defer listener.Close()

	log.Println("SOCKS5 proxy server started on 127.0.0.1:21080")

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

	bufConn := bufio.NewReader(conn)

	// Read the version and number of authentication methods
	versionByte, err := bufConn.ReadByte()
	if err != nil {
		log.Printf("Failed to read version byte: %v", err)
		return
	}
	version := uint8(versionByte)
	if version != socks5Version {
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
	if !contains(methods, 0) {
		log.Printf("No supported authentication methods")
		return
	}

	// Send the response
	_, err = conn.Write([]byte{socks5Version, 0})
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

	// Connect to the target
	targetConn, err := net.Dial("tcp", net.JoinHostPort(targetHost, targetPort))
	if err != nil {
		log.Printf("Failed to connect to target: %v", err)
		conn.Close()
		return
	}

	// Send the response - BND.ADDR and BND.PORT should be the address and port that the proxy is using
	// Get the local address that the proxy is using for this connection
	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	ip := localAddr.IP.To4()
	port := uint16(localAddr.Port)

	// Prepare the response
	response := make([]byte, 10)
	response[0] = socks5Version // VER
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
		if err != nil && !isConnectionClosed(err) {
			log.Printf("Failed to copy data from client to target: %v", err)
		}
	}()

	// Copy from target to client
	go func() {
		defer wg.Done()
		defer conn.Close() // Close client connection when done
		_, err := io.Copy(conn, targetConn)
		if err != nil && !isConnectionClosed(err) {
			log.Printf("Failed to copy data from target to client: %v", err)
		}
	}()

	// Wait for both copy operations to complete
	wg.Wait()
}

func contains(arr []byte, val byte) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
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
