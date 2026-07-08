package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fusheng/envkit/internal/cli/service"
)

type detectView struct {
	table   table.Model
	report  service.DetectReport
	loading loadingModel
	ready   bool
	done    bool
	width   int
	height  int
}

func newDetectView() *detectView {
	return &detectView{
		loading: newLoadingModel("正在检测环境…"),
	}
}

func (m *detectView) Init() tea.Cmd {
	return tea.Batch(m.loading.Init(), func() tea.Msg {
		return detectLoadedMsg{report: service.DetectEnvironment()}
	})
}

type detectLoadedMsg struct {
	report service.DetectReport
}

func (m *detectView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.ready {
			m.table = buildDetectTableFromReport(m.report, m.width)
		}
	case detectLoadedMsg:
		m.report = msg.report
		m.table = buildDetectTableFromReport(msg.report, m.width)
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
	} else if m.table.Rows() != nil {
		var tcmd tea.Cmd
		m.table, tcmd = m.table.Update(msg)
		cmds = append(cmds, tcmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *detectView) View() string {
	if !m.ready {
		return m.loading.View()
	}
	var b strings.Builder
	b.WriteString(renderTitle("系统检测", ""))
	b.WriteString("\n")
	b.WriteString(m.table.View())
	b.WriteString("\n")
	b.WriteString(backKeyHint())
	return b.String()
}

func (m *detectView) Done() bool { return m.done }

func buildDetectTableFromReport(report service.DetectReport, width int) table.Model {
	columns := []table.Column{
		{Title: "类别", Width: 10},
		{Title: "名称", Width: 12},
		{Title: "版本", Width: 22},
		{Title: "状态", Width: 8},
	}
	var rows []table.Row
	rows = append(rows, table.Row{"系统", "OS", string(report.System.OS), "—"})
	rows = append(rows, table.Row{"系统", "架构", string(report.System.Architecture), "—"})
	if report.System.IsLinux() {
		rows = append(rows, table.Row{"系统", "发行版", report.System.Distribution, "—"})
	}
	for name, tool := range report.Languages {
		if tool.Installed {
			rows = append(rows, table.Row{"语言", name, tableCell(tool.Version, 18), statusInstalled})
		}
	}
	for name, tool := range report.Tools {
		if tool.Installed {
			rows = append(rows, table.Row{"工具", name, tableCell(tool.Version, 18), statusInstalled})
		}
	}
	for name, tool := range report.Managers {
		if tool.Installed {
			rows = append(rows, table.Row{"包管理", name, tableCell(tool.Version, 18), statusAvailable})
		}
	}
	return styleDetectTable(rows, columns, width)
}

func styleDetectTable(rows []table.Row, columns []table.Column, width int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(minInt(len(rows)+1, 20)),
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
