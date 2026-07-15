package tui

import (
	"fmt"
	"strconv"
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

// menuEntries is the ordered main menu (index 0 → key "1").
func menuEntries() []menuItem {
	return []menuItem{
		{viewDetect, "🔍  系统检测", "查看已安装的语言、工具与包管理器"},
		{viewList, "📋  组件列表", "支持安装的环境与工具及当前状态"},
		{viewInstall, "📦  安装向导", "多选组件并一键安装"},
		{viewInit, "📝  初始化配置", "从模板生成 dev-env.yaml"},
		{viewUninstall, "🗑  卸载组件", "卸载 EnvKit 已安装的组件"},
		{viewMirror, "🪞  镜像源配置", "配置 npm/pip/go/rust 等镜像"},
		{viewDocker, "🐳  Docker 管理", "启动/停止/列出数据库容器"},
		{viewMenu, "🚪  退出", "返回命令行"},
	}
}

func newRootModel() rootModel {
	entries := menuEntries()
	items := make([]list.Item, len(entries))
	for i, e := range entries {
		// Prefix with number for scannability (1–8)
		num := strconv.Itoa(i + 1)
		items[i] = menuItem{
			id:    e.id,
			title: fmt.Sprintf("%s  %s", num, e.title),
			desc:  e.desc,
		}
	}

	delegate := configureRootListDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return rootModel{menu: l, active: viewMenu}
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) enterView(id viewID, title string) (rootModel, tea.Cmd) {
	if id == viewMenu {
		m.quitting = true
		return m, tea.Quit
	}
	m.active = id
	m.pendingID = id
	m.transitioning = true
	m.loading = newLoadingModel("正在进入 " + title + "…")
	m.loading.SetSize(m.width, m.height)
	return m, tea.Batch(
		m.loading.Init(),
		tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
			return subViewReadyMsg{id: id}
		}),
	)
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// Reserve space for chrome (header ~4 + footer ~2)
		menuH := msg.Height - 8
		if menuH < 6 {
			menuH = 6
		}
		m.menu.SetWidth(msg.Width)
		m.menu.SetHeight(menuH)
		m.loading.SetSize(msg.Width, msg.Height)
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
			// Forward current size so sub-view can layout immediately
			if m.width > 0 {
				return m, tea.Batch(
					m.sub.Init(),
					func() tea.Msg {
						return tea.WindowSizeMsg{Width: m.width, Height: m.height}
					},
				)
			}
			return m, m.sub.Init()
		}
		return m, nil

	case tea.KeyMsg:
		// 任意层级均可 ctrl+c 退出
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.transitioning {
			return m, nil
		}
		if m.sub != nil {
			switch msg.String() {
			case "esc", "q":
				// 进行中禁止丢弃子视图
				if b, ok := m.sub.(busyView); ok && b.Busy() {
					return m, nil
				}
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
				return m.enterView(item.id, stripMenuNumber(item.title))
			}
		case "1", "2", "3", "4", "5", "6", "7", "8":
			// Number shortcut: 1-based index into menu
			idx, _ := strconv.Atoi(msg.String())
			idx-- // 0-based
			if idx >= 0 && idx < len(m.menu.Items()) {
				m.menu.Select(idx)
				if item, ok := m.menu.SelectedItem().(menuItem); ok {
					return m.enterView(item.id, stripMenuNumber(item.title))
				}
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

// stripMenuNumber removes leading "N  " from display titles for loading text.
func stripMenuNumber(title string) string {
	// Titles look like "1  🔍  系统检测"
	for i, r := range title {
		if r == ' ' || (r >= '0' && r <= '9') {
			continue
		}
		// skip spaces after number
		rest := title[i:]
		for len(rest) > 0 && rest[0] == ' ' {
			rest = rest[1:]
		}
		return rest
	}
	return title
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
	body := m.menu.View()
	return RenderChromeSized(
		"主菜单",
		"选择功能后按 Enter 进入，数字键快速直达",
		body,
		hintsMainMenu,
		m.width,
		m.height,
	)
}
