package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/mirror"
)

type mirrorPhase int

const (
	mirrorLang mirrorPhase = iota
	mirrorPick
	mirrorDone
)

type mirrorLangItem struct {
	lang string
}

func (m mirrorLangItem) Title() string       { return m.lang }
func (m mirrorLangItem) FilterValue() string { return m.lang }
func (m mirrorLangItem) Description() string { return "配置 " + m.lang + " 镜像源" }

type mirrorNameItem struct {
	name string
	url  string
}

func (m mirrorNameItem) Title() string       { return m.name }
func (m mirrorNameItem) FilterValue() string { return m.name }
func (m mirrorNameItem) Description() string { return m.url }

type mirrorWizard struct {
	phase    mirrorPhase
	list     list.Model
	language string
	done     bool
	errMsg   string
	okMsg    string
	width    int
	height   int
}

func newMirrorWizard(presetLang string) *mirrorWizard {
	m := &mirrorWizard{}
	if presetLang != "" {
		m.language = presetLang
		m.phase = mirrorPick
		m.list = buildMirrorList(presetLang)
	} else {
		m.phase = mirrorLang
		m.list = buildLangList()
	}
	return m
}

func buildLangList() list.Model {
 langs := service.MirrorLanguages()
	items := make([]list.Item, len(langs))
	for i, l := range langs {
		items[i] = mirrorLangItem{lang: l}
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = "选择语言/工具"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

func buildMirrorList(lang string) list.Model {
	reg := mirror.NewRegistry()
	mirrors := reg.ListMirrors(lang)
	items := make([]list.Item, 0, len(mirrors))
	for name, url := range mirrors {
		items = append(items, mirrorNameItem{name: name, url: url})
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = fmt.Sprintf("选择 %s 镜像源", lang)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

func (m *mirrorWizard) Init() tea.Cmd { return nil }

func (m *mirrorWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 8)
	case mirrorConfiguredMsg:
		m.phase = mirrorDone
		m.okMsg = successStyle.Render("镜像源配置成功！")
	case mirrorErrorMsg:
		m.phase = mirrorDone
		m.errMsg = errorStyle.Render(msg.err)
	case tea.KeyMsg:
		switch m.phase {
		case mirrorLang:
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(mirrorLangItem); ok {
					m.language = item.lang
					m.phase = mirrorPick
					m.list = buildMirrorList(item.lang)
				}
			case "esc", "q":
				m.done = true
			}
		case mirrorPick:
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(mirrorNameItem); ok {
					return m, m.configure(item.name)
				}
			case "esc":
				if m.language != "" && len(service.MirrorLanguages()) > 0 {
					m.phase = mirrorLang
					m.list = buildLangList()
				} else {
					m.done = true
				}
			case "q":
				m.done = true
			}
		case mirrorDone:
			switch msg.String() {
			case "enter", "esc", "q":
				m.done = true
			}
		}
	}
	var cmd tea.Cmd
	if m.phase != mirrorDone {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

type mirrorConfiguredMsg struct{}
type mirrorErrorMsg struct{ err string }

func (m *mirrorWizard) configure(name string) tea.Cmd {
	lang := m.language
	return func() tea.Msg {
		if err := service.ConfigureMirror(lang, name); err != nil {
			return mirrorErrorMsg{err: err.Error()}
		}
		return mirrorConfiguredMsg{}
	}
}

func (m *mirrorWizard) View() string {
	switch m.phase {
	case mirrorLang, mirrorPick:
		var b strings.Builder
		b.WriteString(renderTitle("镜像源配置", ""))
		b.WriteString("\n")
		b.WriteString(m.list.View())
		b.WriteString("\n")
		b.WriteString(backKeyHint())
		return b.String()
	case mirrorDone:
		msg := m.okMsg
		if m.errMsg != "" {
			msg = m.errMsg
		}
		return renderTitle("镜像源", "") + "\n" + msg + "\n" + renderHelp("Enter 返回")
	}
	return ""
}

func (m *mirrorWizard) Done() bool { return m.done }
