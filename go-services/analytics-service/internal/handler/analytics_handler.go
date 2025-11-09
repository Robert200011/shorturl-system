package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"analytics-service/internal/repo"
	"analytics-service/internal/types"
)

// AnalyticsHandler 统计处理器
type AnalyticsHandler struct {
	repo repo.AnalyticsRepo
}

// NewAnalyticsHandler 创建处理器
func NewAnalyticsHandler(repo repo.AnalyticsRepo) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

// GetDailyStats 获取每日统计
func (h *AnalyticsHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	// 从URL路径提取短链码
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/daily/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		h.errorResponse(w, "short_code is required", http.StatusBadRequest)
		return
	}

	// 获取日期范围参数
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// 默认最近7天
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	ctx := context.Background()
	dailies, err := h.repo.GetDailyRange(ctx, shortCode, startDate, endDate)
	if err != nil {
		h.errorResponse(w, "failed to get daily stats", http.StatusInternalServerError)
		return
	}

	h.successResponse(w, dailies)
}

// GetHourlyStats 获取每小时统计
func (h *AnalyticsHandler) GetHourlyStats(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/hourly/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		h.errorResponse(w, "short_code is required", http.StatusBadRequest)
		return
	}

	// 获取日期参数，默认今天
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	startHour := date + " 00"
	endHour := date + " 23"

	ctx := context.Background()
	hourlies, err := h.repo.GetHourlyRange(ctx, shortCode, startHour, endHour)
	if err != nil {
		h.errorResponse(w, "failed to get hourly stats", http.StatusInternalServerError)
		return
	}

	h.successResponse(w, hourlies)
}

// GetBrowserStats 获取浏览器统计
func (h *AnalyticsHandler) GetBrowserStats(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/browser/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		h.errorResponse(w, "short_code is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	browsers, err := h.repo.GetTopBrowsers(ctx, shortCode, 10)
	if err != nil {
		h.errorResponse(w, "failed to get browser stats", http.StatusInternalServerError)
		return
	}

	h.successResponse(w, browsers)
}

// GetDeviceStats 获取设备统计
func (h *AnalyticsHandler) GetDeviceStats(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/device/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		h.errorResponse(w, "short_code is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	devices, err := h.repo.GetDeviceStats(ctx, shortCode)
	if err != nil {
		h.errorResponse(w, "failed to get device stats", http.StatusInternalServerError)
		return
	}

	h.successResponse(w, devices)
}

// GetOSStats 获取操作系统统计
func (h *AnalyticsHandler) GetOSStats(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analytics/os/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		h.errorResponse(w, "short_code is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	osList, err := h.repo.GetTopOS(ctx, shortCode, 10)
	if err != nil {
		h.errorResponse(w, "failed to get os stats", http.StatusInternalServerError)
		return
	}

	h.successResponse(w, osList)
}

// successResponse 成功响应
func (h *AnalyticsHandler) successResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := types.CommonResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// errorResponse 错误响应
func (h *AnalyticsHandler) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := types.CommonResponse{
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
