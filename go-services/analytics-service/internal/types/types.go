package types

// DailyStatsRequest 每日统计请求
type DailyStatsRequest struct {
	ShortCode string `uri:"code" binding:"required"`
	StartDate string `form:"start_date"` // YYYY-MM-DD
	EndDate   string `form:"end_date"`   // YYYY-MM-DD
}

// HourlyStatsRequest 每小时统计请求
type HourlyStatsRequest struct {
	ShortCode string `uri:"code" binding:"required"`
	Date      string `form:"date" binding:"required"` // YYYY-MM-DD
}

// DimensionStatsRequest 维度统计请求
type DimensionStatsRequest struct {
	ShortCode string `uri:"code" binding:"required"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
