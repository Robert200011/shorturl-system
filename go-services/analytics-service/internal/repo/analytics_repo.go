package repo

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"analytics-service/internal/model"
)

// AnalyticsRepo 统计数据仓库接口
type AnalyticsRepo interface {
	// 每日统计
	UpsertDaily(ctx context.Context, daily *model.AnalyticsDaily) error
	GetDaily(ctx context.Context, shortCode, date string) (*model.AnalyticsDaily, error)
	GetDailyRange(ctx context.Context, shortCode, startDate, endDate string) ([]*model.AnalyticsDaily, error)

	// 每小时统计
	UpsertHourly(ctx context.Context, hourly *model.AnalyticsHourly) error
	GetHourlyRange(ctx context.Context, shortCode, startHour, endHour string) ([]*model.AnalyticsHourly, error)

	// 浏览器统计
	UpsertBrowser(ctx context.Context, browser *model.AnalyticsBrowser) error
	GetTopBrowsers(ctx context.Context, shortCode string, limit int) ([]*model.AnalyticsBrowser, error)

	// 设备统计
	UpsertDevice(ctx context.Context, device *model.AnalyticsDevice) error
	GetDeviceStats(ctx context.Context, shortCode string) ([]*model.AnalyticsDevice, error)

	// 操作系统统计
	UpsertOS(ctx context.Context, os *model.AnalyticsOS) error
	GetTopOS(ctx context.Context, shortCode string, limit int) ([]*model.AnalyticsOS, error)
}

// analyticsRepo 实现
type analyticsRepo struct {
	db *gorm.DB
}

// NewAnalyticsRepo 创建仓库实例
func NewAnalyticsRepo(dsn string) (AnalyticsRepo, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.AnalyticsDaily{},
		&model.AnalyticsHourly{},
		&model.AnalyticsBrowser{},
		&model.AnalyticsDevice{},
		&model.AnalyticsOS{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &analyticsRepo{db: db}, nil
}

// UpsertDaily 插入或更新每日统计
func (r *analyticsRepo) UpsertDaily(ctx context.Context, daily *model.AnalyticsDaily) error {
	return r.db.WithContext(ctx).
		Where("short_code = ? AND date = ?", daily.ShortCode, daily.Date).
		Assign(map[string]interface{}{
			"total_visits":    gorm.Expr("total_visits + ?", 1),
			"unique_visitors": daily.UniqueVisitors,
		}).
		FirstOrCreate(daily).Error
}

// GetDaily 获取每日统计
func (r *analyticsRepo) GetDaily(ctx context.Context, shortCode, date string) (*model.AnalyticsDaily, error) {
	var daily model.AnalyticsDaily
	err := r.db.WithContext(ctx).
		Where("short_code = ? AND date = ?", shortCode, date).
		First(&daily).Error
	if err != nil {
		return nil, err
	}
	return &daily, nil
}

// GetDailyRange 获取日期范围内的统计
func (r *analyticsRepo) GetDailyRange(ctx context.Context, shortCode, startDate, endDate string) ([]*model.AnalyticsDaily, error) {
	var dailies []*model.AnalyticsDaily
	err := r.db.WithContext(ctx).
		Where("short_code = ? AND date BETWEEN ? AND ?", shortCode, startDate, endDate).
		Order("date ASC").
		Find(&dailies).Error
	return dailies, err
}

// UpsertHourly 插入或更新每小时统计
func (r *analyticsRepo) UpsertHourly(ctx context.Context, hourly *model.AnalyticsHourly) error {
	return r.db.WithContext(ctx).
		Where("short_code = ? AND hour = ?", hourly.ShortCode, hourly.Hour).
		Assign(map[string]interface{}{
			"visit_count": gorm.Expr("visit_count + ?", 1),
		}).
		FirstOrCreate(hourly).Error
}

// GetHourlyRange 获取小时范围内的统计
func (r *analyticsRepo) GetHourlyRange(ctx context.Context, shortCode, startHour, endHour string) ([]*model.AnalyticsHourly, error) {
	var hourlies []*model.AnalyticsHourly
	err := r.db.WithContext(ctx).
		Where("short_code = ? AND hour BETWEEN ? AND ?", shortCode, startHour, endHour).
		Order("hour ASC").
		Find(&hourlies).Error
	return hourlies, err
}

// UpsertBrowser 插入或更新浏览器统计
func (r *analyticsRepo) UpsertBrowser(ctx context.Context, browser *model.AnalyticsBrowser) error {
	return r.db.WithContext(ctx).
		Where("short_code = ? AND browser = ?", browser.ShortCode, browser.Browser).
		Assign(map[string]interface{}{
			"visit_count": gorm.Expr("visit_count + ?", 1),
		}).
		FirstOrCreate(browser).Error
}

// GetTopBrowsers 获取Top浏览器
func (r *analyticsRepo) GetTopBrowsers(ctx context.Context, shortCode string, limit int) ([]*model.AnalyticsBrowser, error) {
	var browsers []*model.AnalyticsBrowser
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		Order("visit_count DESC").
		Limit(limit).
		Find(&browsers).Error
	return browsers, err
}

// UpsertDevice 插入或更新设备统计
func (r *analyticsRepo) UpsertDevice(ctx context.Context, device *model.AnalyticsDevice) error {
	return r.db.WithContext(ctx).
		Where("short_code = ? AND device_type = ?", device.ShortCode, device.DeviceType).
		Assign(map[string]interface{}{
			"visit_count": gorm.Expr("visit_count + ?", 1),
		}).
		FirstOrCreate(device).Error
}

// GetDeviceStats 获取设备统计
func (r *analyticsRepo) GetDeviceStats(ctx context.Context, shortCode string) ([]*model.AnalyticsDevice, error) {
	var devices []*model.AnalyticsDevice
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		Order("visit_count DESC").
		Find(&devices).Error
	return devices, err
}

// UpsertOS 插入或更新操作系统统计
func (r *analyticsRepo) UpsertOS(ctx context.Context, os *model.AnalyticsOS) error {
	return r.db.WithContext(ctx).
		Where("short_code = ? AND os = ?", os.ShortCode, os.OS).
		Assign(map[string]interface{}{
			"visit_count": gorm.Expr("visit_count + ?", 1),
		}).
		FirstOrCreate(os).Error
}

// GetTopOS 获取Top操作系统
func (r *analyticsRepo) GetTopOS(ctx context.Context, shortCode string, limit int) ([]*model.AnalyticsOS, error) {
	var osList []*model.AnalyticsOS
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		Order("visit_count DESC").
		Limit(limit).
		Find(&osList).Error
	return osList, err
}
