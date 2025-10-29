package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/config"
	"gateway/internal/service"
	"gateway/internal/types"
)

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	limiter *service.RedisRateLimiter
	config  config.RateLimitConfig
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(limiter *service.RedisRateLimiter, cfg config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		config:  cfg,
	}
}

// Handle 处理限流
func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. 全局限流检查
		if m.config.GlobalLimit > 0 {
			allowed, err := m.limiter.Allow(ctx, "global", m.config.GlobalLimit)
			if err != nil {
				m.errorResponse(w, "rate limit check failed")
				return
			}
			if !allowed {
				m.rateLimitResponse(w, "global rate limit exceeded")
				return
			}
		}

		// 2. IP限流检查
		if m.config.IPLimit > 0 {
			ip := m.extractIP(r)
			ipKey := fmt.Sprintf("ip:%s", ip)

			allowed, err := m.limiter.Allow(ctx, ipKey, m.config.IPLimit)
			if err != nil {
				m.errorResponse(w, "rate limit check failed")
				return
			}
			if !allowed {
				m.rateLimitResponse(w, "IP rate limit exceeded")
				return
			}
		}

		// 3. 用户限流检查 (如果已认证)
		if m.config.UserLimit > 0 {
			userID := GetUserID(ctx)
			if userID > 0 {
				userKey := fmt.Sprintf("user:%d", userID)

				allowed, err := m.limiter.Allow(ctx, userKey, m.config.UserLimit)
				if err != nil {
					m.errorResponse(w, "rate limit check failed")
					return
				}
				if !allowed {
					m.rateLimitResponse(w, "user rate limit exceeded")
					return
				}
			}
		}

		// 添加限流信息到响应头
		m.setRateLimitHeaders(w, r)

		next(w, r)
	}
}

// extractIP 提取客户端IP
func (m *RateLimitMiddleware) extractIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-IP 获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用 RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// setRateLimitHeaders 设置限流响应头
func (m *RateLimitMiddleware) setRateLimitHeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// IP限流剩余配额
	if m.config.IPLimit > 0 {
		ip := m.extractIP(r)
		ipKey := fmt.Sprintf("ip:%s", ip)
		remaining, _ := m.limiter.GetRemaining(ctx, ipKey, m.config.IPLimit)
		w.Header().Set("X-RateLimit-IP-Limit", fmt.Sprintf("%d", m.config.IPLimit))
		w.Header().Set("X-RateLimit-IP-Remaining", fmt.Sprintf("%d", remaining))
	}

	// 用户限流剩余配额
	userID := GetUserID(ctx)
	if m.config.UserLimit > 0 && userID > 0 {
		userKey := fmt.Sprintf("user:%d", userID)
		remaining, _ := m.limiter.GetRemaining(ctx, userKey, m.config.UserLimit)
		w.Header().Set("X-RateLimit-User-Limit", fmt.Sprintf("%d", m.config.UserLimit))
		w.Header().Set("X-RateLimit-User-Remaining", fmt.Sprintf("%d", remaining))
	}
}

// rateLimitResponse 返回限流响应
func (m *RateLimitMiddleware) rateLimitResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := types.ErrorResponse{
		Code:    http.StatusTooManyRequests,
		Message: message,
	}

	w.Write([]byte(fmt.Sprintf(`{"code":%d,"message":"%s"}`, response.Code, response.Message)))
}

// errorResponse 返回错误响应
func (m *RateLimitMiddleware) errorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := types.ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: message,
	}

	w.Write([]byte(fmt.Sprintf(`{"code":%d,"message":"%s"}`, response.Code, response.Message)))
}
