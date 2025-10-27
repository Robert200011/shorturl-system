package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"redirect-service/internal/handler"
	"redirect-service/internal/model"
	"redirect-service/internal/repo"
	"redirect-service/internal/service"
)

// 配置信息
const (
	redisAddr    = "localhost:6379"
	mysqlDSN     = "root:122722@tcp(localhost:3306)/shorturl?charset=utf8mb4&parseTime=True&loc=Local"
	shortenerURL = "http://localhost:8001"
	serverPort   = ":8002"
)

type RedirectService struct {
	redisClient  *redis.Client
	visitRepo    repo.VisitLogRepo
	shortenerURL string
}

type ShortLink struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	Status      int8       `json:"status"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
}

func main() {
	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	// 测试连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✓ Connected to Redis")

	// 初始化数据库Repository
	visitRepo, err := repo.NewVisitLogRepo(mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to init visit log repo: %v", err)
	}
	log.Println("✓ Connected to MySQL")

	// 创建服务实例
	svc := &RedirectService{
		redisClient:  redisClient,
		visitRepo:    visitRepo,
		shortenerURL: shortenerURL,
	}

	// 创建统计处理器
	statsHandler := handler.NewStatsHandler(visitRepo)

	// 注册路由
	http.HandleFunc("/api/stats/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[len("/api/stats/"):] != "" {
			if len(r.URL.Path) > len("/api/stats/") &&
				r.URL.Path[len(r.URL.Path)-5:] == "/logs" {
				statsHandler.GetRecentLogs(w, r)
			} else {
				statsHandler.GetStats(w, r)
			}
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})
	http.HandleFunc("/", svc.handleRedirect)

	log.Printf("🚀 Redirect service starting on %s...\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func (s *RedirectService) handleRedirect(w http.ResponseWriter, r *http.Request) {
	// 提取短链码
	shortCode := r.URL.Path[1:]
	if shortCode == "" || shortCode == "api" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 先从Redis缓存查询
	originalURL, err := s.getFromCache(ctx, shortCode)
	if err == nil && originalURL != "" {
		// 异步记录访问日志
		go s.logVisit(shortCode, r)
		http.Redirect(w, r, originalURL, http.StatusFound)
		return
	}

	// 缓存未命中，调用shortener服务API
	originalURL, err = s.getFromAPI(ctx, shortCode)
	if err != nil {
		log.Printf("Failed to get original URL: %v", err)
		http.Error(w, "Short link not found", http.StatusNotFound)
		return
	}

	// 异步记录访问日志
	go s.logVisit(shortCode, r)

	// 重定向
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (s *RedirectService) getFromCache(ctx context.Context, code string) (string, error) {
	key := "short:code:" + code
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return "", err
	}

	var link ShortLink
	if err := json.Unmarshal(data, &link); err != nil {
		return "", err
	}

	if link.Status != 1 {
		return "", fmt.Errorf("link is inactive")
	}

	if link.ExpireAt != nil && time.Now().After(*link.ExpireAt) {
		return "", fmt.Errorf("link is expired")
	}

	return link.OriginalURL, nil
}

func (s *RedirectService) getFromAPI(ctx context.Context, code string) (string, error) {
	url := fmt.Sprintf("%s/api/links/%s", s.shortenerURL, code)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			OriginalURL string `json:"original_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("API returned error code: %d", result.Code)
	}

	return result.Data.OriginalURL, nil
}

func (s *RedirectService) logVisit(shortCode string, r *http.Request) {
	// 解析访问信息
	visitInfo := service.ParseRequest(r)

	// 创建访问日志
	log := &model.VisitLog{
		ShortCode:  shortCode,
		IP:         visitInfo.IP,
		UserAgent:  visitInfo.UserAgent,
		Referer:    visitInfo.Referer,
		DeviceType: visitInfo.DeviceType,
		Browser:    visitInfo.Browser,
		OS:         visitInfo.OS,
		VisitedAt:  time.Now(),
	}

	// 保存到数据库
	ctx := context.Background()
	if err := s.visitRepo.Create(ctx, log); err != nil {
		fmt.Printf("Failed to save visit log: %v\n", err)
	}

	// 增加Redis中的访问计数
	countKey := "visit:count:" + shortCode
	s.redisClient.Incr(ctx, countKey)
}
