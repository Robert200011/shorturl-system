package repo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"redirect-service/internal/model"
)

// VisitLogRepo 访问日志数据库操作接口
type VisitLogRepo interface {
	Create(ctx context.Context, log *model.VisitLog) error
	GetStats(ctx context.Context, shortCode string) (*model.VisitStats, error)
	GetRecentLogs(ctx context.Context, shortCode string, limit int) ([]*model.VisitLog, error)
}

// visitLogRepo 访问日志数据库操作实现
type visitLogRepo struct {
	db *gorm.DB
}

// NewVisitLogRepo 创建访问日志数据库操作实例
func NewVisitLogRepo(dsn string) (VisitLogRepo, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.VisitLog{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &visitLogRepo{db: db}, nil
}

// Create 创建访问日志
func (r *visitLogRepo) Create(ctx context.Context, log *model.VisitLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetStats 获取访问统计
func (r *visitLogRepo) GetStats(ctx context.Context, shortCode string) (*model.VisitStats, error) {
	stats := &model.VisitStats{
		ShortCode: shortCode,
	}

	// 总访问次数
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Where("short_code = ?", shortCode).
		Count(&stats.TotalVisits)

	// 独立访客数（按IP去重）
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Where("short_code = ?", shortCode).
		Distinct("ip").
		Count(&stats.UniqueVisits)

	// 今日访问次数
	today := time.Now().Format("2006-01-02")
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Where("short_code = ? AND DATE(visited_at) = ?", shortCode, today).
		Count(&stats.TodayVisits)

	// Top浏览器
	var browserStats []model.StatItem
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Select("browser as name, COUNT(*) as count").
		Where("short_code = ? AND browser != ''", shortCode).
		Group("browser").
		Order("count DESC").
		Limit(5).
		Find(&browserStats)
	stats.TopBrowsers = browserStats

	// Top设备类型
	var deviceStats []model.StatItem
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Select("device_type as name, COUNT(*) as count").
		Where("short_code = ? AND device_type != ''", shortCode).
		Group("device_type").
		Order("count DESC").
		Limit(5).
		Find(&deviceStats)
	stats.TopDevices = deviceStats

	// Top操作系统
	var osStats []model.StatItem
	r.db.WithContext(ctx).Model(&model.VisitLog{}).
		Select("os as name, COUNT(*) as count").
		Where("short_code = ? AND os != ''", shortCode).
		Group("os").
		Order("count DESC").
		Limit(5).
		Find(&osStats)
	stats.TopOS = osStats

	return stats, nil
}

// GetRecentLogs 获取最近的访问日志
func (r *visitLogRepo) GetRecentLogs(ctx context.Context, shortCode string, limit int) ([]*model.VisitLog, error) {
	var logs []*model.VisitLog
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		Order("visited_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
