package middleware

import (
	"context"
	"net/http"
	"strings"

	"gateway/internal/service"
	"gateway/internal/types"
)

// contextKey 上下文键类型
type contextKey string

const (
	// UserIDKey 用户ID键
	UserIDKey contextKey = "user_id"
	// UsernameKey 用户名键
	UsernameKey contextKey = "username"
)

// AuthMiddleware JWT鉴权中间件
type AuthMiddleware struct {
	jwtService *service.JWTService
	// 不需要鉴权的路径
	publicPaths []string
}

// NewAuthMiddleware 创建鉴权中间件
func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		publicPaths: []string{
			"/api/auth/login",    // 登录接口
			"/api/auth/register", // 注册接口
			"/health",            // 健康检查
		},
	}
}

// Handle 处理鉴权
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是公开路径
		if m.isPublicPath(r.URL.Path) || m.isRedirectPath(r.URL.Path) {
			next(w, r)
			return
		}

		// 获取Token
		token := m.extractToken(r)
		if token == "" {
			m.unauthorizedResponse(w, "missing authorization token")
			return
		}

		// 验证Token
		claims, err := m.jwtService.ParseToken(token)
		if err != nil {
			m.unauthorizedResponse(w, "invalid or expired token")
			return
		}

		// 将用户信息存入上下文
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)

		next(w, r.WithContext(ctx))
	}
}

// isPublicPath 检查是否是公开路径
func (m *AuthMiddleware) isPublicPath(path string) bool {
	for _, publicPath := range m.publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// isRedirectPath 检查是否是短链重定向路径
func (m *AuthMiddleware) isRedirectPath(path string) bool {
	// 短链重定向路径格式: /:code (只包含字母数字)
	if len(path) > 1 && !strings.Contains(path[1:], "/") {
		return true
	}
	return false
}

// extractToken 从请求中提取Token
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// 从 Header 获取
	bearerToken := r.Header.Get("Authorization")
	if bearerToken != "" {
		// Bearer Token格式
		parts := strings.SplitN(bearerToken, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从 Query 参数获取
	return r.URL.Query().Get("token")
}

// unauthorizedResponse 返回未授权响应
func (m *AuthMiddleware) unauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := types.ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: message,
	}

	// 简单的JSON序列化
	w.Write([]byte(`{"code":` + string(rune(response.Code)) + `,"message":"` + response.Message + `"}`))
}

// GetUserID 从上下文获取用户ID
func GetUserID(ctx context.Context) uint64 {
	if userID, ok := ctx.Value(UserIDKey).(uint64); ok {
		return userID
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}
