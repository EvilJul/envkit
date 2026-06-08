package main

import (
	"fmt"
	"os"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/mirror"
	"github.com/fusheng/envkit/internal/templates"
)

const (
	version = "0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "install":
		handleInstall()
	case "detect":
		handleDetect()
	case "mirror":
		handleMirror()
	case "version":
		fmt.Printf("envkit version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit() {
	// 显示模板选择
	tmpl := templates.New()
	list := tmpl.List()

	fmt.Println("📦 EnvKit - 开发环境快速配置工具")
	fmt.Println("\n请选择预设模板:")
	fmt.Println()

	for i, info := range list {
		fmt.Printf("  %d) %s\n", i+1, info.Name)
		fmt.Printf("     %s\n", info.Description)
		fmt.Printf("     语言: %v\n", info.Languages)
		fmt.Println()
	}

	fmt.Print("请输入选项 (1-3): ")
	var choice int
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(list) {
		fmt.Println("❌ 无效的选项")
		os.Exit(1)
	}

	selectedTemplate := list[choice-1]
	cfg, err := tmpl.Get(selectedTemplate.Type)
	if err != nil {
		fmt.Printf("❌ 加载模板失败: %v\n", err)
		os.Exit(1)
	}

	// 保存配置文件
	outputFile := "dev-env.yaml"
	if err := config.Export(cfg, outputFile); err != nil {
		fmt.Printf("❌ 保存配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ 已生成配置文件: %s\n", outputFile)
	fmt.Println("\n下一步:")
	fmt.Printf("  运行 'envkit install -f %s' 开始安装\n", outputFile)
}

func handleInstall() {
	configFile := "dev-env.yaml"

	// 解析命令行参数
	for i, arg := range os.Args {
		if arg == "-f" || arg == "--file" {
			if i+1 < len(os.Args) {
				configFile = os.Args[i+1]
			}
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("❌ 配置文件不存在: %s\n", configFile)
		fmt.Println("提示: 先运行 'envkit init' 生成配置文件")
		os.Exit(1)
	}

	// 解析配置
	parser := config.NewParser(configFile)
	cfg, err := parser.Parse()
	if err != nil {
		fmt.Printf("❌ 解析配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📦 正在安装: %s\n", cfg.Name)
	fmt.Println()

	// 检测系统
	sysInfo := detector.DetectSystem()
	fmt.Printf("系统信息: %s %s\n", sysInfo.OS, sysInfo.Architecture)
	if sysInfo.IsLinux() {
		fmt.Printf("Linux发行版: %s\n", sysInfo.Distribution)
	}
	fmt.Println()

	// 配置镜像源
	registry := mirror.NewRegistry()

	for _, lang := range cfg.Languages {
		fmt.Printf("⚙️  配置 %s 镜像源...\n", lang.Name)

		switch lang.Name {
		case "node", "nodejs":
			configurator := mirror.NewNPMConfigurator(registry)
			if err := configurator.Configure(lang.Mirror); err != nil {
				fmt.Printf("   ⚠️  警告: %v\n", err)
			}
		case "python":
			configurator := mirror.NewPipConfigurator(registry)
			if err := configurator.Configure(lang.Mirror); err != nil {
				fmt.Printf("   ⚠️  警告: %v\n", err)
			}
		case "go", "golang":
			configurator := mirror.NewGoConfigurator(registry)
			if err := configurator.Configure(lang.Mirror); err != nil {
				fmt.Printf("   ⚠️  警告: %v\n", err)
			}
		case "rust":
			configurator := mirror.NewRustConfigurator(registry)
			if err := configurator.Configure(lang.Mirror); err != nil {
				fmt.Printf("   ⚠️  警告: %v\n", err)
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ 镜像源配置完成!")
	fmt.Println()
	fmt.Println("⚠️  注意: 语言和工具的实际安装功能正在开发中")
	fmt.Println("当前版本仅配置国内镜像源")
}

func handleDetect() {
	fmt.Println("🔍 检测系统环境...")
	fmt.Println()

	// 系统信息
	sysInfo := detector.DetectSystem()
	fmt.Printf("操作系统: %s\n", sysInfo.OS)
	fmt.Printf("架构: %s\n", sysInfo.Architecture)
	if sysInfo.IsLinux() {
		fmt.Printf("发行版: %s\n", sysInfo.Distribution)
	}
	fmt.Println()

	// 检测语言
	fmt.Println("📚 已安装的编程语言:")
	languages := detector.DetectLanguages()
	for name, tool := range languages {
		if tool.Installed {
			fmt.Printf("  ✓ %s: %s\n", name, tool.Version)
		}
	}
	fmt.Println()

	// 检测工具
	fmt.Println("🛠️  已安装的开发工具:")
	tools := detector.DetectTools()
	for name, tool := range tools {
		if tool.Installed {
			fmt.Printf("  ✓ %s: %s\n", name, tool.Version)
		}
	}
	fmt.Println()

	// 检测包管理器
	fmt.Println("📦 可用的包管理器:")
	managers := detector.DetectPackageManagers()
	for name, tool := range managers {
		if tool.Installed {
			fmt.Printf("  ✓ %s\n", name)
		}
	}
}

func handleMirror() {
	if len(os.Args) < 3 {
		fmt.Println("用法: envkit mirror <language> [mirror-name]")
		fmt.Println()
		fmt.Println("支持的语言:")
		fmt.Println("  npm, pip, go, rust")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  envkit mirror npm npmmirror")
		fmt.Println("  envkit mirror pip tsinghua")
		fmt.Println("  envkit mirror go goproxy")
		os.Exit(1)
	}

	language := os.Args[2]
	mirrorName := ""
	if len(os.Args) > 3 {
		mirrorName = os.Args[3]
	}

	registry := mirror.NewRegistry()

	switch language {
	case "npm":
		configurator := mirror.NewNPMConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			fmt.Printf("❌ 配置失败: %v\n", err)
			os.Exit(1)
		}
	case "pip":
		configurator := mirror.NewPipConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			fmt.Printf("❌ 配置失败: %v\n", err)
			os.Exit(1)
		}
	case "go":
		configurator := mirror.NewGoConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			fmt.Printf("❌ 配置失败: %v\n", err)
			os.Exit(1)
		}
	case "rust":
		configurator := mirror.NewRustConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			fmt.Printf("❌ 配置失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 不支持的语言: %s\n", language)
		os.Exit(1)
	}

	fmt.Println("✅ 镜像源配置成功!")
}

func printUsage() {
	fmt.Println("EnvKit - 开发环境快速配置工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  envkit <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  init                    交互式生成配置文件")
	fmt.Println("  install [-f file]       安装开发环境 (默认: dev-env.yaml)")
	fmt.Println("  detect                  检测当前系统已安装的工具")
	fmt.Println("  mirror <lang> [name]    单独配置某个语言的镜像源")
	fmt.Println("  version                 显示版本信息")
	fmt.Println("  help                    显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  envkit init                    # 生成配置文件")
	fmt.Println("  envkit install                 # 使用默认配置安装")
	fmt.Println("  envkit install -f custom.yaml  # 使用自定义配置")
	fmt.Println("  envkit detect                  # 检测系统环境")
	fmt.Println("  envkit mirror npm npmmirror    # 配置npm镜像源")
}
