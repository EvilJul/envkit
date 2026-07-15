package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// loadingModel 通用加载过渡（居中 spinner + 文案）
type loadingModel struct {
	spinner spinner.Model
	message string
	width   int
	height  int
}

func newLoadingModel(message string) loadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)
	return loadingModel{spinner: s, message: message}
}

func (m loadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m loadingModel) Update(msg tea.Msg) (loadingModel, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *loadingModel) SetSize(w, h int) {
	m.width, m.height = w, h
}

func (m loadingModel) View() string {
	line := m.spinner.View() + " " + subtitleStyle.Render(m.message)
	content := brandStyle.Render("EnvKit") + "\n\n" + line

	w, h := m.width, m.height
	if w <= 0 {
		w = 60
	}
	if h <= 0 {
		h = 12
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}
