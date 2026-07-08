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

// GetTools 检测本机工具安装状态
func GetTools() []Tool {
	detected := detector.DetectTools()
	tools := []Tool{
		{Name: "git", DisplayName: "Git", Version: "latest", Installed: false},
		{Name: "docker", DisplayName: "Docker", Version: "latest", Installed: false},
		{Name: "code", DisplayName: "VS Code", Version: "latest", Installed: false},
		{Name: "vim", DisplayName: "Vim", Version: "latest", Installed: false},
		{Name: "curl", DisplayName: "cURL", Version: "latest", Installed: false},
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
