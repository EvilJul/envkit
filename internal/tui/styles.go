package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Adaptive palette (dark / light terminals) ──────────────────────────────

var (
	colorPrimary   = lipgloss.AdaptiveColor{Light: "#007aff", Dark: "#0a84ff"}
	colorSuccess   = lipgloss.AdaptiveColor{Light: "#34c759", Dark: "#30d158"}
	colorWarning   = lipgloss.AdaptiveColor{Light: "#ff9500", Dark: "#ff9f0a"}
	colorError     = lipgloss.AdaptiveColor{Light: "#ff3b30", Dark: "#ff453a"}
	colorMuted     = lipgloss.AdaptiveColor{Light: "#8e8e93", Dark: "#98989d"}
	colorBorder    = lipgloss.AdaptiveColor{Light: "#c7c7cc", Dark: "#3a3a3c"}
	colorText      = lipgloss.AdaptiveColor{Light: "#1c1c1e", Dark: "#f5f5f7"}
	colorTextDim   = lipgloss.AdaptiveColor{Light: "#636366", Dark: "#aeaeb2"}
	colorSurface   = lipgloss.AdaptiveColor{Light: "#f2f2f7", Dark: "#2c2c2e"}
	colorOnPrimary = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}
)

// ── Core styles ────────────────────────────────────────────────────────────

var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	warningStyle = lipgloss.NewStyle().Foreground(colorWarning)
	errorStyle   = lipgloss.NewStyle().Foreground(colorError)
	mutedStyle   = lipgloss.NewStyle().Foreground(colorMuted)

	selectedStyle = lipgloss.NewStyle().
			Foreground(colorOnPrimary).
			Background(colorPrimary).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	// Card / panel for confirmations and result boxes
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// Key badge: " enter "
	keyBadgeStyle = lipgloss.NewStyle().
			Foreground(colorOnPrimary).
			Background(colorPrimary).
			Padding(0, 1).
			MarginRight(1)

	keyDescStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			MarginRight(2)

	headerRuleStyle = lipgloss.NewStyle().
			Foreground(colorBorder)

	// List selection indicator bar
	selectedBarStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)
)

// KeyHint is a single shortcut shown in the footer.
type KeyHint struct {
	Key  string
	Desc string
}

// Common hint sets used across views.
var (
	hintsBack = []KeyHint{
		{Key: "esc/q", Desc: "返回"},
	}
	hintsNavSelect = []KeyHint{
		{Key: "↑/↓", Desc: "导航"},
		{Key: "enter", Desc: "选择"},
		{Key: "esc/q", Desc: "返回"},
	}
	hintsMainMenu = []KeyHint{
		{Key: "↑/↓", Desc: "导航"},
		{Key: "1-8", Desc: "直达"},
		{Key: "enter", Desc: "选择"},
		{Key: "q", Desc: "退出"},
	}
	hintsConfirmYN = []KeyHint{
		{Key: "y", Desc: "确认"},
		{Key: "n", Desc: "返回"},
		{Key: "esc/q", Desc: "返回"},
	}
	hintsEnterBack = []KeyHint{
		{Key: "enter", Desc: "返回"},
	}
	hintsMultiSelect = []KeyHint{
		{Key: "↑/↓", Desc: "导航"},
		{Key: "space", Desc: "切换"},
		{Key: "enter", Desc: "继续"},
		{Key: "esc/q", Desc: "返回"},
	}
	hintsUninstallSelect = []KeyHint{
		{Key: "↑/↓", Desc: "导航"},
		{Key: "space", Desc: "切换"},
		{Key: "a", Desc: "全选"},
		{Key: "enter", Desc: "继续"},
		{Key: "esc/q", Desc: "返回"},
	}
)

// RenderKeyHints builds a footer line of key badges.
func RenderKeyHints(hints []KeyHint) string {
	if len(hints) == 0 {
		return ""
	}
	var parts []string
	for _, h := range hints {
		parts = append(parts, keyBadgeStyle.Render(h.Key)+keyDescStyle.Render(h.Desc))
	}
	return helpStyle.Render(strings.Join(parts, ""))
}

// renderHelp keeps backward-compatible plain help text.
func renderHelp(keys string) string {
	return helpStyle.Render(keys)
}

// renderTitle builds the brand + page title header (no footer).
func renderTitle(title, subtitle string) string {
	brand := brandStyle.Render("EnvKit")
	page := titleStyle.Render(title)
	s := brand + "  " + page
	if subtitle != "" {
		s += "\n" + subtitleStyle.Render(subtitle)
	}
	return s
}

// RenderChrome composes a full-page layout: header + body + footer hints.
// width/height are optional (0 = no width constraint / no fill).
func RenderChrome(title, subtitle, body string, hints []KeyHint) string {
	header := renderTitle(title, subtitle)
	footer := RenderKeyHints(hints)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	// subtle separator under header
	rule := headerRuleStyle.Render(strings.Repeat("─", 40))
	b.WriteString(rule)
	b.WriteString("\n\n")
	b.WriteString(body)
	if footer != "" {
		b.WriteString("\n\n")
		b.WriteString(footer)
	}
	return b.String()
}

// RenderChromeSized is like RenderChrome but pads body to fill remaining height
// and places footer near the bottom when height > 0.
func RenderChromeSized(title, subtitle, body string, hints []KeyHint, width, height int) string {
	header := renderTitle(title, subtitle)
	footer := RenderKeyHints(hints)

	ruleW := 40
	if width > 8 {
		ruleW = width - 4
		if ruleW > 72 {
			ruleW = 72
		}
	}
	rule := headerRuleStyle.Render(strings.Repeat("─", ruleW))

	top := header + "\n" + rule + "\n\n"

	if height <= 0 {
		out := top + body
		if footer != "" {
			out += "\n\n" + footer
		}
		return out
	}

	topH := lipgloss.Height(top)
	footerH := 0
	if footer != "" {
		footerH = lipgloss.Height(footer) + 1 // gap above footer
	}
	bodyH := height - topH - footerH
	if bodyH < 1 {
		bodyH = 1
	}

	// Pad body so footer sits lower when there is spare vertical space.
	bodyBlock := body
	contentH := lipgloss.Height(body)
	if contentH < bodyH {
		bodyBlock = body + strings.Repeat("\n", bodyH-contentH)
	}

	if footer == "" {
		return top + bodyBlock
	}
	return top + bodyBlock + "\n" + footer
}

func backKeyHint() string {
	return RenderKeyHints(hintsBack)
}

// ── Step indicator (multi-step wizards) ────────────────────────────────────

// RenderStepIndicator draws a horizontal step bar.
// current is 0-based; steps are labels for each phase.
//
//	[1 选择] ─ [2 确认] ─ [3 执行] ─ [4 完成]
func RenderStepIndicator(steps []string, current int) string {
	if len(steps) == 0 {
		return ""
	}
	doneStyle := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	currStyle := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	todoStyle := lipgloss.NewStyle().Foreground(colorMuted)
	sepDone := lipgloss.NewStyle().Foreground(colorSuccess).Render(" ─ ")
	sepTodo := lipgloss.NewStyle().Foreground(colorBorder).Render(" ─ ")

	var parts []string
	for i, label := range steps {
		num := i + 1
		text := ""
		switch {
		case i < current:
			text = doneStyle.Render(fmt.Sprintf("✓ %s", label))
		case i == current:
			text = currStyle.Render(fmt.Sprintf("● %d %s", num, label))
		default:
			text = todoStyle.Render(fmt.Sprintf("○ %d %s", num, label))
		}
		parts = append(parts, text)
		if i < len(steps)-1 {
			if i < current {
				parts = append(parts, sepDone)
			} else {
				parts = append(parts, sepTodo)
			}
		}
	}
	return strings.Join(parts, "")
}

// RenderBanner wraps a message in a colored card for success / warning / error results.
func RenderBanner(kind string, title, body string) string {
	var border lipgloss.TerminalColor
	var titleStyled string
	switch kind {
	case "success":
		border = colorSuccess
		titleStyled = successStyle.Bold(true).Render("✓  " + title)
	case "warning":
		border = colorWarning
		titleStyled = warningStyle.Bold(true).Render("⚠  " + title)
	case "error":
		border = colorError
		titleStyled = errorStyle.Bold(true).Render("✗  " + title)
	default:
		border = colorPrimary
		titleStyled = brandStyle.Render(title)
	}
	content := titleStyled
	if body != "" {
		content += "\n\n" + body
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Render(content)
}

// RenderConfirmCard lists items inside a primary card.
func RenderConfirmCard(lines []string) string {
	if len(lines) == 0 {
		return boxStyle.Render(mutedStyle.Render("（无项目）"))
	}
	var b strings.Builder
	for i, line := range lines {
		b.WriteString("  •  ")
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
	return boxStyle.Render(b.String())
}
