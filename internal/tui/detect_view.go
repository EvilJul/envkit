package tui

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
)

type detectView struct {
	table   styledTableModel
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
		m.loading.SetSize(msg.Width, msg.Height)
		if m.ready {
			m.table = buildDetectTableFromReport(m.report, m.width, m.height)
		}
	case detectLoadedMsg:
		m.report = msg.report
		m.table = buildDetectTableFromReport(msg.report, m.width, m.height)
		m.ready = true
	case tea.KeyMsg:
		if !m.ready {
			break
		}
		switch msg.String() {
		case "esc", "q":
			m.done = true
		case "up", "k":
			m.table.CursorUp()
		case "down", "j":
			m.table.CursorDown()
		}
	}
	var cmds []tea.Cmd
	if !m.ready {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		cmds = append(cmds, lcmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *detectView) View() string {
	if !m.ready {
		m.loading.SetSize(m.width, m.height)
		return m.loading.View()
	}
	return RenderChromeSized(
		"系统检测",
		"当前机器上的语言、工具与包管理器",
		m.table.View(),
		[]KeyHint{
			{Key: "↑/↓", Desc: "导航"},
			{Key: "esc/q", Desc: "返回"},
		},
		m.width,
		m.height,
	)
}

func (m *detectView) Done() bool { return m.done }

func buildDetectTableFromReport(report service.DetectReport, width, height int) styledTableModel {
	cols := []tableColumn{
		{Title: "名称", Weight: 2, Min: 8},
		{Title: "版本 / 值", Weight: 4, Min: 10},
		{Title: "状态", Weight: 2, Min: 8},
	}

	var data []tableDataRow

	// ── 系统 ──
	data = append(data, tableDataRow{Section: true, Cells: []tableRowCell{{Text: "系统"}}})
	data = append(data, tableDataRow{Cells: []tableRowCell{
		{Text: "OS", Bold: true},
		{Text: string(report.System.OS)},
		{Text: "—", Kind: statusKindMuted},
	}})
	data = append(data, tableDataRow{Cells: []tableRowCell{
		{Text: "架构", Bold: true},
		{Text: string(report.System.Architecture)},
		{Text: "—", Kind: statusKindMuted},
	}})
	if report.System.IsLinux() && report.System.Distribution != "" {
		data = append(data, tableDataRow{Cells: []tableRowCell{
			{Text: "发行版", Bold: true},
			{Text: report.System.Distribution},
			{Text: "—", Kind: statusKindMuted},
		}})
	}

	// ── 语言 ──
	langNames := sortedKeys(report.Languages)
	if len(langNames) > 0 {
		data = append(data, tableDataRow{Section: true, Cells: []tableRowCell{{Text: "语言"}}})
		for _, name := range langNames {
			tool := report.Languages[name]
			if tool == nil || !tool.Installed {
				continue
			}
			st := formatStatusDisplay(statusInstalled)
			data = append(data, tableDataRow{Cells: []tableRowCell{
				{Text: name, Bold: true},
				{Text: tool.Version, Mute: true},
				{Text: st, Kind: statusKindOK},
			}})
		}
	}

	// ── 工具 ──
	toolNames := sortedKeys(report.Tools)
	installedTools := 0
	for _, name := range toolNames {
		if t := report.Tools[name]; t != nil && t.Installed {
			installedTools++
		}
	}
	if installedTools > 0 {
		data = append(data, tableDataRow{Section: true, Cells: []tableRowCell{{Text: "工具"}}})
		for _, name := range toolNames {
			tool := report.Tools[name]
			if tool == nil || !tool.Installed {
				continue
			}
			st := formatStatusDisplay(statusInstalled)
			data = append(data, tableDataRow{Cells: []tableRowCell{
				{Text: name, Bold: true},
				{Text: tool.Version, Mute: true},
				{Text: st, Kind: statusKindOK},
			}})
		}
	}

	// ── 包管理 ──
	mgrNames := sortedKeys(report.Managers)
	installedMgr := 0
	for _, name := range mgrNames {
		if t := report.Managers[name]; t != nil && t.Installed {
			installedMgr++
		}
	}
	if installedMgr > 0 {
		data = append(data, tableDataRow{Section: true, Cells: []tableRowCell{{Text: "包管理"}}})
		for _, name := range mgrNames {
			tool := report.Managers[name]
			if tool == nil || !tool.Installed {
				continue
			}
			st := formatStatusDisplay(statusAvailable)
			data = append(data, tableDataRow{Cells: []tableRowCell{
				{Text: name, Bold: true},
				{Text: tool.Version, Mute: true},
				{Text: st, Kind: statusKindWarn},
			}})
		}
	}

	bodyH := height - 10
	if bodyH < 8 {
		bodyH = 14
	}
	return newStyledTable(cols, data, width, bodyH)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
