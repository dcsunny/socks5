# SOCKS5 代理服务器

一个支持系统代理设置的轻量级SOCKS5代理服务器，可以自动检测并使用系统配置的代理设置，同时支持配置下游代理。

## 功能特点

- 完整实现SOCKS5协议
- 支持自动检测并使用系统代理设置
- 支持配置下游SOCKS5或HTTP代理
- 跨平台支持：Windows、macOS和Linux
- 支持HTTP和SOCKS5下游代理
- 支持IPv4和IPv6地址
- 支持域名解析

## 系统要求

- Go 1.18或更高版本

## 安装

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/dcsunny/socks5.git
cd socks5

# 编译
go build

# 运行
./socks5
```

### 使用Go工具安装

```bash
go install github.com/dcsunny/socks5@latest
```

## 使用方法

### 基本用法

运行程序后，SOCKS5代理服务器将在本地127.0.0.1:21080端口启动。您可以将您的应用程序配置为使用此代理。

```bash
# 使用默认配置启动
./socks5
```

### 命令行参数

程序支持以下命令行参数：

- `-listen`: 指定监听地址和端口，默认为 `127.0.0.1:21080`
- `-upstream-type`: 下游代理类型，支持 `socks5` 和 `http`
- `-upstream-host`: 下游代理主机地址
- `-upstream-port`: 下游代理端口

### 使用下游代理

您可以通过命令行参数配置下游代理，例如：

```bash
# 使用下游SOCKS5代理
./socks5 -upstream-type=socks5 -upstream-host=proxy.example.com -upstream-port=1080

# 使用下游HTTP代理
./socks5 -upstream-type=http -upstream-host=proxy.example.com -upstream-port=8080

# 同时自定义监听地址
./socks5 -listen=0.0.0.0:1080 -upstream-type=socks5 -upstream-host=proxy.example.com -upstream-port=1080
```

### 代理检测

程序会自动检测系统代理设置：

- **Windows**: 从注册表读取系统代理设置
- **macOS**: 使用networksetup命令获取系统代理设置
- **Linux**: 从环境变量读取代理设置

### 配置应用程序

以下是一些常见应用程序的代理配置方法：

#### 浏览器

1. **Chrome/Edge**:
   - 设置 -> 高级 -> 系统 -> 打开代理设置
   - 手动设置代理 -> SOCKS主机：127.0.0.1，端口：21080

2. **Firefox**:
   - 设置 -> 网络设置 -> 配置代理访问
   - 手动代理配置 -> SOCKS主机：127.0.0.1，端口：21080，SOCKS v5

#### 命令行

```bash
# 使用curl通过SOCKS5代理
curl --socks5 127.0.0.1:21080 https://example.com

# 使用wget通过SOCKS5代理
wget --socks-proxy=127.0.0.1:21080 https://example.com

# 设置Git使用SOCKS5代理
git config --global http.proxy socks5://127.0.0.1:21080
```

## 工作原理

1. 程序启动时在指定端口（默认为本地21080端口）监听连接
2. 当接收到SOCKS5客户端连接请求时，进行SOCKS5协议握手
3. 检查是否配置了上游代理：
   - 如果配置了上游代理，则通过上游代理连接目标服务器
   - 如果未配置上游代理，则检测系统代理设置
4. 如果系统配置了代理，则通过系统代理连接目标服务器
5. 如果系统未配置代理，则直接连接目标服务器
6. 在客户端和目标服务器之间转发数据

### 代理链工作流程

当使用上游代理时，数据流向如下：

```
客户端应用 -> SOCKS5代理服务器 -> 上游代理服务器 -> 目标服务器
```

这种链式代理方式可以用于：
- 绕过网络限制
- 增加匿名性
- 在已有代理基础上添加本地代理

## 贡献

欢迎提交问题和拉取请求！

## 许可证

MIT