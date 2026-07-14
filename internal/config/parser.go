package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parser 配置文件解析器
type Parser struct {
	filePath string
}

// NewParser 创建配置解析器
func NewParser(filePath string) *Parser {
	return &Parser{
		filePath: filePath,
	}
}

// Parse 解析配置文件
func (p *Parser) Parse() (*Config, error) {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := p.Validate(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Validate 验证配置
func (p *Parser) Validate(config *Config) error {
	if config.Version == "" {
		return fmt.Errorf("配置版本不能为空")
	}

	if config.Name == "" {
		return fmt.Errorf("配置名称不能为空")
	}

	// 验证语言配置
	for _, lang := range config.Languages {
		if lang.Name == "" {
			return fmt.Errorf("语言名称不能为空")
		}
		if !isValidLanguage(lang.Name) {
			return fmt.Errorf("不支持的语言: %s", lang.Name)
		}
	}

	return nil
}

// isValidLanguage 检查是否为支持的语言
func isValidLanguage(name string) bool {
	validLanguages := []string{
		"node", "nodejs",
		"python", "python3",
		"go", "golang",
		"rust",
		"java", "jdk",
		"bun",
	}

	for _, valid := range validLanguages {
		if name == valid {
			return true
		}
	}
	return false
}

// Export 导出配置到文件
func Export(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}
