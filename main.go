package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 内存复用池：实现 Zero-Allocation，彻底拯救高并发下的 GC 压力
type bytesPool struct {
	pool sync.Pool
}

func (p *bytesPool) Get()[]byte {
	return p.pool.Get().([]byte)
}

func (p *bytesPool) Put(b[]byte) {
	p.pool.Put(b)
}

// 优雅的 API 落地页 (支持 Twitter/X.com 预览卡片抓取)
const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AIzaSy - Gemini API Gateway</title>
    <meta name="description" content="A lightweight, open-source Gemini API gateway. No keys logged.">
    <meta property="og:title" content="AIzaSy - Gemini API Gateway">
    <meta property="og:description" content="A lightweight, open-source proxy for Google Gemini API.">
    <meta name="twitter:card" content="summary">
    <meta name="twitter:creator" content="@ccbkkb">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            color: #24292f;
            background-color: #ffffff;
            line-height: 1.5;
            max-width: 720px;
            margin: 0 auto;
            padding: 50px 20px;
        }
        h1 { font-size: 24px; border-bottom: 1px solid #eaecef; padding-bottom: 10px; margin-bottom: 20px; }
        h2 { font-size: 18px; margin-top: 35px; border-bottom: 1px solid #eaecef; padding-bottom: 5px; }
        p { font-size: 14px; color: #57606a; margin-top: 0; }
        a { color: #0969da; text-decoration: none; }
        a:hover { text-decoration: underline; }
        code, pre { 
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace; 
            background-color: #f6f8fa; 
            border-radius: 6px; 
        }
        code { padding: 0.2em 0.4em; font-size: 13px; }
        pre { padding: 16px; overflow: auto; font-size: 13px; line-height: 1.45; border: 1px solid #d0d7de; }
        .status { color: #1a7f37; font-size: 14px; font-weight: normal; float: right; margin-top: 8px; }
    </style>
</head>
<body>
    <h1>
        AIzaSy Gateway
        <span class="status">[🟢 system.status: ok]</span>
    </h1>
    
    <p>A lightweight, high-performance proxy for the Google Gemini API.</p>

    <h2>Privacy & Source</h2>
    <p>This gateway acts as a dumb pipe. It does <b>not</b> log, store, or inspect your API keys. Log outputs are heavily masked.</p>
    <p>
        Source Code: <a href="https://github.com/ccbkkb/AIzaSy" target="_blank">github.com/ccbkkb/AIzaSy</a><br>
        Maintainer: <a href="https://github.com/ccbkkb" target="_blank">@ccbkkb</a>
    </p>

    <h2>Usage</h2>
    <p>Just replace <code>generativelanguage.googleapis.com</code> with <code>aizasy.com</code>.</p>
    
    <pre><code>curl -H 'Content-Type: application/json' \
     -X POST 'https://aizasy.com/v1beta/models/gemini-1.5-pro:generateContent?key=YOUR_API_KEY' \
     -d '{
       "contents": [{"parts":[{"text": "Explain quantum computing in one sentence."}]}]
     }'</code></pre>
</body>
</html>
`

func main() {
	targetURL, _ := url.Parse("https://generativelanguage.googleapis.com")

	// 读取跨域配置：默认完全开放 (*)，支持随时通过环境变量收紧权限
	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	// 极致优化的 HTTP Transport (连接池、Keep-Alive 彻底拉满)
	optimizedTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			ClientSessionCache: tls.NewLRUClientSessionCache(1000),
		},
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.Transport = optimizedTransport

	// 挂载复用内存池 (复用 32KB 缓冲区)
	reverseProxy.BufferPool = &bytesPool{
		pool: sync.Pool{
			New: func() interface{} { return make([]byte, 32*1024) },
		},
	}

	// 核心：负数代表立即 Flush，彻底消除 AI 打字机流式输出的滞留卡顿
	reverseProxy.FlushInterval = -1

	// 请求拦截器：伪装 Host，抹除客户端溯源 IP，打印脱敏日志
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// 伪装请求 Host 为官方端点
		req.Host = targetURL.Host

		// 彻底剥离可能暴露客户真实 IP 的请求头 (防止被溯源和风控)
		req.Header["X-Forwarded-For"] = nil
		req.Header.Del("X-Real-IP")
		req.Header.Del("True-Client-IP")
		req.Header.Del("CF-Connecting-IP")

		// 记录绝对隐私的安全日志
		logRequest(req)
	}

	// 响应拦截器：注入动态跨域头
	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("Access-Control-Allow-Origin", corsOrigin)
		return nil
	}

	// 配置安全的 HTTP 服务器网关
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. 处理 CORS 跨域预检请求
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 2. 根路径拦截：返回优雅的开源说明落地页
			if r.URL.Path == "/" && r.Method == "GET" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, indexHTML)
				return
			}

			// 3. 路径白名单拦截：防范恶意爬虫和目录扫描工具浪费带宽
			// 仅放行带有 /v1beta/ 或 /v1/ 的标准 API 路径
			if !strings.HasPrefix(r.URL.Path, "/v1beta/") && !strings.HasPrefix(r.URL.Path, "/v1/") {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error": "Forbidden: Please use valid API endpoints. Powered by AIzaSy."}`)
				return
			}

			// 4. 正常转发纯数据流
			reverseProxy.ServeHTTP(w, r)
		}),
	}

	log.Printf("⚡ AIzaSy Kernel-Level Gateway is running on :8080 (CORS: %s)...", corsOrigin)
	log.Fatal(server.ListenAndServe())
}

// 核心公关级安全功能：绝对隐私日志脱敏 (Zero-Knowledge Logging)
func logRequest(req *http.Request) {
	requestURI := req.URL.RequestURI()
	
	// 无差别全域脱敏：匹配 URL 中 ?key= 或 &key= 后面的所有内容，直到遇到下一个参数 & 或结束
	re := regexp.MustCompile(`([?&]key=)[^&]+`)
	
	// 二次处理：只保留 ?key= ，后面的值一根毛都不留，全部替换为 ***
	maskedURI := re.ReplaceAllString(requestURI, "${1}***")
	
	log.Printf("[FORWARD] %s %s", req.Method, maskedURI)
}
