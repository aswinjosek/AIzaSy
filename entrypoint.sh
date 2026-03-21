#!/bin/sh
set -e

WG_CONF="/etc/wireguard/wg0.conf"
mkdir -p /etc/wireguard

if[ ! -f "$WG_CONF" ]; then
    echo "==> [WARP] 未检测到配置，正在全自动初始化 Cloudflare WARP..."
    
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) WGCF_ARCH="amd64" ;;
        aarch64) WGCF_ARCH="arm64" ;;
        *) echo "==> [ERROR] 不支持的架构: $ARCH"; exit 1 ;;
    esac
    
    echo "==> [WARP] 当前系统架构为 $ARCH，正在下载对应的 wgcf ($WGCF_ARCH)..."
    wget -qO wgcf "https://github.com/ViRb3/wgcf/releases/download/v2.2.22/wgcf_2.2.22_linux_${WGCF_ARCH}"
    chmod +x wgcf
    
    echo "==>[WARP] 正在向 CF 注册设备..."
    if ! ./wgcf register --accept-tos; then
        echo "==> [ERROR] WARP 注册失败！"
        exit 1
    fi
    
    echo "==> [WARP] 正在生成 WireGuard 配置文件..."
    ./wgcf generate
    
    mv wgcf-profile.conf "$WG_CONF"
    rm -f wgcf wgcf-account.toml
    echo "==> [WARP] 配置生成成功！"
else
    echo "==> [WARP] 检测到已有配置，跳过注册。"
fi

# ==========================================
# 强力洗白与兼容性处理区
# ==========================================
# 1. 强制 AllowedIPs 仅接管 IPv4 流量
sed -i 's/^AllowedIPs.*/AllowedIPs = 0.0.0.0\/0/g' "$WG_CONF"
# 2. 删除 IPv6 Address 行
sed -i '/Address.*:/d' "$WG_CONF" 
# 3. 删除 DNS 行 (防 resolvconf 崩溃)
sed -i '/^DNS.*/d' "$WG_CONF"

# 4. 【终极暗黑魔法】摘除 wg-quick 中的 sysctl 只读报错触发器！
# 我们直接在 /usr/bin/wg-quick 源码中删掉涉及 src_valid_mark 的行
sed -i '/src_valid_mark/d' /usr/bin/wg-quick
# ==========================================

# 2. 拉起 Linux 内核 WireGuard 网卡
echo "==> [WARP] 正在启动 Linux 内核级 wg0 网卡..."
wg-quick up wg0

# 验证出口 IP 是否洗白
echo "==> [WARP] 当前出口 IP 已变更为："
wget -qO- https://1.1.1.1/cdn-cgi/trace | grep ip=

# 3. 启动 Go 原生网关
echo "==>[Gateway] 正在启动 AIzaSy 极速网关..."
exec ./gateway
