package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/ui"
)

type dockerPhase int

const (
	dockerMenuPhase dockerPhase = iota
	dockerStartForm
	dockerStopForm
	dockerRemoveForm
	dockerRunning
	dockerResult
)

type dockerAction int

const (
	actionList dockerAction = iota
	actionStart
	actionStop
	actionRemove
)

type dockerMenuItem struct {
	action dockerAction
	title  string
	desc   string
}

func (d dockerMenuItem) Title() string       { return d.title }
func (d dockerMenuItem) FilterValue() string { return d.title }
func (d dockerMenuItem) Description() string { return d.desc }

type dockerMenuModel struct {
	phase      dockerPhase
	list       list.Model
	dbInput    textinput.Model
	verInput   textinput.Model
	nameInput  textinput.Model
	loading    loadingModel
	focusIndex int
	done       bool
	result     string
	resultOK   bool
	width      int
	height     int
}

func newDockerMenu() *dockerMenuModel {
	items := []list.Item{
		dockerMenuItem{actionList, "📋  列出容器", "查看运行中的 Docker 容器"},
		dockerMenuItem{actionStart, "▶  启动数据库", "启动 postgres/redis/mysql/mongodb"},
		dockerMenuItem{actionStop, "⏹  停止容器", "按容器名停止"},
		dockerMenuItem{actionRemove, "🗑  删除容器", "按容器名删除（可选删卷）"},
	}
	delegate := configureWizardListDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	db := textinput.New()
	db.Placeholder = "postgres | redis | mysql | mongodb"
	ver := textinput.New()
	ver.Placeholder = "版本，如 16 / 7 / 8.0"
	name := textinput.New()
	name.Placeholder = "容器名称"

	return &dockerMenuModel{
		list:      l,
		dbInput:   db,
		verInput:  ver,
		nameInput: name,
		phase:     dockerMenuPhase,
	}
}

func (m *dockerMenuModel) Init() tea.Cmd { return nil }

func (m *dockerMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.loading.SetSize(msg.Width, msg.Height)
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 8)
	case dockerActionDoneMsg:
		m.phase = dockerResult
		if msg.err != nil {
			m.result = msg.err.Error()
			m.resultOK = false
		} else {
			m.result = msg.message
			m.resultOK = true
		}
	case tea.KeyMsg:
		switch m.phase {
		case dockerMenuPhase:
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(dockerMenuItem); ok {
					switch item.action {
					case actionList:
						m.phase = dockerRunning
						m.loading = newLoadingModel("正在获取容器列表…")
						return m, tea.Batch(m.loading.Init(), m.runList())
					case actionStart:
						m.phase = dockerStartForm
						m.focusIndex = 0
						m.dbInput.Focus()
					case actionStop:
						m.phase = dockerStopForm
						m.nameInput.Focus()
					case actionRemove:
						m.phase = dockerRemoveForm
						m.nameInput.Focus()
					}
				}
			case "esc", "q":
				m.done = true
			}
		case dockerStartForm:
			switch msg.String() {
			case "tab", "down":
				m.focusIndex = (m.focusIndex + 1) % 2
				m.updateFocus()
			case "shift+tab", "up":
				m.focusIndex = (m.focusIndex - 1 + 2) % 2
				m.updateFocus()
			case "enter":
				m.phase = dockerRunning
				m.loading = newLoadingModel("正在启动数据库…")
				return m, tea.Batch(m.loading.Init(), m.runStart())
			case "esc", "q":
				m.phase = dockerMenuPhase
			default:
				var cmd tea.Cmd
				if m.focusIndex == 0 {
					m.dbInput, cmd = m.dbInput.Update(msg)
				} else {
					m.verInput, cmd = m.verInput.Update(msg)
				}
				return m, cmd
			}
		case dockerStopForm, dockerRemoveForm:
			switch msg.String() {
			case "enter":
				if m.phase == dockerStopForm {
					m.phase = dockerRunning
					m.loading = newLoadingModel("正在停止容器…")
					return m, tea.Batch(m.loading.Init(), m.runStop())
				}
				m.phase = dockerRunning
				m.loading = newLoadingModel("正在删除容器…")
				return m, tea.Batch(m.loading.Init(), m.runRemove(false))
			case "esc", "q":
				m.phase = dockerMenuPhase
			default:
				var cmd tea.Cmd
				m.nameInput, cmd = m.nameInput.Update(msg)
				return m, cmd
			}
		case dockerResult:
			switch msg.String() {
			case "enter", "esc", "q":
				m.phase = dockerMenuPhase
				m.result = ""
			}
		}
	}
	var cmd tea.Cmd
	if m.phase == dockerMenuPhase {
		m.list, cmd = m.list.Update(msg)
	} else if m.phase == dockerRunning {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		cmd = lcmd
	}
	return m, cmd
}

func (m *dockerMenuModel) updateFocus() {
	m.dbInput.Blur()
	m.verInput.Blur()
	if m.focusIndex == 0 {
		m.dbInput.Focus()
	} else {
		m.verInput.Focus()
	}
}

type dockerActionDoneMsg struct {
	message string
	err     error
}

func (m *dockerMenuModel) runList() tea.Cmd {
	return func() tea.Msg {
		ui.SetRenderer(ui.NewSilentRenderer())
		defer ui.ResetRenderer()
		rows, err := service.DockerListData()
		if err != nil {
			return dockerActionDoneMsg{err: err}
		}
		if len(rows) == 0 {
			return dockerActionDoneMsg{message: "（没有 envkit-* 容器）"}
		}
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%-24s %-24s %s\n", "NAMES", "STATUS", "PORTS"))
		for _, r := range rows {
			b.WriteString(fmt.Sprintf("%-24s %-24s %s\n", r.Name, r.Status, r.Ports))
		}
		return dockerActionDoneMsg{message: strings.TrimRight(b.String(), "\n")}
	}
}

func (m *dockerMenuModel) runStart() tea.Cmd {
	db := m.dbInput.Value()
	ver := m.verInput.Value()
	return func() tea.Msg {
		if db == "" || ver == "" {
			return dockerActionDoneMsg{err: fmt.Errorf("请填写数据库类型和版本")}
		}
		if err := service.DockerStart(db, ver); err != nil {
			return dockerActionDoneMsg{err: err}
		}
		return dockerActionDoneMsg{message: fmt.Sprintf("已启动 %s %s", db, ver)}
	}
}

func (m *dockerMenuModel) runStop() tea.Cmd {
	name := m.nameInput.Value()
	return func() tea.Msg {
		if name == "" {
			return dockerActionDoneMsg{err: fmt.Errorf("请填写容器名称")}
		}
		if err := service.DockerStop(name); err != nil {
			return dockerActionDoneMsg{err: err}
		}
		return dockerActionDoneMsg{message: "容器已停止: " + name}
	}
}

func (m *dockerMenuModel) runRemove(removeVolume bool) tea.Cmd {
	name := m.nameInput.Value()
	return func() tea.Msg {
		if name == "" {
			return dockerActionDoneMsg{err: fmt.Errorf("请填写容器名称")}
		}
		if err := service.DockerRemove(name, removeVolume); err != nil {
			return dockerActionDoneMsg{err: err}
		}
		return dockerActionDoneMsg{message: "容器已删除: " + name}
	}
}

func (m *dockerMenuModel) View() string {
	switch m.phase {
	case dockerMenuPhase:
		return RenderChrome("Docker 管理", "数据库容器快捷操作", m.list.View(), hintsNavSelect)
	case dockerStartForm:
		body := "数据库: " + m.dbInput.View() + "\n" +
			"版本:   " + m.verInput.View()
		return RenderChrome("启动数据库", "填写数据库类型与版本", body, []KeyHint{
			{Key: "tab", Desc: "切换字段"},
			{Key: "enter", Desc: "启动"},
			{Key: "esc", Desc: "返回"},
		})
	case dockerStopForm:
		body := "容器名: " + m.nameInput.View()
		return RenderChrome("停止容器", "", body, []KeyHint{
			{Key: "enter", Desc: "停止"},
			{Key: "esc", Desc: "返回"},
		})
	case dockerRemoveForm:
		body := "容器名: " + m.nameInput.View()
		return RenderChrome("删除容器", "默认不删除数据卷", body, []KeyHint{
			{Key: "enter", Desc: "删除"},
			{Key: "esc", Desc: "返回"},
		})
	case dockerRunning:
		m.loading.SetSize(m.width, m.height)
		return m.loading.View()
	case dockerResult:
		var body string
		if m.resultOK {
			body = RenderBanner("success", "操作成功", m.result)
		} else {
			body = RenderBanner("error", "操作失败", m.result)
		}
		return RenderChrome("操作结果", "", body, hintsEnterBack)
	}
	return ""
}

func (m *dockerMenuModel) Done() bool { return m.done }

func (m *dockerMenuModel) Busy() bool { return m.phase == dockerRunning }
