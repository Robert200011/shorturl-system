package service

import (
	"net/http"
	"strings"

	"github.com/mileusna/useragent"
)

// VisitInfo 访问信息
type VisitInfo struct {
	IP         string
	UserAgent  string
	Referer    string
	DeviceType string
	Browser    string
	OS         string
}

// ParseRequest 解析HTTP请求，提取访问信息
func ParseRequest(r *http.Request) *VisitInfo {
	info := &VisitInfo{
		IP:        extractIP(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}

	// 解析 User-Agent
	ua := useragent.Parse(r.UserAgent())

	// 设备类型
	if ua.Mobile {
		info.DeviceType = "Mobile"
	} else if ua.Tablet {
		info.DeviceType = "Tablet"
	} else if ua.Desktop {
		info.DeviceType = "Desktop"
	} else if ua.Bot {
		info.DeviceType = "Bot"
	} else {
		info.DeviceType = "Unknown"
	}

	// 浏览器
	if ua.Name != "" {
		info.Browser = ua.Name
		if ua.Version != "" {
			info.Browser += " " + ua.Version
		}
	} else {
		info.Browser = "Unknown"
	}

	// 操作系统
	if ua.OS != "" {
		info.OS = ua.OS
		if ua.OSVersion != "" {
			info.OS += " " + ua.OSVersion
		}
	} else {
		info.OS = "Unknown"
	}

	return info
}

// extractIP 提取真实IP地址
func extractIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-IP 获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用 RemoteAddr
	ip := r.RemoteAddr
	// 去掉端口号
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
