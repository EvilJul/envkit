package appapi

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings 应用设置（与前端 Settings.svelte 字段对齐）
type Settings struct {
	LaunchAtStartup  bool   `json:"launchAtStartup"`
	MinimizeToTray   bool   `json:"minimizeToTray"`
	DefaultNpmMirror string `json:"defaultNpmMirror"`
	DefaultPipMirror string `json:"defaultPipMirror"`
	DefaultGoMirror  string `json:"defaultGoMirror"`
	AndroidMirror    string `json:"androidMirror"`
	GradleMirror     string `json:"gradleMirror"`
	ShowExpertOptions bool  `json:"showExpertOptions"`
	InstallDir       string `json:"installDir"`
	LogLevel         string `json:"logLevel"`
}

func defaultSettings() Settings {
	return Settings{
		LaunchAtStartup:  false,
		MinimizeToTray:   true,
		DefaultNpmMirror: "npmmirror",
		DefaultPipMirror: "tsinghua",
		DefaultGoMirror:  "goproxy",
		AndroidMirror:    "aliyun",
		GradleMirror:     "aliyun",
		ShowExpertOptions: false,
		InstallDir:       "/usr/local",
		LogLevel:         "Info",
	}
}

func settingsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".envkit")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// GetSettings 读取应用设置
func GetSettings() Settings {
	settings := defaultSettings()
	path, err := settingsFilePath()
	if err != nil {
		return settings
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return settings
	}
	_ = json.Unmarshal(data, &settings)
	return settings
}

// SaveSettings 保存应用设置
func SaveSettings(settings Settings) (Settings, error) {
	path, err := settingsFilePath()
	if err != nil {
		return settings, err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return settings, err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return settings, err
	}
	return settings, nil
}

// ResetSettings 重置为默认设置并持久化
func ResetSettings() Settings {
	settings := defaultSettings()
	_, _ = SaveSettings(settings)
	return settings
}
