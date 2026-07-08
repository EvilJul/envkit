package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/fusheng/envkit/internal/appapi"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/progress"
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

type Language = appapi.Language

type Tool = appapi.Tool

type Database struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type Stack struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Services    []string `json:"services"`
	Status      string   `json:"status"`
}

type ProjectTemplate struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
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
	return appapi.GetLanguages()
}

func (a *App) InstallLanguage(name string, version string) error {
	return appapi.InstallLanguageWithProgress(name, version, progress.NewWailsReporter(a.ctx, "task-progress", name))
}

func (a *App) UninstallLanguage(name string) error {
	return appapi.UninstallLanguageWithProgress(name, progress.NewWailsReporter(a.ctx, "task-progress", name))
}

func (a *App) SetLanguageMirror(language string, mirrorName string) error {
	return appapi.SetLanguageMirror(language, mirrorName)
}

func (a *App) GetTools() []Tool {
	return appapi.GetTools()
}

func (a *App) InstallTool(name string) error {
	return appapi.InstallToolWithProgress(name, progress.NewWailsReporter(a.ctx, "task-progress", name))
}

func (a *App) UninstallTool(name string) error {
	return appapi.UninstallToolWithProgress(name, progress.NewWailsReporter(a.ctx, "task-progress", name))
}

type AndroidInfo = appapi.AndroidInfo

func (a *App) InstallAndroid() error {
	return appapi.InstallAndroidWithProgress(progress.NewWailsReporter(a.ctx, "task-progress", "android"))
}

func (a *App) UninstallAndroid() error {
	return appapi.UninstallAndroidWithProgress(progress.NewWailsReporter(a.ctx, "task-progress", "android"))
}

func (a *App) GetAndroidInfo() AndroidInfo {
	return appapi.GetAndroidInfo()
}

func (a *App) ConfigureAndroidMirror(mirrorName string) error {
	return appapi.ConfigureAndroidMirrorWithProgress(mirrorName, progress.NewWailsReporter(a.ctx, "task-progress", "android-mirror"))
}

func (a *App) ConfigureGradleMirror(mirrorName string) error {
	return appapi.ConfigureGradleMirrorWithProgress(mirrorName, progress.NewWailsReporter(a.ctx, "task-progress", "gradle-mirror"))
}

func (a *App) GetAndroidMirrors() []appapi.MirrorOption {
	return appapi.GetAndroidMirrors()
}

func (a *App) GetGradleMirrors() []appapi.MirrorOption {
	return appapi.GetGradleMirrors()
}

type Settings = appapi.Settings

func (a *App) GetSettings() Settings {
	return appapi.GetSettings()
}

func (a *App) SaveSettings(settings Settings) (Settings, error) {
	return appapi.SaveSettings(settings)
}

func (a *App) ResetSettings() Settings {
	return appapi.ResetSettings()
}

func (a *App) GetDatabases() []Database {
	// TODO: 实现 Docker 容器列表获取
	return []Database{}
}

func (a *App) StartDatabase(name string, version string) error {
	return appapi.StartDatabaseWithProgress(name, version, progress.NewWailsReporter(a.ctx, "task-progress", name))
}

func (a *App) StopDatabase(containerName string) error {
	return appapi.StopDatabaseWithProgress(containerName, progress.NewWailsReporter(a.ctx, "task-progress", containerName))
}

func (a *App) RemoveDatabase(containerName string, removeVolume bool) error {
	return appapi.RemoveDatabaseWithProgress(containerName, removeVolume, progress.NewWailsReporter(a.ctx, "task-progress", containerName))
}

func (a *App) GetStacks() []Stack {
	// TODO: 实现运行中的技术栈列表
	return []Stack{}
}

func (a *App) StartStack(name string) error {
	dockerMgr := docker.NewContainerManager()
	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	switch name {
	case "lamp":
		// 启动 Apache + MySQL + PHP
		if err := dockerMgr.StartMySQL("8.0", "mysql"); err != nil {
			return err
		}
		// TODO: 启动 Apache + PHP 容器
		return nil
	case "lemp":
		// 启动 Nginx + MySQL + PHP
		if err := dockerMgr.StartMySQL("8.0", "mysql"); err != nil {
			return err
		}
		// TODO: 启动 Nginx + PHP 容器
		return nil
	case "mean":
		// 启动 MongoDB + Express + Node.js
		if err := dockerMgr.StartMongoDB("6.0"); err != nil {
			return err
		}
		// TODO: 启动 Express + Node.js 容器
		return nil
	case "mern":
		// 启动 MongoDB + Express + React + Node.js
		if err := dockerMgr.StartMongoDB("6.0"); err != nil {
			return err
		}
		// TODO: 启动 MERN 应用容器
		return nil
	case "django":
		// 启动 PostgreSQL + Redis + Python
		if err := dockerMgr.StartPostgreSQL("16", "postgres"); err != nil {
			return err
		}
		if err := dockerMgr.StartRedis("7"); err != nil {
			return err
		}
		return nil
	case "laravel":
		// 启动 Nginx + MySQL + Redis + PHP
		if err := dockerMgr.StartMySQL("8.0", "mysql"); err != nil {
			return err
		}
		if err := dockerMgr.StartRedis("7"); err != nil {
			return err
		}
		// TODO: 启动 Nginx + PHP 容器
		return nil
	default:
		return fmt.Errorf("不支持的技术栈: %s", name)
	}
}

func (a *App) StopStack(name string) error {
	// TODO: 实现停止技术栈
	return fmt.Errorf("功能开发中")
}

func (a *App) GetProjectTemplates() []ProjectTemplate {
	return []ProjectTemplate{
		{
			Name:        "react-ts",
			DisplayName: "React + TypeScript",
			Description: "使用 Vite 构建的现代 React 应用",
			Tech:        []string{"React", "TypeScript", "Vite"},
		},
		{
			Name:        "vue3-ts",
			DisplayName: "Vue 3 + TypeScript",
			Description: "Vue 3 Composition API + TypeScript",
			Tech:        []string{"Vue 3", "TypeScript", "Vite"},
		},
		{
			Name:        "nextjs",
			DisplayName: "Next.js",
			Description: "全栈 React 框架",
			Tech:        []string{"Next.js", "React", "TypeScript"},
		},
		{
			Name:        "express-api",
			DisplayName: "Express API",
			Description: "RESTful API 模板",
			Tech:        []string{"Express", "Node.js", "TypeScript"},
		},
		{
			Name:        "fastapi",
			DisplayName: "FastAPI",
			Description: "现代 Python API 框架",
			Tech:        []string{"FastAPI", "Python", "Pydantic"},
		},
		{
			Name:        "go-web",
			DisplayName: "Go Web Service",
			Description: "使用 Gin 框架的 Go 服务",
			Tech:        []string{"Go", "Gin", "GORM"},
		},
	}
}

func (a *App) CreateProject(templateName string, projectPath string) error {
	// TODO: 实现项目模板创建
	return fmt.Errorf("功能开发中")
}

func (a *App) ScanEnvFiles(projectPath string) []map[string]interface{} {
	// TODO: 实现扫描 .env 文件
	return []map[string]interface{}{}
}

func (a *App) LoadEnvFile(filePath string) (string, error) {
	// TODO: 实现加载 .env 文件
	return "", fmt.Errorf("功能开发中")
}

func (a *App) SaveEnvFile(filePath string, content string) error {
	// TODO: 实现保存 .env 文件
	return fmt.Errorf("功能开发中")
}

func (a *App) CreateEnvFromTemplate(projectPath string, fileName string, templateName string) error {
	// TODO: 实现从模板创建 .env 文件
	return fmt.Errorf("功能开发中")
}

func (a *App) GetEnvVariables() map[string]interface{} {
	homeDir, _ := os.UserHomeDir()

	// 检测当前 shell
	shell := detectShell()

	// 获取配置文件路径
	var configFile string
	if shell == "zsh" {
		configFile = filepath.Join(homeDir, ".zshrc")
	} else {
		configFile = filepath.Join(homeDir, ".bash_profile")
	}

	// 解析用户配置文件
	userVars := parseEnvFile(configFile)

	// 解析系统配置文件
	systemVars := parseEnvFile("/etc/profile")

	// 获取 PATH 变量
	pathValue := os.Getenv("PATH")
	pathEntries := strings.Split(pathValue, ":")

	return map[string]interface{}{
		"user":   userVars,
		"system": systemVars,
		"path":   pathEntries,
		"shell":  shell,
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	} else if strings.Contains(shell, "bash") {
		return "bash"
	}
	return "zsh"
}

func parseEnvFile(filePath string) []map[string]string {
	vars := []map[string]string{}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return vars
	}

	lines := strings.Split(string(content), "\n")
	// 匹配 export VAR=value 格式（支持大小写、下划线、变量引用、引号）
	exportRegex := regexp.MustCompile(`^\s*export\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)

	for _, line := range lines {
		// 不要 trim，保留原始缩进信息
		trimmed := strings.TrimSpace(line)

		// 跳过注释和空行
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}

		// 匹配 export VAR=value 格式
		matches := exportRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			value := matches[2]
			// 移除首尾的引号（如果有）
			value = strings.Trim(value, `"'`)

			vars = append(vars, map[string]string{
				"key":    matches[1],
				"value":  value,
				"source": filePath,
				"scope":  getScopeFromPath(filePath),
			})
		}
	}

	return vars
}

func getScopeFromPath(filePath string) string {
	if strings.HasPrefix(filePath, "/etc") {
		return "system"
	}
	return "user"
}

func (a *App) SetEnvVariable(key string, value string, scope string) error {
	homeDir, _ := os.UserHomeDir()
	shell := detectShell()

	var configFile string
	if scope == "user" {
		if shell == "zsh" {
			configFile = filepath.Join(homeDir, ".zshrc")
		} else {
			configFile = filepath.Join(homeDir, ".bash_profile")
		}
	} else {
		configFile = "/etc/profile"
	}

	// 读取现有内容
	content, err := os.ReadFile(configFile)
	if err != nil {
		// 文件不存在，创建新文件
		content = []byte("")
	}

	lines := strings.Split(string(content), "\n")
	exportLine := fmt.Sprintf("export %s=\"%s\"", key, value)
	varRegex := regexp.MustCompile(fmt.Sprintf(`^\s*export\s+%s\s*=`, key))

	found := false
	for i, line := range lines {
		if varRegex.MatchString(line) {
			lines[i] = exportLine
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, exportLine)
	}

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(configFile, []byte(newContent), 0644)
}

func (a *App) DeleteEnvVariable(key string, scope string) error {
	homeDir, _ := os.UserHomeDir()
	shell := detectShell()

	var configFile string
	if scope == "user" {
		if shell == "zsh" {
			configFile = filepath.Join(homeDir, ".zshrc")
		} else {
			configFile = filepath.Join(homeDir, ".bash_profile")
		}
	} else {
		configFile = "/etc/profile"
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	varRegex := regexp.MustCompile(fmt.Sprintf(`^\s*export\s+%s\s*=`, key))

	newLines := []string{}
	for _, line := range lines {
		if !varRegex.MatchString(line) {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(configFile, []byte(newContent), 0644)
}

func (a *App) GetShellConfig() string {
	return detectShell()
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
				Message: "轻量级跨平台开发环境管理工具 v0.2.0",
			},
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
