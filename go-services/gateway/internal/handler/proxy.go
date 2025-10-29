package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gateway/internal/config"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	services config.ServicesConfig
	client   *http.Client
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(services config.ServicesConfig) *ProxyHandler {
	return &ProxyHandler{
		services: services,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProxyToShortener 代理到短链生成服务
func (h *ProxyHandler) ProxyToShortener(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, h.services.Shortener.Host, "/api", time.Duration(h.services.Shortener.Timeout)*time.Millisecond)
}

// ProxyToRedirect 代理到重定向服务
func (h *ProxyHandler) ProxyToRedirect(w http.ResponseWriter, r *http.Request) {
	// 重定向服务的路径处理
	// /r/:code -> /:code
	path := strings.TrimPrefix(r.URL.Path, "/r")
	h.proxyWithPath(w, r, h.services.Redirect.Host, path, time.Duration(h.services.Redirect.Timeout)*time.Millisecond)
}

// proxy 通用代理方法
func (h *ProxyHandler) proxy(w http.ResponseWriter, r *http.Request, targetHost, pathPrefix string, timeout time.Duration) {
	// 构建目标URL
	targetURL := targetHost + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	h.proxyRequest(w, r, targetURL, timeout)
}

// proxyWithPath 使用自定义路径的代理方法
func (h *ProxyHandler) proxyWithPath(w http.ResponseWriter, r *http.Request, targetHost, path string, timeout time.Duration) {
	// 构建目标URL
	targetURL := targetHost + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	h.proxyRequest(w, r, targetURL, timeout)
}

// proxyRequest 执行代理请求
func (h *ProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string, timeout time.Duration) {
	// 创建带超时的Context
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// 创建新请求
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// 设置Host头
	proxyReq.Host = r.Host

	// 发送请求
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		}
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}
}

// HealthCheck 健康检查
func (h *ProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"gateway","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}
