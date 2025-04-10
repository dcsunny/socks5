# SOCKS5 代理服务器

一个轻量级、高性能的SOCKS5代理服务器，支持系统代理和下游代理配置。

## 特性

- 支持标准SOCKS5协议
- 支持用户名/密码认证
- 自动检测并使用系统代理设置
- 支持配置下游代理（SOCKS5/HTTP/HTTPS）
- 跨平台支持（Windows/Linux/macOS）
- 轻量级设计，低资源占用
- 支持命令行参数配置

## 安装

```bash
go install github.com/dcsunny/socks5
```

## 使用方法

### 命令行参数

```bash
# 基本用法
socks5 [flags]

# 可用参数
-l, --listen string        监听地址 (默认 "0.0.0.0:21080")
-d, --down-proxy string    下游代理地址 (例如: "socks5://127.0.0.1:1080" 或 "http://127.0.0.1:8080")
-s, --system-proxy        是否使用系统代理 (默认 true)
-u, --username string     认证用户名
-p, --password string     认证密码
```

### 示例

1. 启动基本代理服务器：
```bash
socks5
```

2. 指定监听地址：
```bash
socks5 -l 127.0.0.1:1080
```

3. 使用下游SOCKS5代理：
```bash
socks5 -d socks5://127.0.0.1:1080
```

4. 使用下游HTTP代理：
```bash
socks5 -d http://127.0.0.1:8080
```

5. 禁用系统代理：
```bash
socks5 -s=false
```

6. 启用用户名密码认证：
```bash
socks5 -u admin -p password123
```

## sdk 调用
### 示例
``` go

package main

import "github.com/dcsunny/socks5"

func main() {
	socksAddr := "0.0.0.0:21080"
	s := socks5.NewServer(true, socksAddr, "")
	s.Run()
}

```

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

