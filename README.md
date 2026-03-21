# AIzaSy - Gemini API Gateway ⚡

![Docker Pulls](https://img.shields.io/badge/docker-ready-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Architecture](https://img.shields.io/badge/arch-amd64%20|%20arm64-orange.svg)

**AIzaSy** 是一个专为 Google Gemini API 设计的极简、高性能、绝对隐私优先的反向代理网关。

本项目诞生于一个极客目标：**如何在仅有 1 核 1GB 内存 (1C1G) 的低端 VPS 上，实现千万级高并发的 API 转发，并完美规避 Google 严苛的 IP 风控？**

最终，我们融合了 Go 语言并发调优、Docker 网络隔离与 Linux 内核级 WireGuard 路由，打造出了这套开箱即用的终极数据管道。

---

## ✨ 核心特性

*   🛡️ **自带 Cloudflare WARP 出口洗白**
    *   内置自动化脚本，首次启动时**全自动注册并提取 WARP 节点**。
    *   通过赋予容器 `NET_ADMIN` 权限，在 Docker 内部拉起 Linux 内核级 `wg0` 网卡。
    *   所有发往 Google 的 API 请求会被强制路由至 Cloudflare 骨干网，彻底隐藏 VPS 真实机房 IP，**完美免疫 Google WAF 封锁与网段清洗**。
*   🚀 **榨干 1C1G 的极限性能调优**
    *   放弃沉重的框架，采用经过极致优化的 Go 原生 `net/http`。
    *   引入 `sync.Pool` 内存对象复用，彻底拯救低配机器在高并发下的 GC 压力。
    *   内部隔离并注入 C100K 级别的 `sysctl` 操作系统网络参数。
*   🔒 **绝对隐私保护 (Zero-Knowledge Logging)**
    *   严格遵循 **BYOK (Bring Your Own Key)** 原则。
    *   **网关绝对不记录、不存储、不拦截任何 API Key。**
    *   底层日志输出采用“无差别全域抹杀”正则算法，无论你在 URL 中携带多长的私钥，控制台与日志中只会留下 `?key=***`，确保万无一失。
*   ⚡ **完美支持流式输出 (SSE)**
    *   针对 AI 大模型打字机效果深度优化，抛弃响应缓冲 (`FlushInterval = -1`)，实现零延迟数据透传。
*   🐳 **真正的零配置 (Zero-Config) 部署**
    *   一行代码拉起，无需配置繁琐的 WireGuard 参数，无需折腾 Go 编译环境。支持 `x86_64` 与 `arm64` 双架构。

---

## 🚀 快速部署 (Quick Start)

你只需要一台安装了 Docker 和 Docker Compose 的 Linux 服务器（需内核支持 WireGuard，主流 VPS 均满足）。

### 1. 创建 `docker-compose.yml`

创建一个新目录并写入以下配置文件：

```yaml
services:
  aizasy-api:
    image: ghcr.io/ccbkkb/aizasy:latest
    container_name: aizasy-gateway
    restart: always
    ports:
      - "8080:8080" # 左侧可修改为你宿主机想暴露的端口
    
    # 允许配置动态跨域 (留空或不写则默认允许所有来源 '*')
    environment:
      - CORS_ALLOWED_ORIGINS=*
    
    # 【核心：赋予容器操作内核网卡的权限】
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
      
    volumes:
      # 持久化 WARP 配置文件，防止重启后重新注册触发风控
      - warp-data:/etc/wireguard
      
    # 突破文件句柄限制
    ulimits:
      nofile:
        soft: 1048576
        hard: 1048576
        
    # 注入高并发内核参数与路由防崩溃标记
    sysctls:
      - net.ipv4.conf.all.src_valid_mark=1
      - net.core.somaxconn=65535
      - net.ipv4.tcp_tw_reuse=1
      - net.ipv4.ip_local_port_range=1024 65000
      - net.ipv4.tcp_keepalive_time=600

volumes:
  warp-data:
```

### 2. 一键启动

```bash
docker compose up -d
```
> **提示：** 首次启动时，容器会在后台向 Cloudflare 申请免费的 WARP 账户并生成配置，大约需要 3-5 秒。你可以通过 `docker logs -f aizasy-gateway` 观察优雅的启动日志。

---

## 📖 如何使用 (Usage)

部署成功并为你的服务器绑定域名后，只需在任何兼容 Gemini API 的客户端（如 LobeChat、ChatGPT-Next-Web）或代码中，将官方域名：
`generativelanguage.googleapis.com`

**替换为你的网关地址，例如：**
`https://gemini.aizasy.com` (换成你自己的域名)

**CURL 测试示例：**

```bash
curl -H 'Content-Type: application/json' \
     -X POST 'https://你的域名/v1beta/models/gemini-1.5-pro:generateContent?key=你的API_KEY' \
     -d '{
       "contents":[{"parts":[{"text": "用一句话解释量子计算。"}]}]
     }'
```

---

## 👨‍💻 鸣谢与开源声明

*   **Maintainer / Author**: [@ccbkkb](https://github.com/ccbkkb)
*   本项目完全开源，旨在为开发者社区提供稳定、干净的 AI 接口聚合基础设施。
*   **Privacy Pledge**: We act solely as a dumb pipe. Your data is your data.
*   **Linux.do**: The place where the author studies and lives ❤️ & the largest Chinese-language AI discussion community 🔥.

如果你觉得这个项目帮助到了你，欢迎点亮右上角的 ⭐️ **Star** ！
