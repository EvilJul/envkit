package main

import (
	"fmt"
	"os"

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
	case "install":
		handleInstall()
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
	fmt.Scanln(&choice)

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
		ui.Error("配置文件不存在: %s", configFile)
		ui.Info("提示: 先运行 'envkit init' 生成配置文件")
		os.Exit(1)
	}

	// 解析配置
	parser := config.NewParser(configFile)
	cfg, err := parser.Parse()
	if err != nil {
		ui.Error("解析配置文件失败: %v", err)
		os.Exit(1)
	}

	ui.PrintHeader("正在安装: " + cfg.Name)

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
					}
				case "redis":
					if err := dockerMgr.StartRedis(db.Version); err != nil {
						ui.Error("启动失败: %v", err)
					}
				case "mysql":
					password := "mysql"
					if err := dockerMgr.StartMySQL(db.Version, password); err != nil {
						ui.Error("启动失败: %v", err)
					}
				case "mongodb", "mongo":
					if err := dockerMgr.StartMongoDB(db.Version); err != nil {
						ui.Error("启动失败: %v", err)
					}
				default:
					ui.Warning("不支持的数据库: %s", db.Name)
				}
			}
		}
	}

	ui.PrintSection("安装完成")
	ui.Success("开发环境配置完成！")
	fmt.Println()
	ui.Info("提示:")
	ui.Info("  - 运行 'envkit detect' 查看已安装的工具")
	ui.Info("  - 运行 'envkit docker list' 查看运行中的容器")
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
	fmt.Println("  install [-f file]           安装开发环境 (默认: dev-env.yaml)")
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
	fmt.Println("  envkit init                         # 生成配置文件")
	fmt.Println("  envkit install                      # 使用默认配置安装")
	fmt.Println("  envkit install -f custom.yaml       # 使用自定义配置")
	fmt.Println("  envkit detect                       # 检测系统环境")
	fmt.Println("  envkit mirror npm npmmirror         # 配置npm镜像源")
	fmt.Println("  envkit docker start postgres 16     # 启动 PostgreSQL")
	fmt.Println("  envkit docker list                  # 查看运行容器")
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
		fmt.Scanln(&answer)
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
