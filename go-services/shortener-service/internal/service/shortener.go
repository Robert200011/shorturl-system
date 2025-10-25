package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"shortener-service/internal/model"
	"shortener-service/internal/repo"
	"shortener-service/internal/types"
)

var (
	ErrShortCodeExists   = errors.New("short code already exists")
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrURLInvalid        = errors.New("invalid url")
)

// ShortenerService 短链服务接口
type ShortenerService interface {
	CreateShortLink(ctx context.Context, req *types.ShortenRequest) (*types.ShortenResponse, error)
	BatchCreateShortLinks(ctx context.Context, urls []string) (*types.BatchShortenResponse, error)
	GetShortLink(ctx context.Context, code string) (*types.GetLinkResponse, error)
	GetOriginalURL(ctx context.Context, code string) (string, error)
}

// shortenerService 短链服务实现
type shortenerService struct {
	dbRepo    repo.ShortLinkRepo
	redisRepo repo.RedisRepo
	idGen     IDGenerator
	domain    string
	cacheTTL  time.Duration
}

// NewShortenerService 创建短链服务实例
func NewShortenerService(
	dbRepo repo.ShortLinkRepo,
	redisRepo repo.RedisRepo,
	idGen IDGenerator,
	domain string,
	cacheTTL int,
) ShortenerService {
	return &shortenerService{
		dbRepo:    dbRepo,
		redisRepo: redisRepo,
		idGen:     idGen,
		domain:    domain,
		cacheTTL:  time.Duration(cacheTTL) * time.Second,
	}
}

// CreateShortLink 创建短链接
func (s *shortenerService) CreateShortLink(ctx context.Context, req *types.ShortenRequest) (*types.ShortenResponse, error) {
	// 检查是否已存在相同URL的短链
	if code, err := s.redisRepo.GetShortCodeByURL(ctx, req.OriginalURL); err == nil && code != "" {
		link, err := s.dbRepo.GetByShortCode(ctx, code)
		if err == nil && link != nil {
			return s.buildResponse(link), nil
		}
	}

	// 查询数据库
	if link, err := s.dbRepo.GetByOriginalURL(ctx, req.OriginalURL); err == nil {
		// 更新缓存
		_ = s.redisRepo.SetShortLink(ctx, link, s.cacheTTL)
		return s.buildResponse(link), nil
	}

	// 生成短链码
	var shortCode string
	var err error

	if req.CustomCode != "" {
		// 使用自定义短链码
		shortCode = req.CustomCode
		// 检查是否已存在
		if exists, _ := s.redisRepo.Exists(ctx, shortCode); exists {
			return nil, ErrShortCodeExists
		}
		if _, err := s.dbRepo.GetByShortCode(ctx, shortCode); err == nil {
			return nil, ErrShortCodeExists
		}
	} else {
		// 自动生成短链码
		shortCode, err = s.idGen.GenerateShortCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// 创建短链接记录
	link := &model.ShortLink{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		Title:       req.Title,
		Description: req.Description,
		ExpireAt:    req.ExpireAt,
		Status:      1,
	}

	// 保存到数据库
	if err := s.dbRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create short link: %w", err)
	}

	// 缓存到Redis
	_ = s.redisRepo.SetShortLink(ctx, link, s.cacheTTL)

	return s.buildResponse(link), nil
}

// BatchCreateShortLinks 批量创建短链接
func (s *shortenerService) BatchCreateShortLinks(ctx context.Context, urls []string) (*types.BatchShortenResponse, error) {
	response := &types.BatchShortenResponse{
		Results: make([]types.ShortenResponse, 0, len(urls)),
	}

	for _, url := range urls {
		req := &types.ShortenRequest{
			OriginalURL: url,
		}

		result, err := s.CreateShortLink(ctx, req)
		if err != nil {
			response.Failed++
			continue
		}

		response.Results = append(response.Results, *result)
		response.Success++
	}

	return response, nil
}

// GetShortLink 获取短链接详情
func (s *shortenerService) GetShortLink(ctx context.Context, code string) (*types.GetLinkResponse, error) {
	// 先查缓存
	link, err := s.redisRepo.GetShortLink(ctx, code)
	if err == nil && link != nil {
		return s.buildDetailResponse(link), nil
	}

	// 查数据库
	link, err = s.dbRepo.GetByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShortCodeNotFound
		}
		return nil, err
	}

	// 更新缓存
	_ = s.redisRepo.SetShortLink(ctx, link, s.cacheTTL)

	return s.buildDetailResponse(link), nil
}

// GetOriginalURL 获取原始URL（用于重定向）
func (s *shortenerService) GetOriginalURL(ctx context.Context, code string) (string, error) {
	// 先查缓存
	link, err := s.redisRepo.GetShortLink(ctx, code)
	if err == nil && link != nil {
		if !link.IsActive() {
			return "", errors.New("short link is inactive or expired")
		}
		// 异步增加访问计数
		go func() {
			_ = s.dbRepo.IncrementVisitCount(context.Background(), code)
		}()
		return link.OriginalURL, nil
	}

	// 查数据库
	link, err = s.dbRepo.GetByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrShortCodeNotFound
		}
		return "", err
	}

	if !link.IsActive() {
		return "", errors.New("short link is inactive or expired")
	}

	// 更新缓存
	_ = s.redisRepo.SetShortLink(ctx, link, s.cacheTTL)

	// 异步增加访问计数
	go func() {
		_ = s.dbRepo.IncrementVisitCount(context.Background(), code)
	}()

	return link.OriginalURL, nil
}

// buildResponse 构建响应
func (s *shortenerService) buildResponse(link *model.ShortLink) *types.ShortenResponse {
	return &types.ShortenResponse{
		ShortCode:   link.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.domain, link.ShortCode),
		OriginalURL: link.OriginalURL,
		CreatedAt:   link.CreatedAt,
	}
}

// buildDetailResponse 构建详情响应
func (s *shortenerService) buildDetailResponse(link *model.ShortLink) *types.GetLinkResponse {
	return &types.GetLinkResponse{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		Title:       link.Title,
		Description: link.Description,
		VisitCount:  link.VisitCount,
		Status:      link.Status,
		ExpireAt:    link.ExpireAt,
		CreatedAt:   link.CreatedAt,
	}
}
