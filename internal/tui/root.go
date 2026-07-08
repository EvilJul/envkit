package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type viewID int

const (
	viewMenu viewID = iota
	viewDetect
	viewList
	viewInstall
	viewInit
	viewUninstall
	viewMirror
	viewDocker
)

type menuItem struct {
	id          viewID
	title, desc string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) FilterValue() string { return i.title }
func (i menuItem) Description() string { return i.desc }

type subViewReadyMsg struct {
	id viewID
}

type rootModel struct {
	menu          list.Model
	active        viewID
	sub           tea.Model
	loading       loadingModel
	transitioning bool
	pendingID     viewID
	width         int
	height        int
	quitting      bool
}

func newRootModel() rootModel {
	items := []list.Item{
		menuItem{viewDetect, "系统检测", "查看已安装的语言、工具与包管理器"},
		menuItem{viewList, "组件列表", "支持安装的环境与工具及当前状态"},
		menuItem{viewInstall, "安装向导", "多选组件并一键安装"},
		menuItem{viewInit, "初始化配置", "从模板生成 dev-env.yaml"},
		menuItem{viewUninstall, "卸载组件", "卸载 EnvKit 已安装的组件"},
		menuItem{viewMirror, "镜像源配置", "配置 npm/pip/go/rust 等镜像"},
		menuItem{viewDocker, "Docker 管理", "启动/停止/列出数据库容器"},
		menuItem{viewMenu, "退出", "返回命令行"},
	}

	delegate := configureRootListDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "EnvKit 开发环境管理"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return rootModel{menu: l, active: viewMenu}
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.menu.SetWidth(msg.Width)
		m.menu.SetHeight(msg.Height - 4)
		if m.sub != nil {
			var cmd tea.Cmd
			m.sub, cmd = m.sub.Update(msg)
			return m, cmd
		}
		return m, nil

	case subViewReadyMsg:
		m.transitioning = false
		m.sub = m.spawnSub(msg.id)
		if m.sub != nil {
			return m, m.sub.Init()
		}
		return m, nil

	case tea.KeyMsg:
		if m.transitioning {
			return m, nil
		}
		if m.sub != nil {
			switch msg.String() {
			case "esc", "q":
				m.sub = nil
				m.active = viewMenu
				return m, nil
			}
			var cmd tea.Cmd
			m.sub, cmd = m.sub.Update(msg)
			if done, ok := m.sub.(quitter); ok && done.Done() {
				m.sub = nil
				m.active = viewMenu
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.menu.SelectedItem().(menuItem); ok {
				if item.id == viewMenu {
					m.quitting = true
					return m, tea.Quit
				}
				m.active = item.id
				m.pendingID = item.id
				m.transitioning = true
				m.loading = newLoadingModel("正在进入 " + item.title + "…")
				return m, tea.Batch(
					m.loading.Init(),
					tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
						return subViewReadyMsg{id: item.id}
					}),
				)
			}
		}
	}

	if m.transitioning {
		var lcmd tea.Cmd
		m.loading, lcmd = m.loading.Update(msg)
		return m, lcmd
	}

	if m.sub != nil {
		var cmd tea.Cmd
		m.sub, cmd = m.sub.Update(msg)
		if done, ok := m.sub.(quitter); ok && done.Done() {
			m.sub = nil
			m.active = viewMenu
		}
		return m, cmd
	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m rootModel) spawnSub(id viewID) tea.Model {
	switch id {
	case viewDetect:
		return newDetectView()
	case viewList:
		return newListView()
	case viewInstall:
		return newInstallWizard()
	case viewInit:
		return newInitWizard()
	case viewUninstall:
		return newUninstallWizard()
	case viewMirror:
		return newMirrorWizard("")
	case viewDocker:
		return newDockerMenu()
	}
	return nil
}

func (m rootModel) View() string {
	if m.quitting {
		return ""
	}
	if m.transitioning {
		return m.loading.View()
	}
	if m.sub != nil {
		return m.sub.View()
	}
	var s string
	s += renderTitle("主菜单", "选择功能后按 Enter 进入，q 退出")
	s += "\n" + m.menu.View()
	s += "\n" + renderHelp("↑/↓ 导航  enter 选择  q 退出")
	return s
}

type quitter interface {
	Done() bool
}

func backKeyHint() string {
	return renderHelp("esc/q 返回主菜单")
}
