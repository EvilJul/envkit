package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/ui"
)

type uninstallPhase int

const (
	uninstallSelect uninstallPhase = iota
	uninstallConfirm
	uninstallRunning
	uninstallDone
)

type uninstallItem struct {
	item     service.ManifestItem
	selected bool
}

func (i uninstallItem) Title() string {
	mark := "[ ]"
	if i.selected {
		mark = "[x]"
	}
	return fmt.Sprintf("%s %s (%s)", mark, i.item.Name, i.item.Type)
}

func (i uninstallItem) FilterValue() string  { return i.item.Name }
func (i uninstallItem) Description() string { return i.item.Version + " @ " + i.item.InstalledAt }

type uninstallWizard struct {
	phase      uninstallPhase
	list       list.Model
	items      []uninstallItem
	selected   []string
	loading    loadingModel
	ready      bool
	done       bool
	errMsg     string
	finishErrs []string
	width      int
	height     int
}

func newUninstallWizard() *uninstallWizard {
	return &uninstallWizard{
		loading: newLoadingModel("正在加载清单…"),
		phase:   uninstallSelect,
	}
}

func (m *uninstallWizard) Init() tea.Cmd {
	return tea.Batch(m.loading.Init(), func() tea.Msg {
		items := []uninstallItem{}
		manifest, err := service.LoadManifestItems()
		if err != nil {
			return uninstallLoadedMsg{err: err.Error()}
		}
		for _, it := range manifest {
			items = append(items, uninstallItem{item: it})
		}
		return uninstallLoadedMsg{items: items}
	})
}

type uninstallLoadedMsg struct {
	items []uninstallItem
	err   string
}

func (m *uninstallWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.ready {
			m.list.SetWidth(msg.Width)
			m.list.SetHeight(msg.Height - 8)
		}
	case uninstallLoadedMsg:
		if msg.err != "" {
			m.errMsg = msg.err
			m.ready = true
			m.items = nil
			return m, nil
		}
		m.items = msg.items
		listItems := make([]list.Item, len(m.items))
		for i, it := range m.items {
			listItems[i] = it
		}
		delegate := list.NewDefaultDelegate()
		l := list.New(listItems, delegate, 40, 16)
		l.Title = "选择要卸载的组件"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		if m.width > 0 {
			l.SetWidth(m.width)
			l.SetHeight(m.height - 8)
		}
		m.list = l
		m.ready = true
	case uninstallFinishedMsg:
		m.phase = uninstallDone
		m.finishErrs = msg.errs
	case tea.KeyMsg:
		switch m.phase {
		case uninstallSelect:
			switch msg.String() {
			case " ":
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.items) {
					m.items[idx].selected = !m.items[idx].selected
					m.refreshList()
				}
				return m, nil
			case "a":
				for i := range m.items {
					m.items[i].selected = true
				}
				m.refreshList()
			case "enter":
				var names []string
				for _, it := range m.items {
					if it.selected {
						names = append(names, it.item.Name)
					}
				}
				if len(names) == 0 {
					m.errMsg = "请至少选择一个组件"
					return m, nil
				}
				m.selected = names
				m.phase = uninstallConfirm
				m.errMsg = ""
			case "esc", "q":
				m.done = true
			}
		case uninstallConfirm:
			switch msg.String() {
			case "y", "Y":
				m.phase = uninstallRunning
				m.loading = newLoadingModel("正在卸载组件…")
				return m, tea.Batch(m.loading.Init(), m.runUninstall())
			case "n", "N", "esc", "q":
				if msg.String() == "esc" || msg.String() == "q" {
					m.done = true
				} else {
					m.phase = uninstallSelect
				}
			}
		case uninstallDone:
			switch msg.String() {
			case "enter", "esc", "q":
				m.done = true
			}
		}
	}
	var cmds []tea.Cmd
	if !m.ready {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		cmds = append(cmds, lcmd)
	} else if m.phase == uninstallSelect {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.phase == uninstallRunning {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		cmds = append(cmds, lcmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *uninstallWizard) refreshList() {
	items := make([]list.Item, len(m.items))
	for i, it := range m.items {
		items[i] = it
	}
	m.list.SetItems(items)
}

type uninstallFinishedMsg struct {
	errs []string
}

func (m *uninstallWizard) runUninstall() tea.Cmd {
	names := append([]string(nil), m.selected...)
	return func() tea.Msg {
		ui.SetRenderer(ui.NewSilentRenderer())
		defer ui.ResetRenderer()
		var errs []string
		for _, name := range names {
			if err := service.UninstallComponent(name); err != nil {
				errs = append(errs, err.Error())
			}
		}
		return uninstallFinishedMsg{errs: errs}
	}
}

func (m *uninstallWizard) View() string {
	if !m.ready {
		return m.loading.View()
	}
	if m.errMsg != "" && len(m.items) == 0 {
		return renderTitle("卸载组件", "") + "\n" +
			errorStyle.Render(m.errMsg) + "\n" +
			backKeyHint()
	}
	if len(m.items) == 0 {
		return renderTitle("卸载组件", "") + "\n" +
			warningStyle.Render("清单中没有通过 EnvKit 安装的组件。") + "\n" +
			backKeyHint()
	}
	switch m.phase {
	case uninstallSelect:
		var b strings.Builder
		b.WriteString(renderTitle("卸载组件", "空格切换，a 全选，Enter 继续"))
		if m.errMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.errMsg))
		}
		b.WriteString("\n")
		b.WriteString(m.list.View())
		b.WriteString("\n")
		b.WriteString(renderHelp("esc/q 返回"))
		return b.String()
	case uninstallConfirm:
		lines := make([]string, len(m.selected))
		for i, n := range m.selected {
			lines[i] = "  • " + errorStyle.Render(n)
		}
		return renderTitle("确认卸载", "") + "\n" +
			boxStyle.Render(strings.Join(lines, "\n")) + "\n\n" +
			"按 Y 确认卸载，N 返回"
	case uninstallRunning:
		return m.loading.View()
	case uninstallDone:
		msg := successStyle.Render("所选组件已卸载。")
		if len(m.finishErrs) > 0 {
			msg = warningStyle.Render("部分卸载失败:")
			for _, e := range m.finishErrs {
				msg += "\n  • " + errorStyle.Render(e)
			}
		}
		return renderTitle("卸载完成", "") + "\n" +
			boxStyle.Render(msg) + "\n" +
			renderHelp("Enter 返回")
	}
	return ""
}

func (m *uninstallWizard) Done() bool { return m.done }

func (m *uninstallWizard) Busy() bool {
	return m.phase == uninstallRunning || !m.ready
}
