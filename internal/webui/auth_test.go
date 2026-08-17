//go:build fts5

package webui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// okHandler 模拟被保护的内层 handler，永远 200。
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// WithAPIKey 语义（fork 补丁，FINDING-030）：
// key 为空 → 全透传；key 非空 → /api/ 前缀路径要求 Authorization: Bearer，
// 页面/静态等非 /api 路径保持公开（浏览器 Web UI 不受影响）。
func TestWithAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		path   string
		header string
		want   int
	}{
		{"无key透传", "", "/api/search", "", http.StatusOK},
		{"有key无头拒绝", "k1", "/api/search", "", http.StatusUnauthorized},
		{"有key错误token拒绝", "k1", "/api/search", "Bearer bad", http.StatusUnauthorized},
		{"有key正确放行", "k1", "/api/search", "Bearer k1", http.StatusOK},
		{"页面路径不受保护", "k1", "/", "", http.StatusOK},
		{"静态资源不受保护", "k1", "/static/app.js", "", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := WithAPIKey(tt.key, okHandler())
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("path=%s key=%q header=%q: got %d, want %d",
					tt.path, tt.key, tt.header, w.Code, tt.want)
			}
		})
	}
}
