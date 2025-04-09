package socks5

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type HttpProxyDialer struct {
	ProxyUrl *url.URL
}

func (s *HttpProxyDialer) Dial(network, addr string) (net.Conn, error) {
	var transport *http.Transport
	if s.ProxyUrl == nil {
		return nil, errors.New("not set proxy url")
	}
	transport = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			fmt.Printf("Using proxy: %v\n", s.ProxyUrl)
			return s.ProxyUrl, nil
		},
		DialContext: s.defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	conn, err := transport.DialContext(context.Background(), network, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *HttpProxyDialer) defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 使用 http.Transport 的 Proxy 字段获取代理地址
		proxyURL := s.ProxyUrl
		// 使用代理地址建立连接
		conn, err := dialer.DialContext(ctx, network, proxyURL.Host)
		if err != nil {
			return nil, err
		}
		// 发送 CONNECT 请求到代理服务器
		connectReq := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Host: addr},
			Header: make(http.Header),
		}
		if proxyURL.User != nil {
			username := proxyURL.User.Username()
			password, _ := proxyURL.User.Password()
			connectReq.Header.Set("Proxy-Authorization", "Basic "+basicAuth(username, password))
		}

		err = connectReq.Write(conn)
		if err != nil {
			conn.Close()
			return nil, err
		}

		br := bufio.NewReader(conn)
		resp, err := http.ReadResponse(br, connectReq)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if resp.StatusCode != 200 {
			conn.Close()
			return nil, fmt.Errorf("failed to connect to proxy: %v", resp.Status)
		}
		return conn, nil
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
