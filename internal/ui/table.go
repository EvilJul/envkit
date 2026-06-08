package ui

import (
	"fmt"
	"strings"
)

// Table 表格结构
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable 创建新的表格
func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		headers: headers,
		rows:    [][]string{},
		widths:  widths,
	}
}

// AddRow 添加行
func (t *Table) AddRow(cols ...string) {
	if len(cols) != len(t.headers) {
		return
	}

	// 更新列宽
	for i, col := range cols {
		if len(col) > t.widths[i] {
			t.widths[i] = len(col)
		}
	}

	t.rows = append(t.rows, cols)
}

// Render 渲染表格
func (t *Table) Render() {
	// 打印顶部边框
	t.printBorder("┌", "┬", "┐")

	// 打印表头
	t.printRow(t.headers, ColorBold)

	// 打印分隔线
	t.printBorder("├", "┼", "┤")

	// 打印数据行
	for _, row := range t.rows {
		t.printRow(row, "")
	}

	// 打印底部边框
	t.printBorder("└", "┴", "┘")
}

// printBorder 打印边框
func (t *Table) printBorder(left, middle, right string) {
	fmt.Print(left)
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print(middle)
		}
	}
	fmt.Println(right)
}

// printRow 打印行
func (t *Table) printRow(cols []string, color string) {
	fmt.Print("│")
	for i, col := range cols {
		if color != "" {
			fmt.Printf(" %s%-*s%s ", color, t.widths[i], col, ColorReset)
		} else {
			fmt.Printf(" %-*s ", t.widths[i], col)
		}
		fmt.Print("│")
	}
	fmt.Println()
}

// PrintHeader 打印标题
func PrintHeader(title string) {
	width := 60
	padding := (width - len(title) - 2) / 2

	fmt.Println()
	fmt.Println(strings.Repeat("═", width))
	fmt.Printf("%s%s%s%s\n",
		strings.Repeat(" ", padding),
		ColorBold+ColorCyan,
		title,
		ColorReset+strings.Repeat(" ", width-padding-len(title)))
	fmt.Println(strings.Repeat("═", width))
	fmt.Println()
}

// PrintSection 打印章节
func PrintSection(title string) {
	fmt.Printf("\n%s%s%s\n", ColorBold+ColorBlue, title, ColorReset)
	fmt.Println(strings.Repeat("─", 40))
}
