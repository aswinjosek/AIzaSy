package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type bytesPool struct {
	pool sync.Pool
}

func (p *bytesPool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *bytesPool) Put(b[]byte) {
	p.pool.Put(b)
}

func main() {
	targetURL, _ := url.Parse("https://generativelanguage.googleapis.com")

	// 极简且强悍的原生 HTTP Transport
	// 流量会自动被操作系统的 wg0 路由表拦截并送往 Cloudflare 骨干网
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
	
	reverseProxy.BufferPool = &bytesPool{
		pool: sync.Pool{
			New: func() interface{} { return make([]byte, 32*1024) },
		},
	}
	
	// 零延迟流式输出
	reverseProxy.FlushInterval = -1 

	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		
		// 剥离追踪 IP 头
		req.Header["X-Forwarded-For"] = nil
		req.Header.Del("X-Real-IP")
		req.Header.Del("True-Client-IP")
		req.Header.Del("CF-Connecting-IP")
		
		logRequest(req)
	}

	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		return nil
	}

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			reverseProxy.ServeHTTP(w, r)
		}),
	}

	log.Println("⚡ AIzaSy Kernel-Level Gateway is running on :8080 (Routed via wg0)...")
	log.Fatal(server.ListenAndServe())
}

func logRequest(req *http.Request) {
	requestURI := req.URL.RequestURI()
	re := regexp.MustCompile(`([?&]key=)(AIzaSy[A-Za-z0-9_-]{4})([A-Za-z0-9_-]+)`)
	maskedURI := re.ReplaceAllString(requestURI, "${1}${2}***")
	log.Printf("[FORWARD] %s %s", req.Method, maskedURI)
}
