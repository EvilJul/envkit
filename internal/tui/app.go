package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// quitter 子视图结束标志
type quitter interface {
	Done() bool
}

// busyView 安装等进行中时禁止根菜单 esc 中断
type busyView interface {
	Busy() bool
}

// doneQuitModel 包装独立向导：Done() 后自动 tea.Quit
type doneQuitModel struct {
	inner tea.Model
}

func (m doneQuitModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m doneQuitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	if q, ok := m.inner.(quitter); ok && q.Done() {
		return m, tea.Batch(cmd, tea.Quit)
	}
	return m, cmd
}

func (m doneQuitModel) View() string {
	return m.inner.View()
}

func runStandalone(m tea.Model) error {
	p := tea.NewProgram(doneQuitModel{inner: m}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Run 启动主菜单 TUI
func Run() error {
	m := newRootModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunInitWizard 启动 init 向导
func RunInitWizard() error {
	return runStandalone(newInitWizard())
}

// RunInstallWizard 启动 install 向导
func RunInstallWizard() error {
	return runStandalone(newInstallWizard())
}

// RunUninstallWizard 启动 uninstall 向导
func RunUninstallWizard() error {
	return runStandalone(newUninstallWizard())
}

// RunMirrorWizard 启动 mirror 向导（可预选语言）
func RunMirrorWizard(language string) error {
	return runStandalone(newMirrorWizard(language))
}

// RunDockerMenu 启动 docker 管理菜单
func RunDockerMenu() error {
	return runStandalone(newDockerMenu())
}
