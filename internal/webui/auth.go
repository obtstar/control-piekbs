package webui

import (
	"net/http"
	"strings"
)

// WithAPIKey 对 /api/ 前缀路径启用 Bearer 认证（fork 补丁，FINDING-030，待上游评审）。
//
// REST 面（/api/*）此前整体无认证，server.api_key 只保护 /mcp。
// 语义与 mcp.withAuth 对齐：key 为空时全透传（保持现状），非空时
// /api/ 前缀要求 Authorization: Bearer <key>；页面与静态资源不保护，
// 浏览器 Web UI 仍可打开（其 JS 调 /api/* 需部署方自行解决，如反代注入头）。
func WithAPIKey(key string, h http.Handler) http.Handler {
	if key == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != key {
				w.Header().Set("WWW-Authenticate", `Bearer realm="PieKBS API"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
