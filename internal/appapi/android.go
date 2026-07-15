package appapi

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
)

// AndroidInfo Android SDK 状态
type AndroidInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	SdkPath   string `json:"sdkPath"`
	ADBPath   string `json:"adbPath"`
	HasSDK    bool   `json:"hasSdk"`
}

// MirrorOption 镜像源选项（与前端 Tools/Settings 对齐）
type MirrorOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

var mirrorDisplayLabels = map[string]string{
	"aliyun":    "阿里云",
	"tencent":   "腾讯云",
	"huawei":    "华为云",
	"tsinghua":  "清华大学",
	"official":  "Official (官方源)",
	"npmmirror": "npmmirror (淘宝镜像)",
	"goproxy":   "goproxy.cn",
	"ustc":      "中科大",
}

func mirrorLabel(name string) string {
	if label, ok := mirrorDisplayLabels[name]; ok {
		return label
	}
	return name
}

func listMirrorOptions(mirrorType string) []MirrorOption {
	registry := mirror.NewRegistry()
	mirrors := registry.ListMirrors(mirrorType)
	result := make([]MirrorOption, 0, len(mirrors))
	for name, url := range mirrors {
		result = append(result, MirrorOption{
			Value: name,
			Label: mirrorLabel(name),
			URL:   url,
		})
	}
	return result
}

// InstallAndroid 安装 Android SDK
func InstallAndroid() error {
	toolInstaller := installer.GetToolInstaller("android")
	if toolInstaller == nil {
		return fmt.Errorf("不支持的工具: android")
	}
	return toolInstaller.Install()
}

// UninstallAndroid 卸载 Android SDK
func UninstallAndroid() error {
	return installer.UninstallComponent("android")
}

// GetAndroidInfo 获取 Android SDK 信息
func GetAndroidInfo() AndroidInfo {
	info := AndroidInfo{
		Installed: false,
		Version:   "",
		SdkPath:   "",
		ADBPath:   "",
		HasSDK:    false,
	}

	sdkRoot, err := resolveAndroidSdkRoot()
	if err != nil {
		return info
	}
	info.SdkPath = sdkRoot

	adbPath := filepath.Join(sdkRoot, "platform-tools", "adb")
	if runtime.GOOS == "windows" {
		adbPath += ".exe"
	}
	info.ADBPath = adbPath

	if _, err := os.Stat(sdkRoot); err == nil {
		info.HasSDK = true
	}

	if _, err := os.Stat(adbPath); err == nil {
		info.Installed = true
		detected := detector.DetectTools()
		if t, ok := detected["android"]; ok && t != nil && t.Installed && t.Version != "" {
			info.Version = t.Version
		}
	}

	return info
}

// ConfigureAndroidMirror 配置 Android SDK 镜像源
func ConfigureAndroidMirror(mirrorName string) error {
	registry := mirror.NewRegistry()
	return mirror.NewAndroidConfigurator(registry).Configure(mirrorName)
}

// ConfigureGradleMirror 配置 Gradle 镜像源
func ConfigureGradleMirror(mirrorName string) error {
	registry := mirror.NewRegistry()
	return mirror.NewGradleConfigurator(registry).Configure(mirrorName)
}

// GetAndroidMirrors 列出 Android 镜像源
func GetAndroidMirrors() []MirrorOption {
	return listMirrorOptions("android")
}

// GetGradleMirrors 列出 Gradle 镜像源
func GetGradleMirrors() []MirrorOption {
	return listMirrorOptions("gradle")
}

func resolveAndroidSdkRoot() (string, error) {
	if env := os.Getenv("ANDROID_HOME"); env != "" {
		return env, nil
	}
	if env := os.Getenv("ANDROID_SDK_ROOT"); env != "" {
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
