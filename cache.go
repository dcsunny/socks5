package socks5

import "github.com/xmkuban/utils/cache"

var (
	// 缓存
	proxyCache = cache.NewMemoryCache()
)
