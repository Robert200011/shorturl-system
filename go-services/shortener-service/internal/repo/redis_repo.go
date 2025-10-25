package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"shortener-service/internal/model"
)

const (
	// 短链码 -> 原始URL映射的key前缀
	shortCodePrefix = "short:code:"
	// 原始URL -> 短链码映射的key前缀
	originalURLPrefix = "short:url:"
)

// RedisRepo Redis缓存操作接口
type RedisRepo interface {
	SetShortLink(ctx context.Context, link *model.ShortLink, ttl time.Duration) error
	GetShortLink(ctx context.Context, code string) (*model.ShortLink, error)
	GetShortCodeByURL(ctx context.Context, url string) (string, error)
	DeleteShortLink(ctx context.Context, code string) error
	Exists(ctx context.Context, code string) (bool, error)
}

// redisRepo Redis缓存操作实现
type redisRepo struct {
	client *redis.Client
}

// NewRedisRepo 创建Redis缓存操作实例
func NewRedisRepo(host, password string, db int) (RedisRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return &redisRepo{client: client}, nil
}

// SetShortLink 缓存短链接信息
func (r *redisRepo) SetShortLink(ctx context.Context, link *model.ShortLink, ttl time.Duration) error {
	data, err := json.Marshal(link)
	if err != nil {
		return err
	}

	// 缓存短链码 -> 完整信息
	codeKey := shortCodePrefix + link.ShortCode
	if err := r.client.Set(ctx, codeKey, data, ttl).Err(); err != nil {
		return err
	}

	// 缓存原始URL -> 短链码
	urlKey := originalURLPrefix + link.OriginalURL
	return r.client.Set(ctx, urlKey, link.ShortCode, ttl).Err()
}

// GetShortLink 从缓存获取短链接信息
func (r *redisRepo) GetShortLink(ctx context.Context, code string) (*model.ShortLink, error) {
	key := shortCodePrefix + code
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var link model.ShortLink
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

// GetShortCodeByURL 根据原始URL获取短链码
func (r *redisRepo) GetShortCodeByURL(ctx context.Context, url string) (string, error) {
	key := originalURLPrefix + url
	code, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 缓存未命中
		}
		return "", err
	}
	return code, nil
}

// DeleteShortLink 删除短链接缓存
func (r *redisRepo) DeleteShortLink(ctx context.Context, code string) error {
	key := shortCodePrefix + code
	return r.client.Del(ctx, key).Err()
}

// Exists 检查短链码是否存在
func (r *redisRepo) Exists(ctx context.Context, code string) (bool, error) {
	key := shortCodePrefix + code
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
