package socks5

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

type ProxyDialer struct {
	ProxyUrl *url.URL
}

func (s *ProxyDialer) Dial(network, addr string) (net.Conn, error) {
	var transport *http.Transport
	if s.ProxyUrl == nil {
		return nil, errors.New("not set proxy url")
	}
	dialer, err := proxy.FromURL(s.ProxyUrl, proxy.Direct)
	if err != nil {
		return nil, err
	}
	transport = &http.Transport{
		DialContext:           s.defaultTransportDialContext(dialer),
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

func (s *ProxyDialer) defaultTransportDialContext(dialer proxy.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
}
