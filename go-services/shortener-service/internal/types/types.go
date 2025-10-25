package types

import "time"

// ShortenRequest 短链生成请求
type ShortenRequest struct {
	OriginalURL string     `json:"original_url" binding:"required,url"`
	CustomCode  string     `json:"custom_code,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
}

// ShortenResponse 短链生成响应
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// BatchShortenRequest 批量短链生成请求
type BatchShortenRequest struct {
	URLs []string `json:"urls" binding:"required,min=1,max=100"`
}

// BatchShortenResponse 批量短链生成响应
type BatchShortenResponse struct {
	Results []ShortenResponse `json:"results"`
	Success int               `json:"success"`
	Failed  int               `json:"failed"`
}

// GetLinkRequest 查询短链请求
type GetLinkRequest struct {
	ShortCode string `uri:"code" binding:"required"`
}

// GetLinkResponse 查询短链响应
type GetLinkResponse struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	VisitCount  uint64     `json:"visit_count"`
	Status      int8       `json:"status"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
