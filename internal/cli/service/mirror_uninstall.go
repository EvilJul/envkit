package service

import (
	"fmt"

	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
)

// ConfigureMirror 配置镜像源
func ConfigureMirror(language, mirrorName string) error {
	registry := mirror.NewRegistry()
	switch language {
	case "npm":
		return mirror.NewNPMConfigurator(registry).Configure(mirrorName)
	case "pip":
		return mirror.NewPipConfigurator(registry).Configure(mirrorName)
	case "go":
		return mirror.NewGoConfigurator(registry).Configure(mirrorName)
	case "rust":
		return mirror.NewRustConfigurator(registry).Configure(mirrorName)
	case "android":
		return mirror.NewAndroidConfigurator(registry).Configure(mirrorName)
	case "gradle":
		return mirror.NewGradleConfigurator(registry).Configure(mirrorName)
	default:
		return fmt.Errorf("不支持的语言: %s", language)
	}
}

// MirrorLanguages 支持的 mirror 子命令语言
func MirrorLanguages() []string {
	return []string{"npm", "pip", "go", "rust", "gradle", "android"}
}

// ManifestItem 清单项摘要
type ManifestItem struct {
	Name        string
	Type        string
	Version     string
	InstalledAt string
}

// LoadManifestItems 加载已安装组件清单
func LoadManifestItems() ([]ManifestItem, error) {
	m, err := installer.LoadManifest()
	if err != nil {
		return nil, err
	}
	items := make([]ManifestItem, 0, len(m.Items))
	for name, item := range m.Items {
		typeStr := "tool"
		if item.Type == "language" {
			typeStr = "language"
		}
		items = append(items, ManifestItem{
			Name:        name,
			Type:        typeStr,
			Version:     item.Version,
			InstalledAt: item.InstalledAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, nil
}

// UninstallComponent 卸载单个组件
func UninstallComponent(name string) error {
	return installer.UninstallComponent(name)
}

// UninstallAll 卸载全部清单组件
func UninstallAll() []error {
	m, err := installer.LoadManifest()
	if err != nil {
		return []error{err}
	}
	var errs []error
	for name := range m.Items {
		if err := installer.UninstallComponent(name); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errs
}
