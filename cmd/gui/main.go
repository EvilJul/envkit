package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/fusheng/envkit/internal/appapi"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	detector.InitDetectionEnvironment()
}

// Language 语言环境信息
type Language = appapi.Language

// Tool 工具信息
type Tool struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
}

// Database 数据库容器信息
type Database struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"` // running, stopped
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Distribution string `json:"distribution"`
}

// GetSystemInfo 获取系统信息
func (a *App) GetSystemInfo() SystemInfo {
	sys := detector.DetectSystem()
	return SystemInfo{
		OS:           string(sys.OS),
		Architecture: string(sys.Architecture),
		Distribution: sys.Distribution,
	}
}

// GetLanguages 获取所有语言环境
func (a *App) GetLanguages() []Language {
	return appapi.GetLanguages()
}

func (a *App) SetLanguageMirror(language string, mirrorName string) error {
	return appapi.SetLanguageMirror(language, mirrorName)
}

// InstallLanguage 安装语言环境
func (a *App) InstallLanguage(name string, version string) error {
	langInstaller := installer.GetInstaller(name)
	if langInstaller == nil {
		return fmt.Errorf("不支持的语言: %s", name)
	}

	return langInstaller.Install(version)
}

// UninstallLanguage 卸载语言环境
func (a *App) UninstallLanguage(name string) error {
	return installer.UninstallComponent(name)
}

// ConfigureMirror 配置镜像源
func (a *App) ConfigureMirror(language string, mirrorName string) error {
	registry := mirror.NewRegistry()

	switch language {
	case "node", "nodejs":
		configurator := mirror.NewNPMConfigurator(registry)
		return configurator.Configure(mirrorName)
	case "python":
		configurator := mirror.NewPipConfigurator(registry)
		return configurator.Configure(mirrorName)
	case "go", "golang":
		configurator := mirror.NewGoConfigurator(registry)
		return configurator.Configure(mirrorName)
	case "rust":
		configurator := mirror.NewRustConfigurator(registry)
		return configurator.Configure(mirrorName)
	default:
		return fmt.Errorf("不支持的语言: %s", language)
	}
}

// GetTools 获取所有工具
func (a *App) GetTools() []Tool {
	detected := detector.DetectTools()

	tools := []Tool{
		{Name: "git", DisplayName: "Git", Version: "", Installed: false},
		{Name: "docker", DisplayName: "Docker", Version: "", Installed: false},
		{Name: "code", DisplayName: "VS Code", Version: "", Installed: false},
		{Name: "conda", DisplayName: "Miniconda", Version: "", Installed: false},
		{Name: "kubectl", DisplayName: "Kubectl", Version: "", Installed: false},
		{Name: "minikube", DisplayName: "Minikube", Version: "", Installed: false},
	}

	for i := range tools {
		if tool := detected[tools[i].Name]; tool != nil && tool.Installed {
			tools[i].Installed = true
			tools[i].Version = tool.Version
		}
	}

	return tools
}

// InstallTool 安装工具
func (a *App) InstallTool(name string) error {
	toolInstaller := installer.GetToolInstaller(name)
	if toolInstaller == nil {
		return fmt.Errorf("不支持的工具: %s", name)
	}

	return toolInstaller.Install()
}

// UninstallTool 卸载工具
func (a *App) UninstallTool(name string) error {
	return installer.UninstallComponent(name)
}

// GetDatabases 获取数据库容器列表
func (a *App) GetDatabases() []Database {
	// TODO: 实现 Docker 容器列表获取
	return []Database{}
}

// StartDatabase 启动数据库容器
func (a *App) StartDatabase(name string, version string) error {
	dockerMgr := docker.NewContainerManager()

	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	switch name {
	case "postgres", "postgresql":
		return dockerMgr.StartPostgreSQL(version, "postgres")
	case "redis":
		return dockerMgr.StartRedis(version)
	case "mysql":
		return dockerMgr.StartMySQL(version, "mysql")
	case "mongodb", "mongo":
		return dockerMgr.StartMongoDB(version)
	default:
		return fmt.Errorf("不支持的数据库: %s", name)
	}
}

// StopDatabase 停止数据库容器
func (a *App) StopDatabase(containerName string) error {
	dockerMgr := docker.NewContainerManager()
	return dockerMgr.StopContainer(containerName)
}

// RemoveDatabase 删除数据库容器
func (a *App) RemoveDatabase(containerName string, removeVolume bool) error {
	dockerMgr := docker.NewContainerManager()
	return dockerMgr.RemoveContainer(containerName, removeVolume)
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
				Message: "轻量级跨平台开发环境管理工具\nv0.2.0",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
