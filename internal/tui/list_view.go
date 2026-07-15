package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
)

type listView struct {
	table   styledTableModel
	rows    []service.CatalogRow
	loading loadingModel
	ready   bool
	done    bool
	width   int
	height  int
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
		m.width, m.height = msg.Width, msg.Height
		m.loading.SetSize(msg.Width, msg.Height)
		if m.ready {
			m.table = buildCatalogTable(m.rows, msg.Width, msg.Height)
		}
	case listLoadedMsg:
		m.rows = msg.rows
		m.table = buildCatalogTable(msg.rows, m.width, m.height)
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

func (m *listView) View() string {
	if !m.ready {
		return m.loading.View()
	}
	return RenderChromeSized(
		"组件列表",
		"支持安装的环境与工具及当前状态",
		m.table.View(),
		[]KeyHint{
			{Key: "↑/↓", Desc: "导航"},
			{Key: "esc/q", Desc: "返回"},
		},
		m.width,
		m.height,
	)
}

func (m *listView) Done() bool { return m.done }

func buildCatalogTable(rows []service.CatalogRow, width, height int) styledTableModel {
	cols := []tableColumn{
		{Title: "名称", Weight: 2, Min: 8},
		{Title: "版本", Weight: 2, Min: 6},
		{Title: "状态", Weight: 3, Min: 10},
		{Title: "说明", Weight: 5, Min: 10},
	}

	// Group by category with stable order
	order := []string{"language", "tool", "database"}
	labels := map[string]string{
		"language": "语言",
		"tool":     "工具",
		"database": "数据库",
	}
	grouped := map[string][]service.CatalogRow{}
	for _, r := range rows {
		cat := r.Category
		if cat == "" {
			cat = "tool"
		}
		grouped[cat] = append(grouped[cat], r)
	}

	var data []tableDataRow
	for _, cat := range order {
		items := grouped[cat]
		if len(items) == 0 {
			continue
		}
		data = append(data, tableDataRow{
			Section: true,
			Cells:   []tableRowCell{{Text: labels[cat]}},
		})
		for _, r := range items {
			st := formatStatusDisplay(r.Status)
			// databases often have empty status
			if r.Category == "database" && strings.TrimSpace(r.Status) == "" {
				st = formatStatusDisplay("")
			}
			kind := classifyStatus(st)
			data = append(data, tableDataRow{
				Cells: []tableRowCell{
					{Text: r.Name, Bold: true},
					{Text: r.Version, Mute: r.Version == ""},
					{Text: st, Kind: kind},
					{Text: r.Description, Mute: true},
				},
			})
		}
	}

	bodyH := height - 10
	if bodyH < 8 {
		bodyH = 14
	}
	return newStyledTable(cols, data, width, bodyH)
}
