package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/internal/config"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	shortenerProxy *httputil.ReverseProxy
	redirectProxy  *httputil.ReverseProxy
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(cfg config.UpstreamConfig) (*ProxyHandler, error) {
	// 创建 shortener-service 代理
	shortenerURL, err := url.Parse(cfg.ShortenerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid shortener URL: %w", err)
	}
	shortenerProxy := httputil.NewSingleHostReverseProxy(shortenerURL)
	shortenerProxy.ErrorHandler = errorHandler

	// 创建 redirect-service 代理
	redirectURL, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URL: %w", err)
	}
	redirectProxy := httputil.NewSingleHostReverseProxy(redirectURL)
	redirectProxy.ErrorHandler = errorHandler

	return &ProxyHandler{
		shortenerProxy: shortenerProxy,
		redirectProxy:  redirectProxy,
	}, nil
}

// HandleShortener 处理短链生成服务请求
func (h *ProxyHandler) HandleShortener(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("→ Proxying to shortener-service: %s %s\n", r.Method, r.URL.Path)
	h.shortenerProxy.ServeHTTP(w, r)
}

// HandleRedirect 处理重定向服务请求
func (h *ProxyHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("→ Proxying to redirect-service: %s %s\n", r.Method, r.URL.Path)
	h.redirectProxy.ServeHTTP(w, r)
}

// HandleShortCode 处理短链重定向
func (h *ProxyHandler) HandleShortCode(w http.ResponseWriter, r *http.Request) {
	// 短链码在URL路径中，直接转发
	fmt.Printf("→ Redirect short code: %s\n", r.URL.Path)
	h.redirectProxy.ServeHTTP(w, r)
}

// errorHandler 代理错误处理器
func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("❌ Proxy error: %v\n", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)

	errorMsg := "upstream service unavailable"
	if strings.Contains(err.Error(), "connection refused") {
		errorMsg = "upstream service connection refused"
	}

	w.Write([]byte(fmt.Sprintf(`{"code":%d,"message":"%s"}`, http.StatusBadGateway, errorMsg)))
}
