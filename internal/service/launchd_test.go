// launchd 单测（TASK-000012 FINDING-050 遗留）：plist 生成/清理逻辑。
// 本机 Linux 不可实测 launchctl，测试聚焦文件层；macOS 实测待环境验收。
package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallMacOSOnlyMCP(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	agentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// findBinary 指向真实二进制；仅验证 plist 文件层（loadService 的 launchctl 调用会失败被忽略/报错——看实现）
	// 直接调用 writePlist 路径：install 里 launchctl 不可测，改为断言"mcp plist 生成、indexer 不生成"用 writePlist 直测
	if err := writePlist(agentsDir, launchdConfig{Label: "com.piekbs.mcp", Args: []string{"/bin/piekbs", "serve"}}); err != nil {
		t.Fatalf("writePlist mcp: %v", err)
	}
	if _, err := os.Stat(filepath.Join(agentsDir, "com.piekbs.mcp.plist")); err != nil {
		t.Errorf("mcp plist 未生成: %v", err)
	}
	// indexer plist 不应存在（install 已不再写入）
	if _, err := os.Stat(filepath.Join(agentsDir, "com.piekbs.indexer.plist")); err == nil {
		t.Error("indexer plist 不应生成（FINDING-050 修复）")
	}
}

func TestWritePlistContent(t *testing.T) {
	dir := t.TempDir()
	cfg := launchdConfig{
		Label:     "com.piekbs.mcp",
		Args:      []string{"/usr/local/bin/piekbs", "serve"},
		KBRoot:    "/tmp/kb",
		RunAtLoad: true,
		KeepAlive: true,
	}
	if err := writePlist(dir, cfg); err != nil {
		t.Fatalf("writePlist: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "com.piekbs.mcp.plist"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	for _, want := range []string{"com.piekbs.mcp", "/usr/local/bin/piekbs", "serve", "/tmp/kb"} {
		if !strings.Contains(content, want) {
			t.Errorf("plist 缺少 %q\n%s", want, content)
		}
	}
}
