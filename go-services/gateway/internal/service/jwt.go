package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims JWT声明
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	secret     []byte
	expireTime time.Duration
}

// NewJWTService 创建JWT服务
func NewJWTService(secret string, expireTime int64) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		expireTime: time.Duration(expireTime) * time.Second,
	}
}

// GenerateToken 生成Token
func (s *JWTService) GenerateToken(userID uint64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expireTime)),
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
		return s.secret, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
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

	// 如果距离过期时间小于30分钟，允许刷新
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", errors.New("token not close to expiration")
	}

	return s.GenerateToken(claims.UserID, claims.Username)
}
