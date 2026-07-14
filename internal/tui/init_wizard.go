package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
)

type initWizard struct {
	list      list.Model
	done      bool
	saved     bool
	errMsg    string
	width     int
	height    int
}

type templateItem struct {
	index int
	info  string
}

func (t templateItem) Title() string       { return t.info }
func (t templateItem) FilterValue() string { return t.info }
func (t templateItem) Description() string { return fmt.Sprintf("模板 #%d", t.index) }

func newInitWizard() *initWizard {
	templates := service.TemplateList()
	items := make([]list.Item, len(templates))
	for i, t := range templates {
		items[i] = templateItem{
			index: i + 1,
			info:  fmt.Sprintf("%s — %s", t.Name, t.Description),
		}
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 40, 14)
	l.Title = "选择预设模板"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return &initWizard{list: l}
}

func (m *initWizard) Init() tea.Cmd { return nil }

func (m *initWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 8)
	case initSavedMsg:
		m.saved = true
		m.errMsg = successStyle.Render("已生成 dev-env.yaml，运行 envkit install -f dev-env.yaml 开始安装")
	case initErrorMsg:
		m.errMsg = errorStyle.Render(msg.err)
	case tea.KeyMsg:
		if m.saved {
			switch msg.String() {
			case "enter", "esc", "q":
				m.done = true
			}
			return m, nil
		}
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(templateItem); ok {
				return m, m.saveTemplate(item.index)
			}
		case "esc", "q":
			m.done = true
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type initSavedMsg struct{}
type initErrorMsg struct{ err string }

func (m *initWizard) saveTemplate(index int) tea.Cmd {
	return func() tea.Msg {
		if err := service.InitFromTemplate(index, "dev-env.yaml"); err != nil {
			return initErrorMsg{err: err.Error()}
		}
		return initSavedMsg{}
	}
}

func (m *initWizard) View() string {
	var b strings.Builder
	b.WriteString(renderTitle("初始化配置", "选择模板生成 dev-env.yaml"))
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(m.errMsg)
	}
	b.WriteString("\n")
	if !m.saved {
		b.WriteString(m.list.View())
	}
	b.WriteString("\n")
	if m.saved {
		b.WriteString(renderHelp("Enter 返回"))
	} else {
		b.WriteString(backKeyHint())
	}
	return b.String()
}

func (m *initWizard) Done() bool { return m.done }
