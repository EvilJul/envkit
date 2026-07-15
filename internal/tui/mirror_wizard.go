package tui

import (
	"fmt"

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
	delegate := configureWizardListDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return l
}

func buildMirrorList(lang string) list.Model {
	reg := mirror.NewRegistry()
	mirrors := reg.ListMirrors(lang)
	items := make([]list.Item, 0, len(mirrors))
	for name, url := range mirrors {
		items = append(items, mirrorNameItem{name: name, url: url})
	}
	delegate := configureWizardListDelegate()
	l := list.New(items, delegate, 40, 12)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
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
		m.okMsg = "镜像源已写入本地配置。"
	case mirrorErrorMsg:
		m.phase = mirrorDone
		m.errMsg = msg.err
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
	steps := []string{"语言", "镜像", "完成"}

	switch m.phase {
	case mirrorLang:
		stepBar := RenderStepIndicator(steps, 0)
		body := stepBar + "\n\n" + m.list.View()
		return RenderChrome("镜像源配置", "选择语言/工具", body, hintsNavSelect)
	case mirrorPick:
		stepBar := RenderStepIndicator(steps, 1)
		sub := "选择镜像源"
		if m.language != "" {
			sub = fmt.Sprintf("为 %s 选择镜像源", m.language)
		}
		body := stepBar + "\n\n" + m.list.View()
		return RenderChrome("镜像源配置", sub, body, hintsNavSelect)
	case mirrorDone:
		stepBar := RenderStepIndicator(steps, 2)
		var banner string
		if m.errMsg != "" {
			banner = RenderBanner("error", "配置失败", m.errMsg)
		} else {
			detail := m.okMsg
			if m.language != "" {
				detail = mutedStyle.Render("语言/工具: "+m.language) + "\n" + detail
			}
			banner = RenderBanner("success", "镜像源配置成功", detail)
		}
		return RenderChrome("镜像源", "", stepBar+"\n\n"+banner, hintsEnterBack)
	}
	return ""
}

func (m *mirrorWizard) Done() bool { return m.done }
