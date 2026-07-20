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

type Database = appapi.Database

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
	return appapi.GetDatabases()
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

// stackContainers 技术栈对应的 envkit 容器名（完整编排未实现的栈不在此表）
var stackContainers = map[string][]string{
	"django": {"envkit-postgres", "envkit-redis"},
}

func (a *App) StartStack(name string) error {
	dockerMgr := docker.NewContainerManager()
	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	switch name {
	case "django":
		// 仅启动 Django 常用数据库依赖（PostgreSQL + Redis）
		if err := dockerMgr.StartPostgreSQL("16", "postgres"); err != nil {
			return err
		}
		if err := dockerMgr.StartRedis("7"); err != nil {
			return err
		}
		return nil
	case "lamp", "lemp", "mean", "mern", "laravel":
		// 避免假成功：完整应用容器编排尚未实现
		return fmt.Errorf("技术栈 '%s' 的完整编排尚未实现，请在 Database 页单独启动所需数据库", name)
	default:
		return fmt.Errorf("不支持的技术栈: %s", name)
	}
}

func (a *App) StopStack(name string) error {
	containers, ok := stackContainers[name]
	if !ok {
		return fmt.Errorf("技术栈 '%s' 不支持一键停止（或尚未实现），请在 Database 页单独操作容器", name)
	}
	dockerMgr := docker.NewContainerManager()
	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}
	var errs []string
	for _, c := range containers {
		if err := dockerMgr.StopContainer(c); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("部分容器停止失败: %s", strings.Join(errs, "; "))
	}
	return nil
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
	shell := detectShell()
	configFile := userShellConfigFile(shell)

	userVars := parseEnvFile(configFile)
	systemVars := parseEnvFile("/etc/profile")

	pathValue := os.Getenv("PATH")
	sep := pathListSeparator()
	var pathEntries []string
	if pathValue == "" {
		pathEntries = []string{}
	} else {
		pathEntries = strings.Split(pathValue, sep)
	}

	return map[string]interface{}{
		"user":   userVars,
		"system": systemVars,
		"path":   pathEntries,
		"shell":  shell,
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	base := filepath.Base(shell)
	switch base {
	case "zsh":
		return "zsh"
	case "bash":
		return "bash"
	case "fish":
		return "fish"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	// Linux 默认多为 bash；macOS 多为 zsh
	if _, err := os.Stat("/bin/zsh"); err == nil && filepath.Base(os.Getenv("SHELL")) == "zsh" {
		return "zsh"
	}
	return "bash"
}

// userShellConfigFile 返回用户级 shell 配置文件路径（Linux bash 优先 .bashrc）
func userShellConfigFile(shell string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch shell {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc")
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "config.fish")
	default:
		// bash：Linux 交互式 shell 读 .bashrc；.bash_profile 仅登录 shell
		bashrc := filepath.Join(homeDir, ".bashrc")
		bashProfile := filepath.Join(homeDir, ".bash_profile")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile
		}
		return bashrc
	}
}

func pathListSeparator() string {
	// 与 os.PathListSeparator 一致，避免硬编码 ':' 导致 Windows 错误
	return string(os.PathListSeparator)
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
	if key == "" {
		return fmt.Errorf("环境变量名不能为空")
	}
	// 防止通过变量名注入破坏配置文件正则/内容
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(key) {
		return fmt.Errorf("非法的环境变量名: %s", key)
	}

	var configFile string
	if scope == "user" {
		configFile = userShellConfigFile(detectShell())
	} else {
		configFile = "/etc/profile"
	}
	if configFile == "" {
		return fmt.Errorf("无法确定 shell 配置文件路径")
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		content = []byte("")
	}

	// 转义写入值中的双引号与反斜杠
	safeValue := strings.ReplaceAll(value, `\`, `\\`)
	safeValue = strings.ReplaceAll(safeValue, `"`, `\"`)
	exportLine := fmt.Sprintf("export %s=\"%s\"", key, safeValue)
	varRegex := regexp.MustCompile(fmt.Sprintf(`^\s*export\s+%s\s*=`, regexp.QuoteMeta(key)))

	lines := strings.Split(string(content), "\n")
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

	// 确保目录存在（如 fish config）
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0644)
}

func (a *App) DeleteEnvVariable(key string, scope string) error {
	if key == "" {
		return fmt.Errorf("环境变量名不能为空")
	}
	if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(key) {
		return fmt.Errorf("非法的环境变量名: %s", key)
	}

	var configFile string
	if scope == "user" {
		configFile = userShellConfigFile(detectShell())
	} else {
		configFile = "/etc/profile"
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	varRegex := regexp.MustCompile(fmt.Sprintf(`^\s*export\s+%s\s*=`, regexp.QuoteMeta(key)))
	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if !varRegex.MatchString(line) {
			newLines = append(newLines, line)
		}
	}
	return os.WriteFile(configFile, []byte(strings.Join(newLines, "\n")), 0644)
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
