package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fusheng/envkit/internal/cli/service"
)

type listView struct {
	table   table.Model
	rows    []service.CatalogRow
	loading loadingModel
	ready   bool
	done    bool
	width   int
}

func newListView() *listView {
	return &listView{
		loading: newLoadingModel("正在加载组件列表…"),
	}
}

func (m *listView) Init() tea.Cmd {
	return tea.Batch(m.loading.Init(), func() tea.Msg {
		return listLoadedMsg{rows: service.ListCatalog()}
	})
}

type listLoadedMsg struct {
	rows []service.CatalogRow
}

func (m *listView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.ready {
			m.table = buildCatalogTable(m.rows, msg.Width)
		}
	case listLoadedMsg:
		m.rows = msg.rows
		m.table = buildCatalogTable(msg.rows, m.width)
		m.ready = true
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.done = true
		}
	}
	var cmds []tea.Cmd
	if !m.ready {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		cmds = append(cmds, lcmd)
	} else {
		var tcmd tea.Cmd
		m.table, tcmd = m.table.Update(msg)
		cmds = append(cmds, tcmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *listView) View() string {
	if !m.ready {
		return m.loading.View()
	}
	var b strings.Builder
	b.WriteString(renderTitle("组件列表", "支持安装的环境与工具"))
	b.WriteString("\n")
	b.WriteString(m.table.View())
	b.WriteString("\n")
	b.WriteString(backKeyHint())
	return b.String()
}

func (m *listView) Done() bool { return m.done }

func buildCatalogTable(rows []service.CatalogRow, width int) table.Model {
	columns := []table.Column{
		{Title: "名称", Width: 12},
		{Title: "版本", Width: 10},
		{Title: "状态", Width: 16},
		{Title: "说明", Width: 30},
	}
	var tableRows []table.Row
	for _, r := range rows {
		status := r.Status
		if status == "" {
			status = "—"
		} else if status != statusNotInstalled {
			status = tableCell("✓ "+status, 14)
		} else {
			status = "✗ " + statusNotInstalled
		}
		tableRows = append(tableRows, table.Row{
			r.Name,
			tableCell(r.Version, 8),
			status,
			tableCell(r.Description, 28),
		})
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(minInt(len(tableRows)+1, 22)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		BorderBottom(true).
		Bold(true)
	s.Cell = s.Cell.Padding(0, 1)
	t.SetStyles(s)
	if width > 0 {
		t.SetWidth(width - 4)
	}
	return t
}
