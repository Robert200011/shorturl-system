package model

import (
	"time"
)

// ShortLink 短链接模型
type ShortLink struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode   string     `gorm:"uniqueIndex;size:20;not null" json:"short_code"`
	OriginalURL string     `gorm:"size:2048;not null" json:"original_url"`
	UserID      *uint64    `gorm:"index" json:"user_id,omitempty"`
	Title       string     `gorm:"size:255" json:"title,omitempty"`
	Description string     `gorm:"size:500" json:"description,omitempty"`
	VisitCount  uint64     `gorm:"default:0" json:"visit_count"`
	Status      int8       `gorm:"default:1" json:"status"` // 0-禁用 1-启用
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ShortLink) TableName() string {
	return "short_links"
}

// IsExpired 检查是否过期
func (s *ShortLink) IsExpired() bool {
	if s.ExpireAt == nil {
		return false
	}
	return time.Now().After(*s.ExpireAt)
}

// IsActive 检查是否激活
func (s *ShortLink) IsActive() bool {
	return s.Status == 1 && !s.IsExpired()
}
