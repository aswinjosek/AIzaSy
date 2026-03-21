#!/bin/sh
set -e

WG_CONF="/etc/wireguard/wg0.conf"
mkdir -p /etc/wireguard

# 1. 自动生成配置
if [ ! -f "$WG_CONF" ]; then
    echo "==> [WARP] 未检测到配置，正在全自动初始化 Cloudflare WARP..."
    
    wget -qO wgcf https://github.com/ViRb3/wgcf/releases/download/v2.2.22/wgcf_2.2.22_linux_amd64
    chmod +x wgcf
    
    echo "==> [WARP] 正在向 CF 注册设备..."
    if ! ./wgcf register --accept-tos; then
        echo "==> [ERROR] WARP 注册失败！"
        exit 1
    fi
    
    echo "==> [WARP] 正在生成 WireGuard 配置文件..."
    ./wgcf generate
    
    # 移动配置
    mv wgcf-profile.conf "$WG_CONF"
    
    # 【避坑关键】为了防止 Docker 内容器没有开启 IPv6 导致 wg-quick 报错退出
    # 我们将 AllowedIPs 强制改为仅接管 IPv4 流量 (这对访问 Google API 足够了)
    sed -i 's/^AllowedIPs.*/AllowedIPs = 0.0.0.0\/0/g' "$WG_CONF"
    sed -i '/Address.*:/d' "$WG_CONF" # 删除 IPv6 Address 行
    
    rm -f wgcf wgcf-account.toml
    echo "==> [WARP] 配置生成成功！"
else
    echo "==> [WARP] 检测到已有配置，跳过注册。"
fi

# 2. 拉起 Linux 内核 WireGuard 网卡
echo "==> [WARP] 正在启动 Linux 内核级 wg0 网卡..."
wg-quick up wg0

# 验证出口 IP 是否洗白
echo "==> [WARP] 当前出口 IP 已变更为："
wget -qO- https://1.1.1.1/cdn-cgi/trace | grep ip=

# 3. 启动 Go 原生网关
echo "==> [Gateway] 正在启动 AIzaSy 极速网关..."
exec ./gateway
