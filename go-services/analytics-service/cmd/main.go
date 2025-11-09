package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zeromicro/go-zero/core/conf"

	"analytics-service/internal/config"
	"analytics-service/internal/consumer"
	"analytics-service/internal/handler"
	"analytics-service/internal/repo"
	"analytics-service/internal/service"
)

var configFile = flag.String("f", "internal/config/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	fmt.Println("=================================================")
	fmt.Println("🚀 Analytics Service Starting...")
	fmt.Println("=================================================")

	// 初始化数据库Repository
	analyticsRepo, err := repo.NewAnalyticsRepo(c.Mysql.DataSource)
	if err != nil {
		log.Fatalf("❌ Failed to init analytics repo: %v", err)
	}
	fmt.Println("✅ Connected to MySQL and migrated tables")

	// 初始化聚合器
	aggregator := service.NewAggregator(analyticsRepo)
	fmt.Println("✅ Aggregator initialized")

	// 初始化Kafka消费者
	kafkaConsumer, err := consumer.NewKafkaConsumer(
		c.Kafka.Brokers,
		c.Kafka.GroupID,
		c.Kafka.Topic,
		aggregator,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// 启动Kafka消费者（在goroutine中）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		fmt.Println("🔄 Starting Kafka consumer...")
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("❌ Kafka consumer error: %v", err)
		}
	}()

	// 初始化HTTP处理器
	analyticsHandler := handler.NewAnalyticsHandler(analyticsRepo)

	// 注册HTTP路由
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"analytics"}`))
	})

	// 统计API
	mux.HandleFunc("/api/analytics/daily/", analyticsHandler.GetDailyStats)
	mux.HandleFunc("/api/analytics/hourly/", analyticsHandler.GetHourlyStats)
	mux.HandleFunc("/api/analytics/browser/", analyticsHandler.GetBrowserStats)
	mux.HandleFunc("/api/analytics/device/", analyticsHandler.GetDeviceStats)
	mux.HandleFunc("/api/analytics/os/", analyticsHandler.GetOSStats)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Println("=================================================")
	fmt.Printf("🌐 HTTP server listening on: %s\n", addr)
	fmt.Println("=================================================")
	fmt.Println()
	fmt.Println("📋 API Endpoints:")
	fmt.Println("  ✓ GET  /health                           - Health check")
	fmt.Println("  ✓ GET  /api/analytics/daily/:code        - Daily stats")
	fmt.Println("  ✓ GET  /api/analytics/hourly/:code       - Hourly stats")
	fmt.Println("  ✓ GET  /api/analytics/browser/:code      - Browser stats")
	fmt.Println("  ✓ GET  /api/analytics/device/:code       - Device stats")
	fmt.Println("  ✓ GET  /api/analytics/os/:code           - OS stats")
	fmt.Println()
	fmt.Println("📊 Kafka Consumer:")
	fmt.Printf("  • Brokers: %v\n", c.Kafka.Brokers)
	fmt.Printf("  • Topic:   %s\n", c.Kafka.Topic)
	fmt.Printf("  • GroupID: %s\n", c.Kafka.GroupID)
	fmt.Println("=================================================")

	// 启动HTTP服务器（在goroutine中）
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down gracefully...")

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("⚠️  HTTP server shutdown error: %v", err)
	}

	fmt.Println("✅ Service stopped")
}
