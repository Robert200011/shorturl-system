package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shortener-service/internal/service"
	"shortener-service/internal/types"
)

// ShortenHandler 短链生成处理器
type ShortenHandler struct {
	svc service.ShortenerService
}

// NewShortenHandler 创建短链生成处理器
func NewShortenHandler(svc service.ShortenerService) *ShortenHandler {
	return &ShortenHandler{svc: svc}
}

// CreateShortLink 创建短链接
func (h *ShortenHandler) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	var req types.ShortenRequest

	// 手动解析 JSON，不使用 httpx.Parse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	// 基本验证
	if req.OriginalURL == "" {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, types.CommonResponse{
			Code:    1,
			Message: "original_url is required",
		})
		return
	}

	resp, err := h.svc.CreateShortLink(r.Context(), &req)
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

// GetShortLink 获取短链接详情
// GetShortLink 获取短链接详情
func (h *ShortenHandler) GetShortLink(w http.ResponseWriter, r *http.Request) {
	// 从 URL 路径获取参数
	code := r.URL.Path[len("/api/links/"):]
	if code == "" {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, types.CommonResponse{
			Code:    1,
			Message: "short_code is required",
		})
		return
	}

	resp, err := h.svc.GetShortLink(r.Context(), code)
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
