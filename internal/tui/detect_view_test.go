package tui

import (
	"strings"
	"testing"

	"github.com/fusheng/envkit/internal/cli/service"
)

func TestDetectTableNoMojibakeWithPlainStatus(t *testing.T) {
	report := service.DetectEnvironment()
	tbl := buildDetectTableFromReport(report, 100)
	view := tbl.View()
	if strings.Contains(view, "\ufffd") {
		t.Fatalf("table contains replacement char: %q", view)
	}
	if !strings.Contains(view, statusInstalled) && len(report.Languages) > 0 {
		// 有已安装语言时应能看到纯文本状态
		hasInstalled := false
		for _, tool := range report.Languages {
			if tool.Installed {
				hasInstalled = true
				break
			}
		}
		if hasInstalled {
			t.Fatalf("expected %q in table output", statusInstalled)
		}
	}
}

func TestTableCellPlainChinese(t *testing.T) {
	got := tableCell("已安装", 8)
	if got != "已安装" {
		t.Fatalf("got %q", got)
	}
}
