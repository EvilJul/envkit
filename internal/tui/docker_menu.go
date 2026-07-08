package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
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
	width      int
	height     int
}

func newDockerMenu() *dockerMenuModel {
	items := []list.Item{
		dockerMenuItem{actionList, "列出容器", "查看运行中的 Docker 容器"},
		dockerMenuItem{actionStart, "启动数据库", "启动 postgres/redis/mysql/mongodb"},
		dockerMenuItem{actionStop, "停止容器", "按容器名停止"},
		dockerMenuItem{actionRemove, "删除容器", "按容器名删除（可选删卷）"},
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = "Docker 管理"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

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
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 8)
	case dockerActionDoneMsg:
		m.phase = dockerResult
		if msg.err != nil {
			m.result = errorStyle.Render(msg.err.Error())
		} else {
			m.result = successStyle.Render(msg.message)
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
				m.focusIndex = (m.focusIndex + 1) % 2
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
		if err := service.DockerList(); err != nil {
			return dockerActionDoneMsg{err: err}
		}
		return dockerActionDoneMsg{message: "容器列表已输出到终端（TUI 外）"}
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
		var b strings.Builder
		b.WriteString(renderTitle("Docker 管理", ""))
		b.WriteString("\n")
		b.WriteString(m.list.View())
		b.WriteString("\n")
		b.WriteString(backKeyHint())
		return b.String()
	case dockerStartForm:
		return renderTitle("启动数据库", "") + "\n\n" +
			"数据库: " + m.dbInput.View() + "\n" +
			"版本:   " + m.verInput.View() + "\n\n" +
			renderHelp("Tab 切换字段  Enter 启动  esc 返回")
	case dockerStopForm:
		return renderTitle("停止容器", "") + "\n\n" +
			"容器名: " + m.nameInput.View() + "\n\n" +
			renderHelp("Enter 停止  esc 返回")
	case dockerRemoveForm:
		return renderTitle("删除容器", "") + "\n\n" +
			"容器名: " + m.nameInput.View() + "\n\n" +
			renderHelp("Enter 删除（不删卷）  esc 返回")
	case dockerRunning:
		return m.loading.View()
	case dockerResult:
		return renderTitle("操作结果", "") + "\n\n" + m.result + "\n" + renderHelp("Enter 返回")
	}
	return ""
}

func (m *dockerMenuModel) Done() bool { return m.done }
