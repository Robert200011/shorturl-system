package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedirectService struct {
	redisClient  *redis.Client
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
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 测试连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	svc := &RedirectService{
		redisClient:  redisClient,
		shortenerURL: "http://localhost:8001",
	}

	http.HandleFunc("/", svc.handleRedirect)

	fmt.Println("Redirect service starting on :8002...")
	log.Fatal(http.ListenAndServe(":8002", nil))
}

func (s *RedirectService) handleRedirect(w http.ResponseWriter, r *http.Request) {
	// 提取短链码
	shortCode := r.URL.Path[1:] // 去掉开头的 "/"
	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 先从Redis缓存查询
	originalURL, err := s.getFromCache(ctx, shortCode)
	if err == nil && originalURL != "" {
		// 记录访问日志（异步）
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

	// 记录访问日志（异步）
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

	// 检查状态和过期时间
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
	// 这里可以将访问日志发送到Kafka或直接写入数据库
	log.Printf("Visit: code=%s, ip=%s, ua=%s, referer=%s",
		shortCode,
		r.RemoteAddr,
		r.UserAgent(),
		r.Referer(),
	)

	// 增加Redis中的访问计数
	ctx := context.Background()
	countKey := "visit:count:" + shortCode
	s.redisClient.Incr(ctx, countKey)
}
