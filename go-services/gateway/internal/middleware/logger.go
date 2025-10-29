package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct{}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

// responseWriter 包装ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Handle 处理日志记录
func (m *LoggerMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 执行下一个处理器
		next(wrapped, r)

		// 计算耗时
		duration := time.Since(start)

		// 获取用户信息
		userID := GetUserID(r.Context())
		username := GetUsername(r.Context())

		// 记录日志
		logEntry := fmt.Sprintf(
			"[%s] %s %s | Status: %d | Size: %d bytes | Duration: %v | IP: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			wrapped.size,
			duration,
			m.extractIP(r),
		)

		if userID > 0 {
			logEntry += fmt.Sprintf(" | User: %s (ID: %d)", username, userID)
		}

		// 打印日志
		if wrapped.statusCode >= 500 {
			fmt.Printf("❌ ERROR: %s\n", logEntry)
		} else if wrapped.statusCode >= 400 {
			fmt.Printf("⚠️  WARN:  %s\n", logEntry)
		} else {
			fmt.Printf("✅ INFO:  %s\n", logEntry)
		}
	}
}

// extractIP 提取客户端IP
func (m *LoggerMiddleware) extractIP(r *http.Request) string {
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
