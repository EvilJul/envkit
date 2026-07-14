package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	prog "github.com/fusheng/envkit/internal/progress"
)

const maxInstallLogLines = 80

// installProgressEventMsg TUI 安装进度事件
type installProgressEventMsg struct {
	evt prog.Event
}

// teaProgressReporter 将 progress.Event 转为 tea.Msg
type teaProgressReporter struct {
	ch chan<- tea.Msg
}

func (t *teaProgressReporter) Report(e prog.Event) {
	if t == nil || t.ch == nil {
		return
	}
	// 阻塞发送，避免丢弃 error/done 事件导致 TUI 误报成功
	t.ch <- installProgressEventMsg{evt: e}
}

// installProgressModel 安装进度视图
type installProgressModel struct {
	progressBar progress.Model
	spinner     spinner.Model
	viewport    viewport.Model
	logLines    []string
	currentTask string
	currentPct  float64
	totalSteps  int
	width       int
	height      int
	ready       bool
}

func newInstallProgressModel(totalSteps int) installProgressModel {
	if totalSteps < 1 {
		totalSteps = 1
	}
	p := progress.New(progress.WithGradient("#007aff", "#34c759"))
	p.ShowPercentage = true
	s := spinner.New()
	s.Spinner = spinner.Points
	return installProgressModel{
		progressBar: p,
		spinner:     s,
		viewport:    viewport.New(60, 8),
		totalSteps:  totalSteps,
		logLines:    []string{"准备安装..."},
	}
}

func (m installProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.progressBar.Init())
}

func (m *installProgressModel) onResize(w, h int) {
	m.width, m.height = w, h
	barW := w - 8
	if barW < 20 {
		barW = 20
	}
	m.progressBar.Width = barW
	m.viewport.Width = w - 4
	m.viewport.Height = minInt(h-12, 10)
	if len(m.logLines) > 0 {
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))
		m.viewport.GotoBottom()
	}
	m.ready = true
}

func (m *installProgressModel) appendLog(line string) {
	m.logLines = append(m.logLines, line)
	if len(m.logLines) > maxInstallLogLines {
		m.logLines = m.logLines[len(m.logLines)-maxInstallLogLines:]
	}
	m.viewport.SetContent(strings.Join(m.logLines, "\n"))
	m.viewport.GotoBottom()
}

func (m *installProgressModel) applyEvent(evt prog.Event) tea.Cmd {
	m.currentTask = evt.TaskID
	m.currentPct = evt.Percent / 100
	if evt.Message != "" {
		m.appendLog(evt.Message)
	}
	return m.progressBar.SetPercent(m.currentPct)
}

func (m installProgressModel) Update(msg tea.Msg) (installProgressModel, tea.Cmd) {
	var cmds []tea.Cmd
	var scmd tea.Cmd
	m.spinner, scmd = m.spinner.Update(msg)
	cmds = append(cmds, scmd)
	var vcmd tea.Cmd
	m.viewport, vcmd = m.viewport.Update(msg)
	cmds = append(cmds, vcmd)
	var pcmd tea.Cmd
	pModel, pcmd := m.progressBar.Update(msg)
	if pb, ok := pModel.(progress.Model); ok {
		m.progressBar = pb
	}
	cmds = append(cmds, pcmd)
	return m, tea.Batch(cmds...)
}

func (m installProgressModel) View() string {
	if !m.ready && m.width == 0 {
		return renderTitle("正在安装", "初始化...") + "\n"
	}
	task := m.currentTask
	if task == "" {
		task = "install"
	}
	header := fmt.Sprintf("%s  %s", m.spinner.View(), lipgloss.NewStyle().Foreground(colorPrimary).Render(task))
	bar := m.progressBar.View()
	logBox := boxStyle.Width(m.width - 4).Render(m.viewport.View())
	return renderTitle("正在安装", fmt.Sprintf("共 %d 个组件", m.totalSteps)) + "\n\n" +
		header + "\n" + bar + "\n\n" + logBox + "\n"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// configureRootListDelegate 配置主菜单 list 样式
func configureRootListDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#007aff"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#8e8e93"))
	return delegate
}
