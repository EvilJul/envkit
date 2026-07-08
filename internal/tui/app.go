package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run 启动主菜单 TUI
func Run() error {
	m := newRootModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunInitWizard 启动 init 向导
func RunInitWizard() error {
	m := newInitWizard()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunInstallWizard 启动 install 向导
func RunInstallWizard() error {
	m := newInstallWizard()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunUninstallWizard 启动 uninstall 向导
func RunUninstallWizard() error {
	m := newUninstallWizard()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunMirrorWizard 启动 mirror 向导（可预选语言）
func RunMirrorWizard(language string) error {
	m := newMirrorWizard(language)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunDockerMenu 启动 docker 管理菜单
func RunDockerMenu() error {
	m := newDockerMenu()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
