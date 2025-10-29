package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter 包装 ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct{}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

// Handle 处理日志记录
func (m *LoggerMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装 ResponseWriter
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 调用下一个处理器
		next(rw, r)

		// 记录请求信息
		duration := time.Since(start)
		log.Printf("[Gateway] %s %s %d %s %dB %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			rw.written,
			getClientIP(r),
		)
	}
}
