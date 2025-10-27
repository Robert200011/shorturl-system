package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"redirect-service/internal/repo"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	visitRepo repo.VisitLogRepo
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(visitRepo repo.VisitLogRepo) *StatsHandler {
	return &StatsHandler{
		visitRepo: visitRepo,
	}
}

// GetStats 获取访问统计
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// 从URL路径提取短链码
	// /api/stats/:code
	path := strings.TrimPrefix(r.URL.Path, "/api/stats/")
	shortCode := strings.Split(path, "/")[0]

	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	stats, err := h.visitRepo.GetStats(ctx, shortCode)
	if err != nil {
		http.Error(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetRecentLogs 获取最近访问日志
func (h *StatsHandler) GetRecentLogs(w http.ResponseWriter, r *http.Request) {
	// 从URL路径提取短链码
	path := strings.TrimPrefix(r.URL.Path, "/api/stats/")
	parts := strings.Split(path, "/")
	shortCode := parts[0]

	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	// 获取limit参数
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := context.Background()
	logs, err := h.visitRepo.GetRecentLogs(ctx, shortCode, limit)
	if err != nil {
		http.Error(w, "Failed to get logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    logs,
	})
}
