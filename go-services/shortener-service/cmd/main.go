package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"shortener-service/internal/config"
	"shortener-service/internal/handler"
	"shortener-service/internal/repo"
	"shortener-service/internal/service"
)

var configFile = flag.String("f", "internal/config/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化数据库Repository
	dbRepo, err := repo.NewShortLinkRepo(c.Mysql.DataSource)
	if err != nil {
		log.Fatalf("Failed to init db repo: %v", err)
	}

	// 初始化Redis Repository
	redisRepo, err := repo.NewRedisRepo(c.Redis.Host, c.Redis.Pass, c.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to init redis repo: %v", err)
	}

	// 初始化ID生成器
	idGen, err := service.NewSnowflakeIDGen(c.Snowflake.MachineID)
	if err != nil {
		log.Fatalf("Failed to init id generator: %v", err)
	}

	// 初始化短链服务
	shortenerSvc := service.NewShortenerService(
		dbRepo,
		redisRepo,
		idGen,
		c.ShortUrl.Domain,
		c.ShortUrl.CacheTTL,
	)

	// 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册路由
	registerHandlers(server, shortenerSvc)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// registerHandlers 注册路由
func registerHandlers(server *rest.Server, svc service.ShortenerService) {
	// 短链生成处理器
	shortenHandler := handler.NewShortenHandler(svc)
	batchHandler := handler.NewBatchHandler(svc)

	// 路由组
	server.AddRoutes(
		[]rest.Route{
			// 创建短链接
			{
				Method:  "POST",
				Path:    "/api/shorten",
				Handler: shortenHandler.CreateShortLink,
			},
			// 获取短链接详情
			{
				Method:  "GET",
				Path:    "/api/links/:code",
				Handler: shortenHandler.GetShortLink,
			},
			// 批量创建短链接
			{
				Method:  "POST",
				Path:    "/api/batch/shorten",
				Handler: batchHandler.BatchCreateShortLinks,
			},
		},
	)
}
