package service

import (
	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
)

// CatalogRow 组件目录行
type CatalogRow struct {
	Name        string
	Version     string
	Status      string
	Description string
	Category    string // language, tool, database
}

// ListCatalog 返回支持安装的组件及当前状态
func ListCatalog() []CatalogRow {
	languages := detector.DetectLanguages()
	tools := detector.DetectTools()

	rows := []CatalogRow{
		{"node", "20.11.1", StatusLabel(languages["node"]), "Fast Node Manager (fnm)", "language"},
		{"python", "3.10.11", StatusLabel(languages["python"]), "Astral uv", "language"},
		{"go", "1.22.0", StatusLabel(languages["go"]), "官方包或 Homebrew", "language"},
		{"rust", "stable", StatusLabel(languages["rustc"]), "rustup 官方安装器", "language"},
		{"java", "21", StatusLabel(languages["java"]), "SDKMAN! 或 Microsoft OpenJDK", "language"},
		{"bun", "latest", StatusLabel(languages["bun"]), "Bun 官方安装器", "language"},
		{"git", "", StatusLabel(tools["git"]), "版本控制系统", "tool"},
		{"docker", "", StatusLabel(tools["docker"]), "应用容器引擎", "tool"},
		{"vscode", "", StatusLabel(tools["code"]), "Visual Studio Code 编辑器", "tool"},
		{"miniconda", "", StatusLabel(tools["conda"]), "Conda 虚拟环境与包管理器", "tool"},
		{"kubectl", "", StatusLabel(tools["kubectl"]), "Kubernetes 命令行控制工具", "tool"},
		{"minikube", "", StatusLabel(tools["minikube"]), "本地单节点 Kubernetes 集群", "tool"},
		{"espidf", "", StatusLabel(tools["espidf"]), "ESP-IDF (EIM) 乐鑫物联网开发框架", "tool"},
		{"android", "", StatusLabel(tools["android"]), "Android SDK", "tool"},
		{"postgresql", "16", "", "PostgreSQL 关系型数据库", "database"},
		{"redis", "7", "", "Redis 键值内存数据库", "database"},
		{"mysql", "8.0", "", "MySQL 关系型数据库", "database"},
		{"mongodb", "6.0", "", "MongoDB 文档型数据库", "database"},
	}
	return rows
}

// InstallOption 交互式安装选项
type InstallOption struct {
	Name     string
	Key      string
	Category string // language, tool, database
	Version  string
}

// InstallOptions 可交互选择的安装项
func InstallOptions() []InstallOption {
	return []InstallOption{
		{"Node.js", "node", "language", "20.11.1"},
		{"Python", "python", "language", "3.10.11"},
		{"Go", "go", "language", "1.22.0"},
		{"Rust", "rust", "language", "stable"},
		{"Java (JDK)", "java", "language", "21"},
		{"Bun", "bun", "language", "latest"},
		{"Git", "git", "tool", ""},
		{"Docker", "docker", "tool", ""},
		{"VSCode", "vscode", "tool", ""},
		{"Miniconda", "miniconda", "tool", ""},
		{"Kubectl", "kubectl", "tool", ""},
		{"Minikube", "minikube", "tool", ""},
		{"ESP-IDF", "espidf", "tool", ""},
		{"Android SDK", "android", "tool", ""},
		{"Redis", "redis", "database", "7"},
		{"MySQL", "mysql", "database", "8.0"},
	}
}

// BuildConfigFromOptions 从选项构建配置
func BuildConfigFromOptions(opts []InstallOption) *config.Config {
	cfg := &config.Config{
		Version: "0.2.0",
		Name:    "交互式选择环境",
	}
	for _, opt := range opts {
		switch opt.Category {
		case "language":
			cfg.Languages = append(cfg.Languages, config.Language{Name: opt.Key, Version: opt.Version})
		case "tool":
			cfg.Tools = append(cfg.Tools, opt.Key)
		case "database":
			cfg.Databases = append(cfg.Databases, config.Database{Name: opt.Key, Version: opt.Version, Docker: true})
		}
	}
	return cfg
}

// ParseInstallTargets 解析命令行安装目标
func ParseInstallTargets(targets []string) (*config.Config, []string) {
	var languages []config.Language
	var tools []string
	var databases []config.Database
	var unknown []string

	for _, target := range targets {
		switch target {
		case "node", "nodejs":
			languages = append(languages, config.Language{Name: "node", Version: "20.11.1"})
		case "python", "python3":
			languages = append(languages, config.Language{Name: "python", Version: "3.10.11"})
		case "go", "golang":
			languages = append(languages, config.Language{Name: "go", Version: "1.22.0"})
		case "rust":
			languages = append(languages, config.Language{Name: "rust", Version: "stable"})
		case "java", "jdk":
			languages = append(languages, config.Language{Name: "java", Version: "21"})
		case "bun":
			languages = append(languages, config.Language{Name: "bun", Version: "latest"})
		case "git":
			tools = append(tools, "git")
		case "docker":
			tools = append(tools, "docker")
		case "vscode", "code":
			tools = append(tools, "vscode")
		case "miniconda", "conda":
			tools = append(tools, "miniconda")
		case "kubectl":
			tools = append(tools, "kubectl")
		case "minikube":
			tools = append(tools, "minikube")
		case "espidf", "esp-idf":
			tools = append(tools, "espidf")
		case "android", "android-sdk":
			tools = append(tools, "android")
		case "postgresql", "postgres":
			databases = append(databases, config.Database{Name: "postgresql", Version: "16", Docker: true})
		case "redis":
			databases = append(databases, config.Database{Name: "redis", Version: "7", Docker: true})
		case "mysql":
			databases = append(databases, config.Database{Name: "mysql", Version: "8.0", Docker: true})
		case "mongodb", "mongo":
			databases = append(databases, config.Database{Name: "mongodb", Version: "6.0", Docker: true})
		default:
			unknown = append(unknown, target)
		}
	}

	if len(languages) == 0 && len(tools) == 0 && len(databases) == 0 {
		return nil, unknown
	}
	return &config.Config{
		Version:   "0.2.0",
		Name:      "命令行指定安装",
		Languages: languages,
		Tools:     tools,
		Databases: databases,
	}, unknown
}
