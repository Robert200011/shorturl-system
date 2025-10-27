package handler

import (
	"encoding/json" // 🆕 添加这行
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shortener-service/internal/service"
	"shortener-service/internal/types"
)

// BatchHandler 批量处理器
type BatchHandler struct {
	svc service.ShortenerService
}

// NewBatchHandler 创建批量处理器
func NewBatchHandler(svc service.ShortenerService) *BatchHandler {
	return &BatchHandler{svc: svc}
}

// BatchCreateShortLinks 批量创建短链接
func (h *BatchHandler) BatchCreateShortLinks(w http.ResponseWriter, r *http.Request) {
	var req types.BatchShortenRequest

	// 手动解析 JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	// 基本验证
	if len(req.URLs) == 0 {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, types.CommonResponse{
			Code:    1,
			Message: "urls is required",
		})
		return
	}

	resp, err := h.svc.BatchCreateShortLinks(r.Context(), req.URLs)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	httpx.OkJsonCtx(r.Context(), w, types.CommonResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}
