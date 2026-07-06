package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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

type AndroidInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	SdkPath   string `json:"sdkPath"`
	ADBPath   string `json:"adbPath"`
	HasSDK    bool   `json:"hasSdk"`
}

type MirrorInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
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

func (a *App) GetTools() []Tool {
	detected := detector.DetectTools()

	tools := []Tool{
		{Name: "git", DisplayName: "Git", Version: "latest", Installed: false},
		{Name: "docker", DisplayName: "Docker", Version: "latest", Installed: false},
		{Name: "code", DisplayName: "VSCode", Version: "latest", Installed: false},
		{Name: "vim", DisplayName: "Vim", Version: "latest", Installed: false},
		{Name: "curl", DisplayName: "Curl", Version: "latest", Installed: false},
		{Name: "wget", DisplayName: "Wget", Version: "latest", Installed: false},
		{Name: "conda", DisplayName: "Miniconda", Version: "latest", Installed: false},
		{Name: "kubectl", DisplayName: "Kubectl", Version: "latest", Installed: false},
		{Name: "minikube", DisplayName: "Minikube", Version: "latest", Installed: false},
		{Name: "espidf", DisplayName: "ESP-IDF", Version: "latest", Installed: false},
		{Name: "android", DisplayName: "Android SDK", Version: "latest", Installed: false},
	}

	for i := range tools {
		key := tools[i].Name
		if key == "code" {
			key = "vscode"
		}
		if tool := detected[key]; tool != nil && tool.Installed {
			tools[i].Installed = true
			tools[i].Version = tool.Version
		}
	}
	return tools
}

func (a *App) InstallTool(name string) error {
	toolInstaller := installer.GetToolInstaller(name)
	if toolInstaller == nil {
		return fmt.Errorf("不支持的工具: %s", name)
	}
	return toolInstaller.Install()
}

func (a *App) UninstallTool(name string) error {
	return installer.UninstallComponent(name)
}

func (a *App) InstallAndroid() error {
	toolInstaller := installer.GetToolInstaller("android")
	if toolInstaller == nil {
		return fmt.Errorf("不支持的工具: android")
	}
	return toolInstaller.Install()
}

func (a *App) UninstallAndroid() error {
	return installer.UninstallComponent("android")
}

func (a *App) GetAndroidInfo() AndroidInfo {
	info := AndroidInfo{
		Installed: false,
		Version:   "",
		SdkPath:   "",
		ADBPath:   "",
		HasSDK:    false,
	}

	// 计算 SDK 根目录（与 internal/installer/android.go 中逻辑保持一致）
	sdkRoot, err := resolveAndroidSdkRoot()
	if err != nil {
		return info
	}
	info.SdkPath = sdkRoot

	// adb 路径
	adbPath := filepath.Join(sdkRoot, "platform-tools", "adb")
	if runtime.GOOS == "windows" {
		adbPath += ".exe"
	}
	info.ADBPath = adbPath

	// 检查 SDK 目录是否存在
	if _, err := os.Stat(sdkRoot); err == nil {
		info.HasSDK = true
	}

	// 检查 adb 是否存在
	if _, err := os.Stat(adbPath); err == nil {
		info.Installed = true
		// 读取 adb 版本
		detected := detector.DetectTools()
		if t, ok := detected["android"]; ok && t != nil && t.Installed && t.Version != "" {
			info.Version = t.Version
		}
	}

	return info
}

func (a *App) ConfigureAndroidMirror(mirrorName string) error {
	registry := mirror.NewRegistry()
	configurator := mirror.NewAndroidConfigurator(registry)
	return configurator.Configure(mirrorName)
}

func (a *App) ConfigureGradleMirror(mirrorName string) error {
	registry := mirror.NewRegistry()
	configurator := mirror.NewGradleConfigurator(registry)
	return configurator.Configure(mirrorName)
}

func (a *App) GetAndroidMirrors() []MirrorInfo {
	registry := mirror.NewRegistry()
	mirrors := registry.ListMirrors("android")
	result := make([]MirrorInfo, 0, len(mirrors))
	for name, url := range mirrors {
		result = append(result, MirrorInfo{Name: name, URL: url})
	}
	return result
}

func (a *App) GetGradleMirrors() []MirrorInfo {
	registry := mirror.NewRegistry()
	mirrors := registry.ListMirrors("gradle")
	result := make([]MirrorInfo, 0, len(mirrors))
	for name, url := range mirrors {
		result = append(result, MirrorInfo{Name: name, URL: url})
	}
	return result
}

// resolveAndroidSdkRoot 计算 Android SDK 根目录（与 internal/installer/android.go.getSdkRoot 保持一致）
func resolveAndroidSdkRoot() (string, error) {
	if env := os.Getenv("ANDROID_HOME"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Android", "sdk"), nil
	case "linux":
		return filepath.Join(home, "Android", "Sdk"), nil
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Android", "Sdk"), nil
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
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
