package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims JWT声明
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"` // user, admin
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	secret     []byte
	expireHour int
}

// NewJWTService 创建JWT服务
func NewJWTService(secret string, expireHour int) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		expireHour: expireHour,
	}
}

// GenerateToken 生成Token
func (s *JWTService) GenerateToken(userID uint64, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ParseToken 解析Token
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, ErrExpiredToken
		}
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新Token
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果Token还有超过24小时才过期，不允许刷新
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > 24*time.Hour {
		return "", errors.New("token is still valid, cannot refresh yet")
	}

	return s.GenerateToken(claims.UserID, claims.Username, claims.Role)
}
