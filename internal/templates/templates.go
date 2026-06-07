package templates

import (
	_ "embed"

	"github.com/fusheng/envkit/internal/config"
	"gopkg.in/yaml.v3"
)

//go:embed frontend.yaml
var frontendTemplate []byte

//go:embed backend.yaml
var backendTemplate []byte

//go:embed fullstack.yaml
var fullstackTemplate []byte

// TemplateType 模板类型
type TemplateType string

const (
	Frontend  TemplateType = "frontend"
	Backend   TemplateType = "backend"
	Fullstack TemplateType = "fullstack"
)

// Templates 模板管理器
type Templates struct {
	templates map[TemplateType][]byte
}

// New 创建模板管理器
func New() *Templates {
	return &Templates{
		templates: map[TemplateType][]byte{
			Frontend:  frontendTemplate,
			Backend:   backendTemplate,
			Fullstack: fullstackTemplate,
		},
	}
}

// Get 获取指定类型的模板配置
func (t *Templates) Get(templateType TemplateType) (*config.Config, error) {
	data, ok := t.templates[templateType]
	if !ok {
		return nil, ErrTemplateNotFound
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// List 列出所有可用模板
func (t *Templates) List() []TemplateInfo {
	return []TemplateInfo{
		{
			Type:        Frontend,
			Name:        "前端开发环境",
			Description: "Node.js + pnpm，适合前端开发",
			Languages:   []string{"Node.js"},
		},
		{
			Type:        Backend,
			Name:        "后端开发环境",
			Description: "Go + Python + Docker，适合后端开发",
			Languages:   []string{"Go", "Python"},
		},
		{
			Type:        Fullstack,
			Name:        "全栈开发环境",
			Description: "Node.js + Go + Python + Docker，适合全栈开发",
			Languages:   []string{"Node.js", "Go", "Python"},
		},
	}
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	Type        TemplateType
	Name        string
	Description string
	Languages   []string
}

var (
	ErrTemplateNotFound = &TemplateError{"template not found"}
)

// TemplateError 模板错误
type TemplateError struct {
	msg string
}

func (e *TemplateError) Error() string {
	return e.msg
}
