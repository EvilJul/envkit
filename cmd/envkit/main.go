package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
	"github.com/fusheng/envkit/internal/templates"
	"github.com/fusheng/envkit/internal/ui"
)

const (
	version = "0.2.0"
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
	case "list":
		handleList()
	case "install":
		handleInstall()
	case "uninstall":
		handleUninstall()
	case "detect":
		handleDetect()
	case "mirror":
		handleMirror()
	case "docker":
		handleDocker()
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

	ui.PrintHeader("EnvKit - 开发环境快速配置工具")
	fmt.Println("请选择预设模板:")
	fmt.Println()

	for i, info := range list {
		fmt.Printf("  %d) %s\n", i+1, info.Name)
		fmt.Printf("     %s\n", info.Description)
		fmt.Printf("     语言: %v\n", info.Languages)
		fmt.Println()
	}

	fmt.Print("请输入选项 (1-3): ")
	var choice int
	_, _ = fmt.Scanln(&choice)

	if choice < 1 || choice > len(list) {
		ui.Error("无效的选项")
		os.Exit(1)
	}

	selectedTemplate := list[choice-1]
	cfg, err := tmpl.Get(selectedTemplate.Type)
	if err != nil {
		ui.Error("加载模板失败: %v", err)
		os.Exit(1)
	}

	// 保存配置文件
	outputFile := "dev-env.yaml"
	if err := config.Export(cfg, outputFile); err != nil {
		ui.Error("保存配置文件失败: %v", err)
		os.Exit(1)
	}

	ui.Success("已生成配置文件: %s", outputFile)
	fmt.Println("\n下一步:")
	fmt.Printf("  运行 'envkit install -f %s' 开始安装\n", outputFile)
}

func handleInstall() {
	configFile := ""
	hasFileArg := false

	// 1. 优先解析直接在命令行参数中指定的安装目标 (例如 ./envkit install node espidf)
	var directTargets []string
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			if (arg == "-f" || arg == "--file") && i+1 < len(os.Args) {
				i++ // 跳过文件名参数
			}
			continue
		}
		directTargets = append(directTargets, arg)
	}

	// 如果指定了单项或多项直接安装目标
	if len(directTargets) > 0 {
		var languages []config.Language
		var tools []string
		var databases []config.Database

		for _, target := range directTargets {
			target = strings.ToLower(target)
			switch target {
			// 语言类
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

			// 工具类
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

			// 数据库类
			case "postgresql", "postgres":
				databases = append(databases, config.Database{Name: "postgresql", Version: "16", Docker: true})
			case "redis":
				databases = append(databases, config.Database{Name: "redis", Version: "7", Docker: true})
			case "mysql":
				databases = append(databases, config.Database{Name: "mysql", Version: "8.0", Docker: true})
			case "mongodb", "mongo":
				databases = append(databases, config.Database{Name: "mongodb", Version: "6.0", Docker: true})

			default:
				ui.Warning("未知组件: %s，已跳过", target)
			}
		}

		if len(languages) > 0 || len(tools) > 0 || len(databases) > 0 {
			cfg := &config.Config{
				Version:   "0.2.0",
				Name:      "命令行指定安装",
				Languages: languages,
				Tools:     tools,
				Databases: databases,
			}
			runInstallation(cfg)
			return
		}
		ui.Error("未选择任何有效的组件进行安装。")
		os.Exit(1)
	}

	// 2. 解析配置文件命令行参数
	for i, arg := range os.Args {
		if arg == "-f" || arg == "--file" {
			if i+1 < len(os.Args) {
				configFile = os.Args[i+1]
				hasFileArg = true
			}
		}
	}

	// 如果没有指定配置文件参数，则尝试使用默认文件，没有默认文件则开启交互式安装
	if !hasFileArg {
		if _, err := os.Stat("dev-env.yaml"); err == nil {
			configFile = "dev-env.yaml"
		} else {
			handleInteractiveInstall()
			return
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		ui.Error("配置文件不存在: %s", configFile)
		ui.Info("提示: 先运行 'envkit init' 生成配置文件，或直接运行 'envkit install' 开启交互式安装")
		os.Exit(1)
	}

	// 解析配置
	parser := config.NewParser(configFile)
	cfg, err := parser.Parse()
	if err != nil {
		ui.Error("解析配置文件失败: %v", err)
		os.Exit(1)
	}

	runInstallation(cfg)
}

func handleInteractiveInstall() {
	ui.PrintHeader("交互式开发环境安装程序")
	fmt.Println("请选择您要安装的组件 (多选，输入数字编号并用空格或逗号分隔，如 '1 3'):")
	fmt.Println()

	options := []struct {
		name     string
		key      string
		category string // "language", "tool", "database"
		version  string
	}{
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
		{"Redis", "redis", "database", "7"},
		{"MySQL", "mysql", "database", "8.0"},
	}

	for i, opt := range options {
		typeStr := "工具"
		if opt.category == "language" {
			typeStr = "语言"
		} else if opt.category == "database" {
			typeStr = "数据库"
		}
		if opt.version != "" {
			fmt.Printf("  %2d) %-12s (%s) [默认版本: %s]\n", i+1, opt.name, typeStr, opt.version)
		} else {
			fmt.Printf("  %2d) %-12s (%s)\n", i+1, opt.name, typeStr)
		}
	}
	fmt.Println()

	fmt.Print("请输入选项 (如 '1, 3 5'): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	input = strings.ReplaceAll(input, ",", " ")
	tokens := strings.Fields(input)

	var languages []config.Language
	var tools []string
	var databases []config.Database

	for _, token := range tokens {
		var index int
		_, err := fmt.Sscanf(token, "%d", &index)
		if err != nil || index < 1 || index > len(options) {
			continue
		}

		opt := options[index-1]
		if opt.category == "language" {
			languages = append(languages, config.Language{
				Name:    opt.key,
				Version: opt.version,
			})
		} else if opt.category == "tool" {
			tools = append(tools, opt.key)
		} else if opt.category == "database" {
			databases = append(databases, config.Database{
				Name:    opt.key,
				Version: opt.version,
				Docker:  true,
			})
		}
	}

	if len(languages) == 0 && len(tools) == 0 && len(databases) == 0 {
		ui.Warning("您没有选择任何有效的组件进行安装。")
		return
	}

	// 确认安装
	fmt.Println()
	ui.PrintSection("待安装的组件")
	for _, lang := range languages {
		fmt.Printf("  - %s (版本: %s)\n", ui.Cyan(lang.Name), lang.Version)
	}
	for _, tool := range tools {
		fmt.Printf("  - %s\n", ui.Cyan(tool))
	}
	for _, db := range databases {
		fmt.Printf("  - %s (数据库容器, 版本: %s)\n", ui.Cyan(db.Name), db.Version)
	}
	fmt.Println()

	fmt.Print("是否开始安装? (y/N): ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		ui.Info("安装已取消。")
		return
	}

	cfg := &config.Config{
		Version:   "0.2.0",
		Name:      "交互式选择环境",
		Languages: languages,
		Tools:     tools,
		Databases: databases,
	}

	runInstallation(cfg)
}

func runInstallation(cfg *config.Config) {
	ui.PrintHeader("正在安装: " + cfg.Name)

	var failedComponents []string

	// 检测系统
	sysInfo := detector.DetectSystem()
	ui.PrintSection("系统信息")
	fmt.Printf("操作系统: %s\n", ui.Cyan(string(sysInfo.OS)))
	fmt.Printf("架构: %s\n", ui.Cyan(string(sysInfo.Architecture)))
	if sysInfo.IsLinux() {
		fmt.Printf("发行版: %s\n", ui.Cyan(sysInfo.Distribution))
	}

	// 1. 安装编程语言
	if len(cfg.Languages) > 0 {
		ui.PrintSection("安装编程语言")
		for _, lang := range cfg.Languages {
			langInstaller := installer.GetInstaller(lang.Name)
			if langInstaller == nil {
				ui.Warning("不支持的语言: %s", lang.Name)
				continue
			}

			// 检查是否已安装
			if langInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", lang.Name, langInstaller.GetVersion())
			} else {
				ui.Info("正在安装 %s %s...", lang.Name, lang.Version)
				if err := langInstaller.Install(lang.Version); err != nil {
					ui.Error("安装 %s 失败: %v", lang.Name, err)
					failedComponents = append(failedComponents, lang.Name)
					continue
				}
				ui.Success("%s 安装成功！", lang.Name)
			}

			// 配置镜像源
			if lang.Mirror != "" {
				ui.Info("配置 %s 镜像源: %s", lang.Name, lang.Mirror)
				registry := mirror.NewRegistry()

				switch lang.Name {
				case "node", "nodejs":
					configurator := mirror.NewNPMConfigurator(registry)
					if err := configurator.Configure(lang.Mirror); err != nil {
						ui.Warning("配置镜像源失败: %v", err)
					} else {
						ui.Success("镜像源配置成功")
					}
				case "python":
					configurator := mirror.NewPipConfigurator(registry)
					if err := configurator.Configure(lang.Mirror); err != nil {
						ui.Warning("配置镜像源失败: %v", err)
					} else {
						ui.Success("镜像源配置成功")
					}
				case "go", "golang":
					configurator := mirror.NewGoConfigurator(registry)
					if err := configurator.Configure(lang.Mirror); err != nil {
						ui.Warning("配置镜像源失败: %v", err)
					} else {
						ui.Success("镜像源配置成功")
					}
				case "rust":
					configurator := mirror.NewRustConfigurator(registry)
					if err := configurator.Configure(lang.Mirror); err != nil {
						ui.Warning("配置镜像源失败: %v", err)
					} else {
						ui.Success("镜像源配置成功")
					}
				}
			}
		}
	}

	// 2. 安装开发工具
	if len(cfg.Tools) > 0 {
		ui.PrintSection("安装开发工具")
		for _, tool := range cfg.Tools {
			toolInstaller := installer.GetToolInstaller(tool)
			if toolInstaller == nil {
				ui.Warning("不支持的工具: %s", tool)
				continue
			}

			// 检查是否已安装
			if toolInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", tool, toolInstaller.GetVersion())
			} else {
				ui.Info("正在安装 %s...", tool)
				if err := toolInstaller.Install(); err != nil {
					ui.Error("安装 %s 失败: %v", tool, err)
					failedComponents = append(failedComponents, tool)
					continue
				}
				ui.Success("%s 安装成功！", tool)
			}
		}
	}

	// 3. 启动数据库容器
	if len(cfg.Databases) > 0 {
		ui.PrintSection("启动数据库容器")
		dockerMgr := docker.NewContainerManager()

		// 检查 Docker 是否运行
		if !dockerMgr.IsDockerRunning() {
			ui.Warning("Docker 未运行，跳过数据库容器启动")
			ui.Info("请先启动 Docker，然后运行 'envkit docker start' 启动数据库")
		} else {
			for _, db := range cfg.Databases {
				if !db.Docker {
					ui.Info("跳过 %s (docker: false)", db.Name)
					continue
				}

				ui.Info("正在启动 %s %s...", db.Name, db.Version)

				switch db.Name {
				case "postgresql", "postgres":
					password := "postgres"
					if err := dockerMgr.StartPostgreSQL(db.Version, password); err != nil {
						ui.Error("启动失败: %v", err)
						failedComponents = append(failedComponents, db.Name)
					}
				case "redis":
					if err := dockerMgr.StartRedis(db.Version); err != nil {
						ui.Error("启动失败: %v", err)
						failedComponents = append(failedComponents, db.Name)
					}
				case "mysql":
					password := "mysql"
					if err := dockerMgr.StartMySQL(db.Version, password); err != nil {
						ui.Error("启动失败: %v", err)
						failedComponents = append(failedComponents, db.Name)
					}
				case "mongodb", "mongo":
					if err := dockerMgr.StartMongoDB(db.Version); err != nil {
						ui.Error("启动失败: %v", err)
						failedComponents = append(failedComponents, db.Name)
					}
				default:
					ui.Warning("不支持的数据库: %s", db.Name)
				}
			}
		}
	}

	ui.PrintSection("安装完成")
	if len(failedComponents) > 0 {
		ui.Warning("开发环境配置部分组件安装或启动失败:")
		for _, comp := range failedComponents {
			fmt.Printf("  - %s\n", ui.Red(comp))
		}
	} else {
		ui.Success("开发环境配置完成！")
	}
	fmt.Println()
	ui.Info("提示:")
	ui.Info("  - 运行 'envkit detect' 查看已安装的工具")
	ui.Info("  - 运行 'envkit docker list' 查看运行中的容器")
}

func handleList() {
	ui.PrintHeader("支持的一键安装环境与开发工具列表")

	languages := detector.DetectLanguages()
	tools := detector.DetectTools()

	ui.PrintSection("编程语言 (Languages)")
	langTable := ui.NewTable("名称", "默认版本", "当前状态", "安装源 / 说明")
	langTable.AddRow("node", "20.11.1", getStatusString(languages["node"]), "Fast Node Manager (fnm)")
	langTable.AddRow("python", "3.10.11", getStatusString(languages["python"]), "Astral uv")
	langTable.AddRow("go", "1.22.0", getStatusString(languages["go"]), "官方包或 Homebrew")
	langTable.AddRow("rust", "stable", getStatusString(languages["rustc"]), "rustup 官方安装器")
	langTable.AddRow("java", "21", getStatusString(languages["java"]), "SDKMAN! 或 Microsoft OpenJDK")
	langTable.AddRow("bun", "latest", getStatusString(languages["bun"]), "Bun 官方安装器")
	langTable.Render()

	ui.PrintSection("开发工具 (Tools)")
	toolTable := ui.NewTable("名称", "当前状态", "说明")
	toolTable.AddRow("git", getStatusString(tools["git"]), "版本控制系统")
	toolTable.AddRow("docker", getStatusString(tools["docker"]), "应用容器引擎")
	toolTable.AddRow("vscode", getStatusString(tools["code"]), "Visual Studio Code 编辑器")
	toolTable.AddRow("miniconda", getStatusString(tools["conda"]), "Conda 虚拟环境与包管理器 (清华源)")
	toolTable.AddRow("kubectl", getStatusString(tools["kubectl"]), "Kubernetes 命令行控制工具")
	toolTable.AddRow("minikube", getStatusString(tools["minikube"]), "本地单节点 Kubernetes 集群运行工具")
	toolTable.AddRow("espidf", getStatusString(tools["espidf"]), "ESP-IDF (EIM) 乐鑫物联网开发框架")
	toolTable.Render()

	ui.PrintSection("数据库容器 (Databases via Docker)")
	dbTable := ui.NewTable("名称", "默认版本", "说明")
	dbTable.AddRow("postgresql", "16", "PostgreSQL 关系型数据库")
	dbTable.AddRow("redis", "7", "Redis 键值内存数据库")
	dbTable.AddRow("mysql", "8.0", "MySQL 关系型数据库")
	dbTable.AddRow("mongodb", "6.0", "MongoDB 文档型数据库")
	dbTable.Render()

	fmt.Println()
	ui.Info("提示:")
	ui.Info("  - 运行 'envkit install' 可进入交互式菜单，选择单/多语言及工具进行一键安装")
	ui.Info("  - 运行 'envkit install -f <config.yaml>' 按照配置文件进行非交互式安装")
}

func getStatusString(tool *detector.Tool) string {
	if tool != nil && tool.Installed {
		return ui.Green("✓ 已安装 (" + tool.Version + ")")
	}
	return ui.Yellow("✗ 未安装")
}

func handleDetect() {
	ui.PrintHeader("系统环境检测")

	// 系统信息
	sysInfo := detector.DetectSystem()
	ui.PrintSection("系统信息")
	fmt.Printf("操作系统: %s\n", ui.Cyan(string(sysInfo.OS)))
	fmt.Printf("架构: %s\n", ui.Cyan(string(sysInfo.Architecture)))
	if sysInfo.IsLinux() {
		fmt.Printf("发行版: %s\n", ui.Cyan(sysInfo.Distribution))
	}

	// 检测语言
	ui.PrintSection("已安装的编程语言")
	languages := detector.DetectLanguages()

	langTable := ui.NewTable("语言", "版本", "状态")
	for name, tool := range languages {
		if tool.Installed {
			langTable.AddRow(name, tool.Version, ui.Green("✓ 已安装"))
		}
	}
	langTable.Render()

	// 检测工具
	ui.PrintSection("已安装的开发工具")
	tools := detector.DetectTools()

	toolTable := ui.NewTable("工具", "版本", "状态")
	for name, tool := range tools {
		if tool.Installed {
			toolTable.AddRow(name, tool.Version, ui.Green("✓ 已安装"))
		}
	}
	toolTable.Render()

	// 检测包管理器
	ui.PrintSection("可用的包管理器")
	managers := detector.DetectPackageManagers()
	for name, tool := range managers {
		if tool.Installed {
			ui.Success("%s", name)
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
			ui.Error("配置失败: %v", err)
			os.Exit(1)
		}
	case "pip":
		configurator := mirror.NewPipConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			ui.Error("配置失败: %v", err)
			os.Exit(1)
		}
	case "go":
		configurator := mirror.NewGoConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			ui.Error("配置失败: %v", err)
			os.Exit(1)
		}
	case "rust":
		configurator := mirror.NewRustConfigurator(registry)
		if err := configurator.Configure(mirrorName); err != nil {
			ui.Error("配置失败: %v", err)
			os.Exit(1)
		}
	default:
		ui.Error("不支持的语言: %s", language)
		os.Exit(1)
	}

	ui.Success("镜像源配置成功！")
}

func printUsage() {
	ui.PrintHeader("EnvKit - 开发环境快速配置工具")

	fmt.Println("用法:")
	fmt.Println("  envkit <command> [options]")
	fmt.Println()

	fmt.Println(ui.Bold("命令:"))
	fmt.Println("  init                        交互式生成配置文件")
	fmt.Println("  list                        列出所有支持的一键安装环境与工具状态")
	fmt.Println("  install [components...]     安装开发环境 (支持直接指定单/多个组件，无参时开启交互式)")
	fmt.Println("  uninstall [component]       卸载已安装的组件及配置 (不指定组件时开启交互式卸载)")
	fmt.Println("  detect                      检测当前系统已安装的工具")
	fmt.Println("  mirror <lang> [name]        单独配置某个语言的镜像源")
	fmt.Println("  docker <subcommand>         管理 Docker 容器")
	fmt.Println("  version                     显示版本信息")
	fmt.Println("  help                        显示帮助信息")
	fmt.Println()

	fmt.Println(ui.Bold("Docker 子命令:"))
	fmt.Println("  start <db> <version>        启动数据库容器")
	fmt.Println("  stop <container>            停止容器")
	fmt.Println("  list                        列出所有容器")
	fmt.Println("  remove <container>          删除容器")
	fmt.Println()

	fmt.Println(ui.Bold("示例:"))
	fmt.Println("  ./envkit init                         # 生成配置文件")
	fmt.Println("  ./envkit list                         # 列出支持的环境列表")
	fmt.Println("  ./envkit install                      # 交互式选择组件安装")
	fmt.Println("  ./envkit install node                 # 直接一键安装单个组件 (如 Node.js)")
	fmt.Println("  ./envkit install go rust redis        # 一键安装指定的多个语言与工具")
	fmt.Println("  ./envkit install -f custom.yaml       # 使用自定义配置文件安装")
	fmt.Println("  ./envkit uninstall                    # 交互式选择组件卸载")
	fmt.Println("  ./envkit uninstall node               # 卸载 Node.js 并清理环境变量")
	fmt.Println("  ./envkit uninstall --all              # 卸载所有组件并清理环境配置")
	fmt.Println("  ./envkit detect                       # 检测系统环境")
	fmt.Println("  ./envkit mirror npm npmmirror         # 配置npm镜像源")
	fmt.Println("  ./envkit docker start postgres 16     # 启动 PostgreSQL")
	fmt.Println("  ./envkit docker list                  # 查看运行容器")
}

func handleDocker() {
	if len(os.Args) < 3 {
		fmt.Println("用法: envkit docker <subcommand>")
		fmt.Println()
		fmt.Println("子命令:")
		fmt.Println("  start <database> <version>  启动数据库容器")
		fmt.Println("  stop <container>            停止容器")
		fmt.Println("  list                        列出所有容器")
		fmt.Println("  remove <container>          删除容器")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  envkit docker start postgres 16")
		fmt.Println("  envkit docker start redis 7")
		fmt.Println("  envkit docker list")
		fmt.Println("  envkit docker stop envkit-postgres")
		os.Exit(1)
	}

	subcommand := os.Args[2]
	dockerMgr := docker.NewContainerManager()

	switch subcommand {
	case "start":
		if len(os.Args) < 5 {
			ui.Error("用法: envkit docker start <database> <version>")
			os.Exit(1)
		}
		database := os.Args[3]
		version := os.Args[4]

		if !dockerMgr.IsDockerRunning() {
			ui.Error("Docker 未运行，请先启动 Docker")
			os.Exit(1)
		}

		switch database {
		case "postgresql", "postgres":
			password := "postgres"
			if err := dockerMgr.StartPostgreSQL(version, password); err != nil {
				ui.Error("启动失败: %v", err)
				os.Exit(1)
			}
		case "redis":
			if err := dockerMgr.StartRedis(version); err != nil {
				ui.Error("启动失败: %v", err)
				os.Exit(1)
			}
		case "mysql":
			password := "mysql"
			if err := dockerMgr.StartMySQL(version, password); err != nil {
				ui.Error("启动失败: %v", err)
				os.Exit(1)
			}
		case "mongodb", "mongo":
			if err := dockerMgr.StartMongoDB(version); err != nil {
				ui.Error("启动失败: %v", err)
				os.Exit(1)
			}
		default:
			ui.Error("不支持的数据库: %s", database)
			ui.Info("支持的数据库: postgres, redis, mysql, mongodb")
			os.Exit(1)
		}

	case "stop":
		if len(os.Args) < 4 {
			ui.Error("用法: envkit docker stop <container>")
			os.Exit(1)
		}
		containerName := os.Args[3]
		if err := dockerMgr.StopContainer(containerName); err != nil {
			ui.Error("停止失败: %v", err)
			os.Exit(1)
		}

	case "list", "ls":
		if err := dockerMgr.ListContainers(); err != nil {
			ui.Error("获取容器列表失败: %v", err)
			os.Exit(1)
		}

	case "remove", "rm":
		if len(os.Args) < 4 {
			ui.Error("用法: envkit docker remove <container>")
			os.Exit(1)
		}
		containerName := os.Args[3]

		// 询问是否删除数据卷
		fmt.Print("是否同时删除数据卷? (y/N): ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		removeVolume := answer == "y" || answer == "Y"

		if err := dockerMgr.RemoveContainer(containerName, removeVolume); err != nil {
			ui.Error("删除失败: %v", err)
			os.Exit(1)
		}

	default:
		ui.Error("未知子命令: %s", subcommand)
		os.Exit(1)
	}
}

func handleUninstall() {
	m, err := installer.LoadManifest()
	if err != nil {
		ui.Error("加载清单失败: %v", err)
		os.Exit(1)
	}

	if len(m.Items) == 0 {
		ui.Info("清单中没有通过 EnvKit 安装的组件记录。")
		return
	}

	// 1. 如果提供了参数
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "--all" || arg == "-a" {
			// 询问确认
			fmt.Print(ui.Red("警告: 您将卸载通过 EnvKit 安装的所有组件及配置。确定继续吗? (y/N): "))
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				ui.Info("操作已取消。")
				return
			}

			// 开始全部卸载
			var names []string
			for name := range m.Items {
				names = append(names, name)
			}

			for _, name := range names {
				if err := installer.UninstallComponent(name); err != nil {
					ui.Error("卸载 %s 失败: %v", name, err)
				}
			}
			return
		}

		// 卸载指定组件
		name := strings.ToLower(arg)
		if err := installer.UninstallComponent(name); err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}
		return
	}

	// 2. 交互式卸载菜单
	ui.PrintHeader("EnvKit 卸载中心")
	fmt.Println("检测到以下通过 EnvKit 安装的组件:")
	fmt.Println()

	var names []string
	for name := range m.Items {
		names = append(names, name)
	}

	for i, name := range names {
		item := m.Items[name]
		typeStr := "工具"
		if item.Type == "language" {
			typeStr = "语言"
		}
		fmt.Printf("  %d) %-10s (%s) [版本: %s, 安装时间: %s]\n",
			i+1, name, typeStr, item.Version, item.InstalledAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("  %d) 卸载所有组件 (--all)\n", len(names)+1)
	fmt.Println()

	fmt.Print("请输入要卸载的组件选项 (如 '1, 2'，直接回车取消): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, ",", " ")
	tokens := strings.Fields(input)

	if len(tokens) == 0 {
		ui.Info("未选择任何组件。")
		return
	}

	var toUninstall []string
	uninstallAll := false

	for _, token := range tokens {
		var index int
		_, err := fmt.Sscanf(token, "%d", &index)
		if err != nil || index < 1 || index > len(names)+1 {
			continue
		}

		if index == len(names)+1 {
			uninstallAll = true
			break
		}
		toUninstall = append(toUninstall, names[index-1])
	}

	if uninstallAll {
		fmt.Print(ui.Red("警告: 您将卸载通过 EnvKit 安装的所有组件及配置。确定继续吗? (y/N): "))
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			ui.Info("操作已取消。")
			return
		}
		for _, name := range names {
			if err := installer.UninstallComponent(name); err != nil {
				ui.Error("卸载 %s 失败: %v", name, err)
			}
		}
		return
	}

	if len(toUninstall) == 0 {
		ui.Warning("您没有选择任何有效的组件进行卸载。")
		return
	}

	fmt.Println()
	ui.PrintSection("待卸载的组件")
	for _, name := range toUninstall {
		fmt.Printf("  - %s\n", ui.Red(name))
	}
	fmt.Println()

	fmt.Print("是否开始卸载? (y/N): ")
	var confirm string
	_, _ = fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		ui.Info("卸载已取消。")
		return
	}

	for _, name := range toUninstall {
		if err := installer.UninstallComponent(name); err != nil {
			ui.Error("卸载 %s 失败: %v", name, err)
		}
	}
}
