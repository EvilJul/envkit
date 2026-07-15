package appapi

import (
	"fmt"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/mirror"
)

// Language GUI 语言卡片数据（与前端 Languages.svelte 对齐）
type Language struct {
	Name              string                         `json:"name"`
	DisplayName       string                         `json:"displayName"`
	Description       string                         `json:"description"`
	Icon              string                         `json:"icon"`
	Version           string                         `json:"version"`
	Installed         bool                           `json:"installed"`
	Mirror            string                         `json:"mirror"`
	AvailableVersions []config.LanguageVersionOption `json:"availableVersions"`
	MirrorOptions     []string                       `json:"mirrorOptions"`
}

// GetLanguages 检测本机语言环境并合并目录元数据
func GetLanguages() []Language {
	detected := detector.DetectLanguages()
	catalog := config.LanguageCatalog()
	languages := make([]Language, 0, len(catalog))

	for _, entry := range catalog {
		lang := Language{
			Name:              entry.Name,
			DisplayName:       entry.DisplayName,
			Description:       entry.Description,
			Icon:              entry.Icon,
			Version:           entry.DefaultVersion,
			Installed:         false,
			Mirror:            entry.DefaultMirror,
			AvailableVersions: entry.AvailableVersions,
			MirrorOptions:     entry.MirrorOptions,
		}

		key := entry.Name
		if key == "rust" {
			key = "rustc"
		}
		if tool := detected[key]; tool != nil && tool.Installed {
			lang.Installed = true
			lang.Version = tool.Version
		}
		languages = append(languages, lang)
	}
	return languages
}

// SetLanguageMirror 配置语言包管理器镜像源
func SetLanguageMirror(language string, mirrorName string) error {
	registry := mirror.NewRegistry()
	switch language {
	case "node", "nodejs", "bun":
		return mirror.NewNPMConfigurator(registry).Configure(mirrorName)
	case "python":
		return mirror.NewPipConfigurator(registry).Configure(mirrorName)
	case "go", "golang":
		return mirror.NewGoConfigurator(registry).Configure(mirrorName)
	case "rust":
		return mirror.NewRustConfigurator(registry).Configure(mirrorName)
	default:
		return fmt.Errorf("不支持的语言镜像配置: %s", language)
	}
}
