package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(requests int, windowSec int, burst int) *RateLimiter {
	r := rate.Limit(float64(requests) / float64(windowSec))
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

// GetLimiter 获取访问者的限流器
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// Cleanup 清理过期的限流器
func (rl *RateLimiter) Cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			// 如果限流器在过去1分钟内没有被使用，删除它
			if limiter.Allow() {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	limiter *RateLimiter
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(requests int, windowSec int, burst int) *RateLimitMiddleware {
	limiter := NewRateLimiter(requests, windowSec, burst)
	go limiter.Cleanup() // 启动清理协程
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

// Handle 处理限流
func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端IP
		ip := getClientIP(r)

		// 获取该IP的限流器
		limiter := m.limiter.GetLimiter(ip)

		// 检查是否允许请求
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", string(rune(m.limiter.rate)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":429,"message":"too many requests"}`))
			return
		}

		// 设置限流响应头
		w.Header().Set("X-RateLimit-Limit", string(rune(m.limiter.rate)))
		w.Header().Set("X-RateLimit-Remaining", string(rune(limiter.Tokens())))

		next(w, r)
	}
}

// getClientIP 获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// 尝试从 X-Real-IP 获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	return r.RemoteAddr
}
