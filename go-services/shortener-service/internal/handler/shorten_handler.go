package handler

import (
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

	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
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
func (h *ShortenHandler) GetShortLink(w http.ResponseWriter, r *http.Request) {
	var req types.GetLinkRequest

	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	resp, err := h.svc.GetShortLink(r.Context(), req.ShortCode)
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
