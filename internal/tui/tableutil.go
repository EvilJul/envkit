package tui

import "github.com/mattn/go-runewidth"

// tableCell 表格单元格纯文本（禁止在单元格内使用 lipgloss，否则会乱码）
func tableCell(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	return runewidth.Truncate(s, maxWidth, "…")
}

const (
	statusInstalled = "已安装"
	statusAvailable = "可用"
	statusNotInstalled = "未安装"
)
