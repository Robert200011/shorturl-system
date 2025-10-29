package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int) (bool, error)
	AllowN(ctx context.Context, key string, limit int, n int) (bool, error)
}

// RedisRateLimiter 基于Redis的限流器 (滑动窗口算法)
type RedisRateLimiter struct {
	client *redis.Client
	window time.Duration // 时间窗口
}

// NewRedisRateLimiter 创建Redis限流器
func NewRedisRateLimiter(client *redis.Client, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		window: window,
	}
}

// Allow 检查是否允许单个请求
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	return r.AllowN(ctx, key, limit, 1)
}

// AllowN 检查是否允许N个请求 (滑动窗口算法)
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, limit int, n int) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-r.window)

	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

	// Lua脚本实现原子操作
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local count = tonumber(ARGV[4])
		
		-- 移除过期的记录
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- 获取当前窗口内的请求数
		local current = redis.call('ZCARD', key)
		
		-- 检查是否超限
		if current + count > limit then
			return 0
		end
		
		-- 添加新的请求记录
		for i = 1, count do
			redis.call('ZADD', key, now, now .. ':' .. i)
		end
		
		-- 设置过期时间
		redis.call('EXPIRE', key, 60)
		
		return 1
	`)

	result, err := script.Run(
		ctx,
		r.client,
		[]string{rateLimitKey},
		now.UnixNano(),
		windowStart.UnixNano(),
		limit,
		n,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// GetRemaining 获取剩余配额
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string, limit int) (int, error) {
	now := time.Now()
	windowStart := now.Add(-r.window)

	rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

	// 清理过期数据
	if err := r.client.ZRemRangeByScore(ctx, rateLimitKey, "0", fmt.Sprintf("%d", windowStart.UnixNano())).Err(); err != nil {
		return 0, err
	}

	// 获取当前数量
	current, err := r.client.ZCard(ctx, rateLimitKey).Result()
	if err != nil {
		return 0, err
	}

	remaining := limit - int(current)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
