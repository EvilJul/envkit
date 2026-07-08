package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// loadingModel 通用加载过渡（spinner + 文案）
type loadingModel struct {
	spinner spinner.Model
	message string
}

func newLoadingModel(message string) loadingModel {
	s := spinner.New()
	s.Spinner = spinner.Line
	return loadingModel{spinner: s, message: message}
}

func (m loadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m loadingModel) Update(msg tea.Msg) (loadingModel, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m loadingModel) View() string {
	return renderTitle("EnvKit", m.message) + "\n\n" +
		m.spinner.View() + " " + subtitleStyle.Render(m.message) + "\n"
}

func renderLoadingOverlay(title, message string, s spinner.Model) string {
	return renderTitle(title, "") + "\n\n" +
		s.View() + " " + subtitleStyle.Render(message) + "\n"
}
