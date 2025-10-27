package model

import "time"

// VisitLog 访问日志模型
type VisitLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	IP         string    `gorm:"size:45" json:"ip"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Referer    string    `gorm:"size:500" json:"referer"`
	Country    string    `gorm:"size:50" json:"country,omitempty"`
	Province   string    `gorm:"size:50" json:"province,omitempty"`
	City       string    `gorm:"size:50" json:"city,omitempty"`
	DeviceType string    `gorm:"size:20" json:"device_type"`
	Browser    string    `gorm:"size:50" json:"browser"`
	OS         string    `gorm:"size:50" json:"os"`
	VisitedAt  time.Time `gorm:"index;not null" json:"visited_at"`
}

// TableName 指定表名
func (VisitLog) TableName() string {
	return "visit_logs"
}

// VisitStats 访问统计
type VisitStats struct {
	ShortCode    string     `json:"short_code"`
	TotalVisits  int64      `json:"total_visits"`
	UniqueVisits int64      `json:"unique_visits"`
	TodayVisits  int64      `json:"today_visits"`
	TopBrowsers  []StatItem `json:"top_browsers"`
	TopDevices   []StatItem `json:"top_devices"`
	TopOS        []StatItem `json:"top_os"`
}

// StatItem 统计项
type StatItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
