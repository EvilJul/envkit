package service

import (
	"fmt"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/templates"
)

// InitFromTemplate 从模板生成配置文件
func InitFromTemplate(templateIndex int, outputFile string) error {
	tmpl := templates.New()
	list := tmpl.List()
	if templateIndex < 1 || templateIndex > len(list) {
		return fmt.Errorf("无效的模板选项: %d", templateIndex)
	}
	selected := list[templateIndex-1]
	cfg, err := tmpl.Get(selected.Type)
	if err != nil {
		return fmt.Errorf("加载模板失败: %w", err)
	}
	if outputFile == "" {
		outputFile = "dev-env.yaml"
	}
	if err := config.Export(cfg, outputFile); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}
	return nil
}

// TemplateList 返回可用模板
func TemplateList() []templates.TemplateInfo {
	return templates.New().List()
}
