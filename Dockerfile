# 阶段 1：编译 Go 程序
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init aizasy && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway main.go

# 阶段 2：运行环境 (Alpine + WireGuard 内核控制工具)
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata wireguard-tools iptables iproute2 wget curl
WORKDIR /app

# 拷贝二进制文件和启动脚本
COPY --from=builder /app/gateway .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080
CMD ["./entrypoint.sh"]
