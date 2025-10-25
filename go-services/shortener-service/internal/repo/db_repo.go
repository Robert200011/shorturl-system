package repo

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"shortener-service/internal/model"
)

// ShortLinkRepo 短链接数据库操作接口
type ShortLinkRepo interface {
	Create(ctx context.Context, link *model.ShortLink) error
	GetByShortCode(ctx context.Context, code string) (*model.ShortLink, error)
	GetByOriginalURL(ctx context.Context, url string) (*model.ShortLink, error)
	Update(ctx context.Context, link *model.ShortLink) error
	IncrementVisitCount(ctx context.Context, code string) error
	List(ctx context.Context, offset, limit int) ([]*model.ShortLink, int64, error)
}

// shortLinkRepo 短链接数据库操作实现
type shortLinkRepo struct {
	db *gorm.DB
}

// NewShortLinkRepo 创建短链接数据库操作实例
func NewShortLinkRepo(dsn string) (ShortLinkRepo, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.ShortLink{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &shortLinkRepo{db: db}, nil
}

// Create 创建短链接
func (r *shortLinkRepo) Create(ctx context.Context, link *model.ShortLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

// GetByShortCode 根据短链码查询
func (r *shortLinkRepo) GetByShortCode(ctx context.Context, code string) (*model.ShortLink, error) {
	var link model.ShortLink
	err := r.db.WithContext(ctx).Where("short_code = ?", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// GetByOriginalURL 根据原始URL查询
func (r *shortLinkRepo) GetByOriginalURL(ctx context.Context, url string) (*model.ShortLink, error) {
	var link model.ShortLink
	err := r.db.WithContext(ctx).Where("original_url = ?", url).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// Update 更新短链接
func (r *shortLinkRepo) Update(ctx context.Context, link *model.ShortLink) error {
	return r.db.WithContext(ctx).Save(link).Error
}

// IncrementVisitCount 增加访问次数
func (r *shortLinkRepo) IncrementVisitCount(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Model(&model.ShortLink{}).
		Where("short_code = ?", code).
		UpdateColumn("visit_count", gorm.Expr("visit_count + ?", 1)).Error
}

// List 分页查询短链接列表
func (r *shortLinkRepo) List(ctx context.Context, offset, limit int) ([]*model.ShortLink, int64, error) {
	var links []*model.ShortLink
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.ShortLink{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&links).Error

	return links, total, err
}
