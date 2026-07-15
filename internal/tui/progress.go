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
	t.ch <- installProgressEventMsg{evt: e}
}

type logLevel int

const (
	logInfo logLevel = iota
	logSuccess
	logWarn
	logError
)

type coloredLogLine struct {
	text  string
	level logLevel
}

// installProgressModel 安装进度视图
type installProgressModel struct {
	progressBar progress.Model
	spinner     spinner.Model
	viewport    viewport.Model
	logLines    []coloredLogLine
	currentTask string
	currentPct  float64
	stepCurrent int
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
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)
	return installProgressModel{
		progressBar: p,
		spinner:     s,
		viewport:    viewport.New(60, 8),
		totalSteps:  totalSteps,
		logLines:    []coloredLogLine{{text: "准备安装…", level: logInfo}},
	}
}

func (m installProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.progressBar.Init())
}

func (m *installProgressModel) onResize(w, h int) {
	m.width, m.height = w, h
	barW := w - 10
	if barW < 20 {
		barW = 20
	}
	m.progressBar.Width = barW
	m.viewport.Width = w - 8
	// reserve header + steps + task + bar + chrome
	logH := h - 16
	if logH < 4 {
		logH = 4
	}
	if logH > 16 {
		logH = 16
	}
	m.viewport.Height = logH
	m.refreshViewport()
	m.ready = true
}

func (m *installProgressModel) refreshViewport() {
	var lines []string
	for _, l := range m.logLines {
		lines = append(lines, colorizeLogLine(l))
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

func colorizeLogLine(l coloredLogLine) string {
	prefix := mutedStyle.Render("· ")
	switch l.level {
	case logSuccess:
		return successStyle.Render("✓ ") + l.text
	case logWarn:
		return warningStyle.Render("⚠ ") + l.text
	case logError:
		return errorStyle.Render("✗ ") + l.text
	default:
		return prefix + mutedStyle.Render(l.text)
	}
}

func inferLogLevel(stage, message string) logLevel {
	s := strings.ToLower(stage + " " + message)
	switch {
	case strings.Contains(s, "fail") || strings.Contains(s, "error") ||
		strings.Contains(s, "失败") || strings.Contains(s, "错误") ||
		strings.Contains(s, "panic"):
		return logError
	case strings.Contains(s, "warn") || strings.Contains(s, "跳过") ||
		strings.Contains(s, "skip") || strings.Contains(s, "警告"):
		return logWarn
	case strings.Contains(s, "success") || strings.Contains(s, "done") ||
		strings.Contains(s, "完成") || strings.Contains(s, "成功") ||
		strings.Contains(s, "installed"):
		return logSuccess
	default:
		return logInfo
	}
}

func (m *installProgressModel) appendLog(line string, level logLevel) {
	m.logLines = append(m.logLines, coloredLogLine{text: line, level: level})
	if len(m.logLines) > maxInstallLogLines {
		m.logLines = m.logLines[len(m.logLines)-maxInstallLogLines:]
	}
	m.refreshViewport()
}

func (m *installProgressModel) applyEvent(evt prog.Event) tea.Cmd {
	if evt.TaskID != "" {
		m.currentTask = evt.TaskID
	}
	// Percent < 0 表示「仅日志、不改进度」（如 error）
	if evt.Percent >= 0 {
		newPct := evt.Percent / 100
		// 进度只增不减，避免异常事件把条拉回 0
		if newPct >= m.currentPct || (evt.Stage == "preparing" && evt.TaskID == "install") {
			m.currentPct = newPct
		}
		if m.totalSteps > 0 {
			completed := int(m.currentPct*float64(m.totalSteps) + 1e-9)
			switch evt.Stage {
			case "installing", "starting":
				// 正在做第 completed+1 步
				m.stepCurrent = completed + 1
			default:
				m.stepCurrent = completed
			}
			if m.stepCurrent < 1 {
				m.stepCurrent = 1
			}
			if m.stepCurrent > m.totalSteps {
				m.stepCurrent = m.totalSteps
			}
		}
	}
	if evt.Message != "" {
		m.appendLog(evt.Message, inferLogLevel(evt.Stage, evt.Message))
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
	steps := []string{"选择", "确认", "执行", "完成"}
	stepBar := RenderStepIndicator(steps, 2) // running = step 3 (0-based index 2)

	if !m.ready && m.width == 0 {
		body := stepBar + "\n\n" + m.spinner.View() + " " + subtitleStyle.Render("准备中…")
		return RenderChrome("正在安装", "初始化…", body, nil)
	}

	task := m.currentTask
	if task == "" {
		task = "install"
	}
	stepLabel := fmt.Sprintf("步骤 %d / %d", maxInt(m.stepCurrent, 1), m.totalSteps)
	if m.stepCurrent == 0 {
		stepLabel = fmt.Sprintf("共 %d 个组件", m.totalSteps)
	}

	taskLine := fmt.Sprintf("%s  %s  %s",
		m.spinner.View(),
		brandStyle.Render(task),
		mutedStyle.Render(stepLabel),
	)
	bar := m.progressBar.View()
	pct := int(m.currentPct * 100)
	pctLine := mutedStyle.Render(fmt.Sprintf("%d%%", pct))

	boxW := m.width - 6
	if boxW < 20 {
		boxW = 40
	}
	logLabel := mutedStyle.Render("日志")
	logBox := panelStyle.Width(boxW).Render(m.viewport.View())

	body := stepBar + "\n\n" +
		taskLine + "  " + pctLine + "\n\n" +
		bar + "\n\n" +
		logLabel + "\n" + logBox

	return RenderChrome("正在安装", fmt.Sprintf("共 %d 个组件", m.totalSteps), body, nil)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// configureRootListDelegate 配置主菜单 list 样式（现代极简选中态）
func configureRootListDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorPrimary).
		Foreground(colorPrimary).
		Bold(true).
		Padding(0, 0, 0, 1)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorPrimary).
		Foreground(colorMuted).
		Padding(0, 0, 0, 1)

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(colorText).
		Padding(0, 0, 0, 2)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 0, 0, 2)

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 0, 0, 2)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 0, 0, 2)

	return delegate
}

// configureWizardListDelegate 配置向导列表样式（与主菜单一致）
func configureWizardListDelegate() list.DefaultDelegate {
	return configureRootListDelegate()
}
