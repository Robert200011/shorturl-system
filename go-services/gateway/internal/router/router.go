package router

import (
	"net/http"
	"strings"

	"gateway/internal/handler"
	"gateway/internal/middleware"
)

// Router 路由器
type Router struct {
	proxyHandler        *handler.ProxyHandler
	authMiddleware      *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	loggerMiddleware    *middleware.LoggerMiddleware
}

// NewRouter 创建路由器
func NewRouter(
	proxyHandler *handler.ProxyHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	loggerMiddleware *middleware.LoggerMiddleware,
) *Router {
	return &Router{
		proxyHandler:        proxyHandler,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		loggerMiddleware:    loggerMiddleware,
	}
}

// RegisterRoutes 注册路由
func (router *Router) RegisterRoutes(mux *http.ServeMux) {
	// 健康检查 (无需鉴权和限流)
	mux.HandleFunc("/health", router.handleHealth)

	// API路由 (需要鉴权和限流)
	mux.HandleFunc("/api/", router.chain(
		router.handleAPI,
		router.loggerMiddleware.Handle,
		router.rateLimitMiddleware.Handle,
		router.authMiddleware.Handle,
	))

	// 短链重定向 (无需鉴权，但需要限流)
	mux.HandleFunc("/", router.chain(
		router.handleShortCode,
		router.loggerMiddleware.Handle,
		router.rateLimitMiddleware.Handle,
	))
}

// handleAPI 处理API请求
func (router *Router) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 路由到 shortener-service
	if strings.HasPrefix(path, "/api/shorten") ||
		strings.HasPrefix(path, "/api/links/") ||
		strings.HasPrefix(path, "/api/batch/") {
		router.proxyHandler.HandleShortener(w, r)
		return
	}

	// 路由到 redirect-service (统计接口)
	if strings.HasPrefix(path, "/api/stats/") {
		router.proxyHandler.HandleRedirect(w, r)
		return
	}

	// 未知路由
	http.NotFound(w, r)
}

// handleShortCode 处理短链重定向
func (router *Router) handleShortCode(w http.ResponseWriter, r *http.Request) {
	// 排除特殊路径
	if r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/health") {
		http.NotFound(w, r)
		return
	}

	// 转发到 redirect-service
	router.proxyHandler.HandleShortCode(w, r)
}

// handleHealth 健康检查
func (router *Router) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"gateway"}`))
}

// chain 中间件链
func (router *Router) chain(endpoint http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	// 从右到左应用中间件
	for i := len(middlewares) - 1; i >= 0; i-- {
		endpoint = middlewares[i](endpoint)
	}
	return endpoint
}
