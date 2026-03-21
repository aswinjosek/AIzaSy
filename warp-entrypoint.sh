#!/bin/sh
set -e

CONF_DIR="/etc/sing-box"
mkdir -p $CONF_DIR
cd $CONF_DIR

# 检查是否已经存在配置，如果有就跳过注册（防止重启容器触发限流）
if[ ! -f "config.json" ]; then
    echo "==> [WARP] 没有检测到配置，正在全自动初始化 Cloudflare WARP..."
    
    # 安装所需工具
    apk add --no-cache wget grep
    
    # 下载 wgcf 注册工具
    wget -qO wgcf https://github.com/ViRb3/wgcf/releases/download/v2.2.22/wgcf_2.2.22_linux_amd64
    chmod +x wgcf
    
    echo "==> [WARP] 正在向 Cloudflare 注册设备并接受 TOS..."
    if ! ./wgcf register --accept-tos; then
        echo "==> [ERROR] WARP 注册失败！请稍后重启容器重试。"
        exit 1
    fi
    
    echo "==> [WARP] 正在生成 WireGuard 配置文件..."
    ./wgcf generate
    
    # 从生成的 conf 中提取核心参数
    PRIVATE_KEY=$(grep '^PrivateKey' wgcf-profile.conf | cut -d'=' -f2 | tr -d ' ')
    V4=$(grep '^Address' wgcf-profile.conf | grep '\.' | cut -d'=' -f2 | tr -d ' ')
    V6=$(grep '^Address' wgcf-profile.conf | grep ':' | cut -d'=' -f2 | tr -d ' ')
    PEER_PUB=$(grep '^PublicKey' wgcf-profile.conf | cut -d'=' -f2 | tr -d ' ')
    
    # 动态组装 sing-box 的 config.json
    cat > config.json <<EOF
{
  "log": { "level": "warn" },
  "inbounds":[
    {
      "type": "socks",
      "tag": "socks-in",
      "listen": "0.0.0.0",
      "listen_port": 1080
    }
  ],
  "outbounds":[
    {
      "type": "wireguard",
      "tag": "warp-out",
      "server": "engage.cloudflareclient.com",
      "server_port": 2408,
      "local_address":[ "$V4", "$V6" ],
      "private_key": "$PRIVATE_KEY",
      "peer_public_key": "$PEER_PUB",
      "mtu": 1280
    }
  ]
}
EOF
    echo "==> [WARP] WARP 节点洗白配置生成成功！"
    # 清理安装包
    rm -f wgcf wgcf-account.toml wgcf-profile.conf
else
    echo "==> [WARP] 已检测到现有配置，跳过注册步骤。"
fi

echo "==> [WARP] 正在启动 sing-box SOCKS5 代理引擎..."
# 接管主进程，启动代理
exec sing-box run -c config.json
