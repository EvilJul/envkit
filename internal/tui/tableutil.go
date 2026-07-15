package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// tableCell 表格单元格纯文本截断（中文宽度安全）
func tableCell(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	return runewidth.Truncate(s, maxWidth, "…")
}

const (
	statusInstalled    = "已安装"
	statusAvailable    = "可用"
	statusNotInstalled = "未安装"
)

// statusKind classifies install-ish status for styling.
type statusKind int

const (
	statusKindNone statusKind = iota
	statusKindOK
	statusKindWarn
	statusKindBad
	statusKindMuted
)

func classifyStatus(raw string) statusKind {
	s := strings.TrimSpace(raw)
	if s == "" || s == "—" {
		return statusKindMuted
	}
	if strings.Contains(s, statusNotInstalled) || strings.HasPrefix(s, "✗") {
		return statusKindBad
	}
	if strings.Contains(s, statusAvailable) {
		return statusKindWarn
	}
	if strings.Contains(s, statusInstalled) || strings.HasPrefix(s, "✓") {
		return statusKindOK
	}
	return statusKindMuted
}

// formatStatusDisplay returns plain status text with symbol (for width calc / storage).
func formatStatusDisplay(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "—"
	}
	switch classifyStatus(s) {
	case statusKindOK:
		if strings.HasPrefix(s, "✓") {
			return s
		}
		return "✓ " + s
	case statusKindBad:
		if strings.HasPrefix(s, "✗") {
			return s
		}
		return "✗ " + s
	case statusKindWarn:
		if strings.HasPrefix(s, "●") {
			return s
		}
		return "● " + s
	default:
		return s
	}
}

func padCell(s string, width int) string {
	if width <= 0 {
		return s
	}
	w := runewidth.StringWidth(s)
	if w > width {
		return runewidth.Truncate(s, width, "…")
	}
	return s + strings.Repeat(" ", width-w)
}

// allocColumnWidths distributes available width by weight ratios.
func allocColumnWidths(total int, weights []int, mins []int) []int {
	n := len(weights)
	if n == 0 {
		return nil
	}
	if len(mins) != n {
		mins = make([]int, n)
		for i := range mins {
			mins[i] = 4
		}
	}
	gap := n - 1
	avail := total - gap
	if avail < 0 {
		avail = 0
	}

	sumW := 0
	for _, w := range weights {
		sumW += w
	}
	if sumW < 1 {
		sumW = n
	}

	out := make([]int, n)
	used := 0
	for i := 0; i < n-1; i++ {
		w := avail * weights[i] / sumW
		if w < mins[i] {
			w = mins[i]
		}
		out[i] = w
		used += w
	}
	rest := avail - used
	if rest < mins[n-1] {
		rest = mins[n-1]
	}
	out[n-1] = rest
	return out
}

// tableColumn describes a column header.
type tableColumn struct {
	Title  string
	Weight int
	Min    int
}

// tableRowCell is one cell in a custom table.
type tableRowCell struct {
	Text string
	Kind statusKind // if != None, apply status coloring
	Bold bool
	Mute bool
}

// tableDataRow is a navigable data row (or section divider).
type tableDataRow struct {
	Cells   []tableRowCell
	Section bool
}

// styledTableModel is a keyboard-navigable table with per-cell styling.
type styledTableModel struct {
	cols     []tableColumn
	rows     []tableDataRow
	cursor   int
	width    int
	height   int
	colWidth []int
}

func newStyledTable(cols []tableColumn, rows []tableDataRow, width, maxHeight int) styledTableModel {
	m := styledTableModel{
		cols:   cols,
		rows:   rows,
		width:  width,
		height: maxHeight,
	}
	if m.height < 5 {
		m.height = 14
	}
	m.recalc()
	m.cursor = m.firstSelectable(0)
	return m
}

func (m *styledTableModel) recalc() {
	weights := make([]int, len(m.cols))
	mins := make([]int, len(m.cols))
	for i, c := range m.cols {
		weights[i] = c.Weight
		if weights[i] < 1 {
			weights[i] = 1
		}
		mins[i] = c.Min
		if mins[i] < 1 {
			mins[i] = 3
		}
	}
	inner := m.width - 6
	if inner < 24 {
		inner = 48
	}
	m.colWidth = allocColumnWidths(inner, weights, mins)
}

func (m *styledTableModel) SetSize(w, h int) {
	m.width = w
	if h > 0 {
		// h is full body budget; header uses ~2 lines
		m.height = h
		if m.height < 5 {
			m.height = 5
		}
	}
	m.recalc()
}

func (m *styledTableModel) CursorUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.rows[i].Section {
			m.cursor = i
			return
		}
	}
}

func (m *styledTableModel) CursorDown() {
	for i := m.cursor + 1; i < len(m.rows); i++ {
		if !m.rows[i].Section {
			m.cursor = i
			return
		}
	}
}

func (m *styledTableModel) firstSelectable(from int) int {
	for i := from; i < len(m.rows); i++ {
		if !m.rows[i].Section {
			return i
		}
	}
	return 0
}

func renderCell(cell tableRowCell, width int) string {
	plain := padCell(cell.Text, width)
	content := strings.TrimRight(plain, " ")
	pad := width - runewidth.StringWidth(content)
	if pad < 0 {
		pad = 0
	}
	spaces := strings.Repeat(" ", pad)

	var styled string
	switch {
	case cell.Kind != statusKindNone:
		styled = styleStatusText(content)
	case cell.Bold:
		styled = lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(content)
	case cell.Mute:
		styled = mutedStyle.Render(content)
	default:
		styled = lipgloss.NewStyle().Foreground(colorText).Render(content)
	}
	return styled + spaces
}

func styleStatusText(display string) string {
	switch classifyStatus(display) {
	case statusKindOK:
		return successStyle.Render(display)
	case statusKindBad:
		return errorStyle.Render(display)
	case statusKindWarn:
		return warningStyle.Render(display)
	default:
		return mutedStyle.Render(display)
	}
}

func (m styledTableModel) View() string {
	if len(m.cols) == 0 {
		return mutedStyle.Render("（空）")
	}

	var b strings.Builder

	// Header
	var hdrParts []string
	for i, c := range m.cols {
		w := 8
		if i < len(m.colWidth) {
			w = m.colWidth[i]
		}
		hdrParts = append(hdrParts, padCell(c.Title, w))
	}
	headerLine := "  " + strings.Join(hdrParts, " ")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Render(headerLine))
	b.WriteString("\n")
	ruleW := runewidth.StringWidth(headerLine)
	if ruleW < 20 {
		ruleW = 20
	}
	if ruleW > 72 {
		ruleW = 72
	}
	b.WriteString(headerRuleStyle.Render(strings.Repeat("─", ruleW)))
	b.WriteString("\n")

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		row := m.rows[i]
		if row.Section {
			title := ""
			if len(row.Cells) > 0 {
				title = row.Cells[0].Text
			}
			b.WriteString(sectionHeaderStyle.Render("── " + title + " "))
			b.WriteString("\n")
			continue
		}

		selected := i == m.cursor
		prefix := "  "
		if selected {
			prefix = selectedBarStyle.Render("▸") + " "
		}

		var parts []string
		for ci := 0; ci < len(m.cols); ci++ {
			w := 8
			if ci < len(m.colWidth) {
				w = m.colWidth[ci]
			}
			cell := tableRowCell{}
			if ci < len(row.Cells) {
				cell = row.Cells[ci]
			}
			parts = append(parts, renderCell(cell, w))
		}
		line := prefix + strings.Join(parts, " ")
		if selected {
			line = lipgloss.NewStyle().Background(colorSurface).Render(line)
		}
		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m styledTableModel) visibleRange() (start, end int) {
	n := len(m.rows)
	if n == 0 {
		return 0, 0
	}
	h := m.height
	if h < 1 {
		h = 12
	}
	if n <= h {
		return 0, n
	}
	start = m.cursor - h/2
	if start < 0 {
		start = 0
	}
	end = start + h
	if end > n {
		end = n
		start = end - h
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

var sectionHeaderStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Bold(true)
