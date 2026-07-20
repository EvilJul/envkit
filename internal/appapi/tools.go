package appapi

import (
	"github.com/fusheng/envkit/internal/detector"
)

// Tool GUI 工具卡片数据
type Tool struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
}

// GetTools 检测本机工具安装状态。
// Name 与 installer 清单键一致，便于安装/卸载对齐（vscode 而非 code）。
func GetTools() []Tool {
	detected := detector.DetectTools()
	// 仅包含 GetToolInstaller 支持的工具，避免前端出现无法安装的假按钮
	tools := []Tool{
		{Name: "git", DisplayName: "Git", Version: "latest", Installed: false},
		{Name: "docker", DisplayName: "Docker", Version: "latest", Installed: false},
		{Name: "vscode", DisplayName: "VS Code", Version: "latest", Installed: false},
		{Name: "uv", DisplayName: "uv", Version: "latest", Installed: false},
		{Name: "miniconda", DisplayName: "Miniconda", Version: "latest", Installed: false},
		{Name: "kubectl", DisplayName: "Kubectl", Version: "latest", Installed: false},
		{Name: "minikube", DisplayName: "Minikube", Version: "latest", Installed: false},
		{Name: "espidf", DisplayName: "ESP-IDF", Version: "latest", Installed: false},
		{Name: "android", DisplayName: "Android SDK", Version: "latest", Installed: false},
	}

	for i := range tools {
		key := tools[i].Name
		// detector 对 VS Code 使用 "code" 作为键
		detectKey := key
		if key == "vscode" {
			detectKey = "code"
		}
		if key == "miniconda" {
			detectKey = "conda"
		}
		if tool := detected[detectKey]; tool != nil && tool.Installed {
			tools[i].Installed = true
			tools[i].Version = tool.Version
		}
	}

	android := GetAndroidInfo()
	for i := range tools {
		if tools[i].Name == "android" {
			tools[i].Installed = android.Installed
			if android.Version != "" {
				tools[i].Version = android.Version
			}
		}
	}

	return tools
}
