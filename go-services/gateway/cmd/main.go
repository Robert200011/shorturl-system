package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/internal/service"
)

var configFile = flag.String("f", "internal/config/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建JWT服务
	jwtService := service.NewJWTService(c.JWT.Secret, c.JWT.ExpireHour)

	// 创建中间件
	loggerMw := middleware.NewLoggerMiddleware()
	var rateLimitMw *middleware.RateLimitMiddleware
	var authMw *middleware.AuthMiddleware

	if c.RateLimit.Enabled {
		rateLimitMw = middleware.NewRateLimitMiddleware(
			c.RateLimit.Requests,
			c.RateLimit.WindowSec,
			c.RateLimit.BurstSize,
		)
		log.Println("✓ Rate limiting enabled")
	}

	if c.Auth.Enabled {
		authMw = middleware.NewAuthMiddleware(
			jwtService,
			c.Auth.ExcludePaths,
			c.Auth.AdminPaths,
		)
		log.Println("✓ Authentication enabled")
	}

	// 创建代理处理器
	proxyHandler := handler.NewProxyHandler(c.Services)

	// 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册路由
	registerRoutes(server, proxyHandler, jwtService, loggerMw, rateLimitMw, authMw, &c)

	fmt.Printf("🚀 Gateway starting at %s:%d...\n", c.Host, c.Port)
	fmt.Printf("📝 Config loaded:\n")
	fmt.Printf("   - Shortener Service: %s\n", c.Services.Shortener.Host)
	fmt.Printf("   - Redirect Service: %s\n", c.Services.Redirect.Host)
	fmt.Printf("   - Rate Limit: %v (requests: %d/%ds)\n", c.RateLimit.Enabled, c.RateLimit.Requests, c.RateLimit.WindowSec)
	fmt.Printf("   - Authentication: %v\n", c.Auth.Enabled)

	server.Start()
}

// registerRoutes 注册路由
func registerRoutes(
	server *rest.Server,
	proxyHandler *handler.ProxyHandler,
	jwtService *service.JWTService,
	loggerMw *middleware.LoggerMiddleware,
	rateLimitMw *middleware.RateLimitMiddleware,
	authMw *middleware.AuthMiddleware,
	c *config.Config,
) {
	// 应用全局中间件链
	applyMiddleware := func(h http.HandlerFunc) http.HandlerFunc {
		// 先应用日志中间件
		wrapped := loggerMw.Handle(h)

		// 然后应用限流中间件
		if rateLimitMw != nil {
			wrapped = rateLimitMw.Handle(wrapped)
		}

		// 最后应用鉴权中间件
		if authMw != nil {
			wrapped = authMw.Handle(wrapped)
		}

		return wrapped
	}

	// 健康检查（不需要中间件）
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/health",
		Handler: proxyHandler.HealthCheck,
	})

	// 认证相关路由（示例）
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/auth/login",
		Handler: authLoginHandler(jwtService),
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/auth/refresh",
		Handler: authRefreshHandler(jwtService),
	})

	// 短链生成服务代理（需要鉴权和限流）
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/shorten",
		Handler: applyMiddleware(proxyHandler.ProxyToShortener),
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/links/:code",
		Handler: applyMiddleware(proxyHandler.ProxyToShortener),
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/batch/shorten",
		Handler: applyMiddleware(proxyHandler.ProxyToShortener),
	})

	// 统计服务代理
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/stats/:code",
		Handler: applyMiddleware(proxyHandler.ProxyToRedirect),
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/stats/:code/logs",
		Handler: applyMiddleware(proxyHandler.ProxyToRedirect),
	})

	// 重定向服务代理（不需要鉴权）
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/r/:code",
		Handler: loggerMw.Handle(proxyHandler.ProxyToRedirect),
	})
}

// authLoginHandler 登录处理器（示例）
func authLoginHandler(jwtService *service.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 这里应该验证用户名密码，这只是示例
		// 实际项目中应该从数据库验证用户信息

		// 示例：生成测试Token
		token, err := jwtService.GenerateToken(1, "testuser", "user")
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"code":0,"message":"success","data":{"token":"%s"}}`, token)
	}
}

// authRefreshHandler 刷新Token处理器
func authRefreshHandler(jwtService *service.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从Header获取旧Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// 解析Bearer Token
		parts := make([]string, 0)
		for _, part := range []rune(authHeader) {
			if part == ' ' {
				break
			}
			parts = append(parts, string(part))
		}

		if len(parts) < 2 {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		oldToken := authHeader[len("Bearer "):]

		// 刷新Token
		newToken, err := jwtService.RefreshToken(oldToken)
		if err != nil {
			http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"code":0,"message":"success","data":{"token":"%s"}}`, newToken)
	}
}
