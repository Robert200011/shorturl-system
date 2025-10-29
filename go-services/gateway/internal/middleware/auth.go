package middleware

import (
	"context"
	"net/http"
	"strings"

	"gateway/internal/service"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	UserRoleKey contextKey = "user_role"
)

// AuthMiddleware JWT鉴权中间件
type AuthMiddleware struct {
	jwtService   *service.JWTService
	excludePaths []string
	adminPaths   []string
}

// NewAuthMiddleware 创建鉴权中间件
func NewAuthMiddleware(jwtService *service.JWTService, excludePaths, adminPaths []string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:   jwtService,
		excludePaths: excludePaths,
		adminPaths:   adminPaths,
	}
}

// Handle 处理鉴权
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 检查是否在排除列表中
		if m.isExcludedPath(path) {
			next(w, r)
			return
		}

		// 从Header获取Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// 验证Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		// 解析Token
		claims, err := m.jwtService.ParseToken(parts[1])
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// 检查管理员权限
		if m.isAdminPath(path) && claims.Role != "admin" {
			respondError(w, http.StatusForbidden, "admin permission required")
			return
		}

		// 将用户信息存入Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		// 将用户信息添加到Header，传递给后端服务
		r.Header.Set("X-User-ID", string(rune(claims.UserID)))
		r.Header.Set("X-Username", claims.Username)
		r.Header.Set("X-User-Role", claims.Role)

		next(w, r.WithContext(ctx))
	}
}

// isExcludedPath 检查路径是否在排除列表中
func (m *AuthMiddleware) isExcludedPath(path string) bool {
	for _, excluded := range m.excludePaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}

// isAdminPath 检查是否为管理员路径
func (m *AuthMiddleware) isAdminPath(path string) bool {
	for _, admin := range m.adminPaths {
		if strings.HasPrefix(path, admin) {
			return true
		}
	}
	return false
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"code":` + string(rune(code)) + `,"message":"` + message + `"}`))
}
