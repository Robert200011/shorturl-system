package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/conf"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/internal/router"
	"gateway/internal/service"
)

var configFile = flag.String("f", "internal/config/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	fmt.Println("=================================================")
	fmt.Println("🚀 Gateway Service Starting...")
	fmt.Println("=================================================")

	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
		DB:       c.Redis.DB,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	fmt.Println("✅ Connected to Redis")

	// 初始化JWT服务
	jwtService := service.NewJWTService(c.JWT.Secret, c.JWT.ExpireTime)
	fmt.Println("✅ JWT Service initialized")

	// 初始化限流器 (1分钟窗口)
	rateLimiter := service.NewRedisRateLimiter(redisClient, time.Minute)
	fmt.Println("✅ Rate Limiter initialized")

	// 初始化代理处理器
	proxyHandler, err := handler.NewProxyHandler(c.Upstream)
	if err != nil {
		log.Fatalf("❌ Failed to create proxy handler: %v", err)
	}
	fmt.Println("✅ Proxy Handler initialized")

	// 初始化中间件
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter, c.RateLimit)
	loggerMiddleware := middleware.NewLoggerMiddleware()
	fmt.Println("✅ Middlewares initialized")

	// 创建路由器
	routerInstance := router.NewRouter(
		proxyHandler,
		authMiddleware,
		rateLimitMiddleware,
		loggerMiddleware,
	)

	// 注册路由
	mux := http.NewServeMux()
	routerInstance.RegisterRoutes(mux)
	fmt.Println("✅ Routes registered")

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Println("=================================================")
	fmt.Printf("🌐 Gateway listening on: %s\n", addr)
	fmt.Println("=================================================")
	fmt.Println()
	fmt.Println("📋 Route Table:")
	fmt.Println("  ✓ GET  /health                  - Health check")
	fmt.Println("  ✓ POST /api/shorten             - Create short link (Auth)")
	fmt.Println("  ✓ GET  /api/links/:code         - Get link details (Auth)")
	fmt.Println("  ✓ POST /api/batch/shorten       - Batch create (Auth)")
	fmt.Println("  ✓ GET  /api/stats/:code         - Get stats (Auth)")
	fmt.Println("  ✓ GET  /api/stats/:code/logs    - Get logs (Auth)")
	fmt.Println("  ✓ GET  /:code                   - Redirect (Public)")
	fmt.Println()
	fmt.Println("⚙️  Features:")
	fmt.Println("  • JWT Authentication")
	fmt.Println("  • Rate Limiting (Global/IP/User)")
	fmt.Println("  • Request Logging")
	fmt.Println("  • Reverse Proxy")
	fmt.Println()
	fmt.Println("🔒 Rate Limits:")
	fmt.Printf("  • Global: %d req/min\n", c.RateLimit.GlobalLimit)
	fmt.Printf("  • IP:     %d req/min\n", c.RateLimit.IPLimit)
	fmt.Printf("  • User:   %d req/min\n", c.RateLimit.UserLimit)
	fmt.Println("=================================================")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
