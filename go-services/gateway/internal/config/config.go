package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth      AuthConfig
	RateLimit RateLimitConfig
	Services  ServicesConfig
	JWT       JWTConfig
}

type AuthConfig struct {
	Enabled      bool     `json:",default=true"`
	ExcludePaths []string // 不需要鉴权的路径
	AdminPaths   []string // 需要管理员权限的路径
}

type RateLimitConfig struct {
	Enabled   bool `json:",default=true"`
	Requests  int  `json:",default=100"` // 每个时间窗口的请求数
	WindowSec int  `json:",default=60"`  // 时间窗口（秒）
	BurstSize int  `json:",default=10"`  // 突发请求数
}

type ServicesConfig struct {
	Shortener ServiceEndpoint
	Redirect  ServiceEndpoint
}

type ServiceEndpoint struct {
	Host    string
	Timeout int `json:",default=5000"` // 超时时间（毫秒）
}

type JWTConfig struct {
	Secret     string
	ExpireHour int `json:",default=168"` // 7天
}
