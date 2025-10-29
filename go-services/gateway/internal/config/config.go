package config

import "github.com/zeromicro/go-zero/rest"

// Config 网关配置
type Config struct {
	rest.RestConf

	// 上游服务配置
	Upstream UpstreamConfig

	// Redis配置
	Redis RedisConfig

	// JWT配置
	JWT JWTConfig

	// 限流配置
	RateLimit RateLimitConfig
}

// UpstreamConfig 上游服务配置
type UpstreamConfig struct {
	ShortenerURL string // 短链生成服务
	RedirectURL  string // 重定向服务
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host string
	Pass string
	DB   int
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string // JWT密钥
	ExpireTime int64  // 过期时间(秒)
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	GlobalLimit int // 全局限流 (请求/分钟)
	IPLimit     int // IP限流 (请求/分钟)
	UserLimit   int // 用户限流 (请求/分钟)
}
