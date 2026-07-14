package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/config"
	prog "github.com/fusheng/envkit/internal/progress"
	"github.com/fusheng/envkit/internal/ui"
)

type installPhase int

const (
	installSelect installPhase = iota
	installConfirm
	installRunning
	installDone
)

type installItem struct {
	opt      service.InstallOption
	selected bool
}

func (i installItem) Title() string {
	mark := "[ ]"
	if i.selected {
		mark = "[x]"
	}
	cat := "工具"
	switch i.opt.Category {
	case "language":
		cat = "语言"
	case "database":
		cat = "数据库"
	}
	if i.opt.Version != "" {
		return fmt.Sprintf("%s %s (%s)", mark, i.opt.Name, cat)
	}
	return fmt.Sprintf("%s %s (%s)", mark, i.opt.Name, cat)
}

func (i installItem) FilterValue() string  { return i.opt.Key }
func (i installItem) Description() string { return i.opt.Key + " " + i.opt.Version }

type installWizard struct {
	phase          installPhase
	list           list.Model
	items          []installItem
	installProg    installProgressModel
	installCh      chan tea.Msg
	result         service.InstallResult
	cfg            *config.Config
	done           bool
	errMsg         string
	width          int
	height         int
}

func newInstallWizard() *installWizard {
	opts := service.InstallOptions()
	items := make([]installItem, len(opts))
	listItems := make([]list.Item, len(opts))
	for i, o := range opts {
		items[i] = installItem{opt: o}
		listItems[i] = items[i]
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(listItems, delegate, 40, 20)
	l.Title = "选择要安装的组件"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return &installWizard{
		phase: installSelect,
		list:  l,
		items: items,
	}
}

func (m *installWizard) Init() tea.Cmd {
	return nil
}

func (m *installWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 8)
		if m.phase == installRunning {
			m.installProg.onResize(msg.Width, msg.Height)
		}

	case installProgressEventMsg:
		if m.phase == installRunning {
			var cmds []tea.Cmd
			if cmd := m.installProg.applyEvent(msg.evt); cmd != nil {
				cmds = append(cmds, cmd)
			}
			cmds = append(cmds, listenInstallMsg(m.installCh))
			var ucmd tea.Cmd
			m.installProg, ucmd = m.installProg.Update(msg)
			cmds = append(cmds, ucmd)
			return m, tea.Batch(cmds...)
		}

	case installFinishedMsg:
		m.phase = installDone
		m.result = msg.result

	case tea.KeyMsg:
		switch m.phase {
		case installSelect:
			switch msg.String() {
			case " ":
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.items) {
					m.items[idx].selected = !m.items[idx].selected
					m.refreshList()
				}
			case "enter":
				var selected []service.InstallOption
				for _, it := range m.items {
					if it.selected {
						selected = append(selected, it.opt)
					}
				}
				if len(selected) == 0 {
					m.errMsg = "请至少选择一个组件 (空格切换选中)"
					return m, nil
				}
				m.cfg = service.BuildConfigFromOptions(selected)
				m.phase = installConfirm
				m.errMsg = ""
			case "esc", "q":
				m.done = true
			}
		case installConfirm:
			switch msg.String() {
			case "y", "Y":
				m.phase = installRunning
				m.installProg = newInstallProgressModel(service.CountInstallSteps(m.cfg))
				if m.width > 0 {
					m.installProg.onResize(m.width, m.height)
				}
				m.installCh = make(chan tea.Msg, 256)
				return m, tea.Batch(m.installProg.Init(), m.beginInstall(), listenInstallMsg(m.installCh))
			case "n", "N", "esc", "q":
				if msg.String() == "esc" || msg.String() == "q" {
					m.done = true
				} else {
					m.phase = installSelect
				}
			}
		case installDone:
			switch msg.String() {
			case "enter", "esc", "q":
				m.done = true
			}
		}
	}

	var cmd tea.Cmd
	if m.phase == installSelect {
		m.list, cmd = m.list.Update(msg)
	}
	if m.phase == installRunning {
		var ucmd tea.Cmd
		m.installProg, ucmd = m.installProg.Update(msg)
		cmd = tea.Batch(cmd, ucmd)
	}
	return m, cmd
}

func (m *installWizard) refreshList() {
	items := make([]list.Item, len(m.items))
	for i, it := range m.items {
		items[i] = it
	}
	m.list.SetItems(items)
}

type installFinishedMsg struct {
	result service.InstallResult
}

func listenInstallMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return installFinishedMsg{}
		}
		return msg
	}
}

func (m *installWizard) beginInstall() tea.Cmd {
	cfg := m.cfg
	ch := m.installCh
	return func() tea.Msg {
		go func() {
			rep := &teaProgressReporter{ch: ch}
			oldRep := prog.SetReporter(rep)
			defer prog.SetReporter(oldRep)
			ui.SetRenderer(ui.NewSilentRenderer())
			defer ui.ResetRenderer()
			result := service.RunInstallation(cfg)
			ch <- installFinishedMsg{result: result}
			close(ch)
		}()
		return nil
	}
}

func (m *installWizard) View() string {
	switch m.phase {
	case installSelect:
		var b strings.Builder
		b.WriteString(renderTitle("安装向导", "空格切换选中，Enter 继续"))
		if m.errMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.errMsg))
		}
		b.WriteString("\n")
		b.WriteString(m.list.View())
		b.WriteString("\n")
		b.WriteString(backKeyHint())
		return b.String()

	case installConfirm:
		var lines []string
		for _, lang := range m.cfg.Languages {
			lines = append(lines, fmt.Sprintf("  • %s %s", lang.Name, lang.Version))
		}
		for _, tool := range m.cfg.Tools {
			lines = append(lines, fmt.Sprintf("  • %s", tool))
		}
		for _, db := range m.cfg.Databases {
			lines = append(lines, fmt.Sprintf("  • %s %s (Docker)", db.Name, db.Version))
		}
		body := strings.Join(lines, "\n")
		return renderTitle("确认安装", "") + "\n" +
			boxStyle.Render(body) + "\n\n" +
			"按 Y 开始安装，N 返回\n" +
			backKeyHint()

	case installRunning:
		return m.installProg.View()

	case installDone:
		var msg string
		if len(m.result.FailedComponents) > 0 {
			msg = warningStyle.Render("部分组件安装失败:")
			for _, c := range m.result.FailedComponents {
				msg += "\n  • " + errorStyle.Render(c)
			}
			if len(m.result.Succeeded) > 0 {
				msg += "\n\n" + successStyle.Render("已成功:")
				for _, c := range m.result.Succeeded {
					msg += "\n  • " + c
				}
			}
		} else {
			msg = successStyle.Render("全部安装成功！")
			for _, c := range m.result.Succeeded {
				msg += "\n  • " + c
			}
		}
		msg += "\n\n" + mutedStyle.Render("若终端找不到命令，请 source ~/.bashrc 或重开终端")
		return renderTitle("安装结果", "") + "\n\n" +
			boxStyle.Render(msg) + "\n\n" +
			renderHelp("Enter 返回")
	}
	return ""
}

func (m *installWizard) Done() bool { return m.done }

// Busy 安装运行中
func (m *installWizard) Busy() bool { return m.phase == installRunning }
