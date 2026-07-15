package tui

import (
	"strings"
	"testing"

	"github.com/fusheng/envkit/internal/cli/service"
)

func TestDetectTableNoMojibakeWithPlainStatus(t *testing.T) {
	report := service.DetectEnvironment()
	tbl := buildDetectTableFromReport(report, 100, 30)
	view := tbl.View()
	if strings.Contains(view, "\ufffd") {
		t.Fatalf("table contains replacement char: %q", view)
	}
	if !strings.Contains(view, statusInstalled) && len(report.Languages) > 0 {
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

func TestFormatStatusDisplay(t *testing.T) {
	cases := []struct {
		in, wantPrefix string
		kind           statusKind
	}{
		{"已安装", "✓", statusKindOK},
		{"已安装 (1.2.3)", "✓", statusKindOK},
		{"未安装", "✗", statusKindBad},
		{"可用", "●", statusKindWarn},
		{"", "—", statusKindMuted},
	}
	for _, c := range cases {
		got := formatStatusDisplay(c.in)
		if !strings.HasPrefix(got, c.wantPrefix) && got != c.wantPrefix {
			t.Errorf("formatStatusDisplay(%q)=%q, want prefix %q", c.in, got, c.wantPrefix)
		}
		if classifyStatus(got) != c.kind && classifyStatus(c.in) != c.kind {
			t.Errorf("classifyStatus(%q)=%v, want %v", c.in, classifyStatus(c.in), c.kind)
		}
	}
}

func TestAllocColumnWidths(t *testing.T) {
	w := allocColumnWidths(60, []int{2, 2, 3, 5}, []int{6, 6, 8, 8})
	if len(w) != 4 {
		t.Fatalf("len=%d", len(w))
	}
	sum := 0
	for _, x := range w {
		sum += x
	}
	// sum of widths + 3 gaps should be around 60
	if sum < 40 {
		t.Fatalf("widths too small: %v sum=%d", w, sum)
	}
}

func TestRenderStepIndicator(t *testing.T) {
	s := RenderStepIndicator([]string{"选择", "确认", "完成"}, 1)
	if s == "" {
		t.Fatal("empty step indicator")
	}
	if !strings.Contains(s, "选择") || !strings.Contains(s, "确认") {
		t.Fatalf("missing labels: %q", s)
	}
}

func TestCatalogTableSections(t *testing.T) {
	rows := service.ListCatalog()
	tbl := buildCatalogTable(rows, 100, 30)
	view := tbl.View()
	if !strings.Contains(view, "语言") && !strings.Contains(view, "工具") {
		t.Fatalf("expected category sections in view: %q", view)
	}
}
