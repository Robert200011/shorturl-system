package model

import "time"

// AnalyticsDaily 每日统计
type AnalyticsDaily struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode      string    `gorm:"index;size:20;not null" json:"short_code"`
	Date           string    `gorm:"index;size:10;not null" json:"date"` // YYYY-MM-DD
	TotalVisits    int64     `gorm:"default:0" json:"total_visits"`
	UniqueVisitors int64     `gorm:"default:0" json:"unique_visitors"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AnalyticsDaily) TableName() string {
	return "analytics_daily"
}

// AnalyticsHourly 每小时统计
type AnalyticsHourly struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	Hour       string    `gorm:"index;size:13;not null" json:"hour"` // YYYY-MM-DD HH
	VisitCount int64     `gorm:"default:0" json:"visit_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AnalyticsHourly) TableName() string {
	return "analytics_hourly"
}

// AnalyticsBrowser 浏览器统计
type AnalyticsBrowser struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	Browser    string    `gorm:"size:50;not null" json:"browser"`
	VisitCount int64     `gorm:"default:0" json:"visit_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AnalyticsBrowser) TableName() string {
	return "analytics_browser"
}

// AnalyticsDevice 设备统计
type AnalyticsDevice struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	DeviceType string    `gorm:"size:20;not null" json:"device_type"`
	VisitCount int64     `gorm:"default:0" json:"visit_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AnalyticsDevice) TableName() string {
	return "analytics_device"
}

// AnalyticsOS 操作系统统计
type AnalyticsOS struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	OS         string    `gorm:"size:50;not null" json:"os"`
	VisitCount int64     `gorm:"default:0" json:"visit_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AnalyticsOS) TableName() string {
	return "analytics_os"
}

// VisitEvent 访问事件（从Kafka接收）
type VisitEvent struct {
	ShortCode  string `json:"short_code"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Referer    string `json:"referer"`
	DeviceType string `json:"device_type"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	Timestamp  int64  `json:"timestamp"`
}
