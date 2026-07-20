//go:build ignore

package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
)

//go:embed all:../../frontend/dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type Language struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
	Mirror      string `json:"mirror"`
}

type Tool struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
}

type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Distribution string `json:"distribution"`
}

func (a *App) GetSystemInfo() SystemInfo {
	sys := detector.DetectSystem()
	return SystemInfo{
		OS:           string(sys.OS),
		Architecture: string(sys.Architecture),
		Distribution: sys.Distribution,
	}
}

func (a *App) GetLanguages() []Language {
	detected := detector.DetectLanguages()

	languages := []Language{
		{Name: "node", DisplayName: "Node.js", Version: "20.11.1", Installed: false},
		{Name: "python", DisplayName: "Python", Version: "3.10.11", Installed: false},
		{Name: "go", DisplayName: "Go", Version: "1.22.0", Installed: false},
		{Name: "rust", DisplayName: "Rust", Version: "stable", Installed: false},
		{Name: "java", DisplayName: "Java", Version: "21", Installed: false},
		{Name: "bun", DisplayName: "Bun", Version: "latest", Installed: false},
	}

	for i := range languages {
		key := languages[i].Name
		if key == "rust" {
			key = "rustc"
		}
		if tool := detected[key]; tool != nil && tool.Installed {
			languages[i].Installed = true
			languages[i].Version = tool.Version
		}
	}
	return languages
}

func (a *App) InstallLanguage(name string, version string) error {
	langInstaller := installer.GetInstaller(name)
	if langInstaller == nil {
		return fmt.Errorf("不支持的语言: %s", name)
	}
	return langInstaller.Install(version)
}

func (a *App) UninstallLanguage(name string) error {
	return installer.UninstallComponent(name)
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "EnvKit",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   "EnvKit",
				Message: "轻量级跨平台开发环境管理工具",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
