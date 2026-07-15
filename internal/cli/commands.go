package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/tui"
	"github.com/fusheng/envkit/internal/ui"
	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "检测当前系统已安装的工具",
	RunE: func(cmd *cobra.Command, args []string) error {
		PrintDetectReport(service.DetectEnvironment())
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有支持的一键安装环境与工具状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		PrintCatalog()
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("envkit version %s\n", Version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "交互式生成配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		templateN, _ := cmd.Flags().GetInt("template")
		if templateN > 0 {
			if err := service.InitFromTemplate(templateN, "dev-env.yaml"); err != nil {
				return err
			}
			ui.Success("已生成配置文件: dev-env.yaml")
			fmt.Println("\n下一步:")
			fmt.Println("  运行 'envkit install -f dev-env.yaml' 开始安装")
			return nil
		}
		if err := requireTTY("init"); err != nil {
			return err
		}
		return tui.RunInitWizard()
	},
}

var installCmd = &cobra.Command{
	Use:   "install [components...]",
	Short: "安装开发环境",
	RunE:  runInstallCmd,
}

func init() {
	initCmd.Flags().IntP("template", "t", 0, "非交互：模板编号 (1-3)")
	installCmd.Flags().StringP("file", "f", "", "配置文件路径")
}

func runInstallCmd(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("file")

	if len(args) > 0 {
		cfg, unknown := service.ParseInstallTargets(args)
		for _, u := range unknown {
			ui.Warning("未知组件: %s，已跳过", u)
		}
		if cfg == nil {
			return fmt.Errorf("未选择任何有效的组件进行安装")
		}
		result := service.RunInstallation(cfg)
		if len(result.FailedComponents) > 0 {
			return result
		}
		return nil
	}

	if configFile != "" {
		return installFromFile(configFile)
	}

	if _, err := os.Stat("dev-env.yaml"); err == nil {
		return installFromFile("dev-env.yaml")
	}

	if err := requireTTY("install"); err != nil {
		ui.Info("提示: envkit install node | envkit install -f dev-env.yaml")
		return err
	}
	return tui.RunInstallWizard()
}

func installFromFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		ui.Error("配置文件不存在: %s", path)
		ui.Info("提示: 先运行 'envkit init' 生成配置文件，或直接运行 'envkit install' 开启交互式安装")
		return fmt.Errorf("配置文件不存在: %s", path)
	}
	parser := config.NewParser(path)
	cfg, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	result := service.RunInstallation(cfg)
	if len(result.FailedComponents) > 0 {
		return result
	}
	return nil
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [component]",
	Short: "卸载已安装的组件及配置",
	RunE:  runUninstallCmd,
}

func init() {
	uninstallCmd.Flags().Bool("all", false, "卸载所有组件")
}

func runUninstallCmd(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool("all")

	items, err := service.LoadManifestItems()
	if err != nil {
		return fmt.Errorf("加载清单失败: %w", err)
	}
	if len(items) == 0 {
		ui.Info("清单中没有通过 EnvKit 安装的组件记录。")
		return nil
	}

	if allFlag || (len(args) > 0 && (args[0] == "--all" || args[0] == "-a")) {
		fmt.Print(ui.Red("警告: 您将卸载通过 EnvKit 安装的所有组件及配置。确定继续吗? (y/N): "))
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			ui.Info("操作已取消。")
			return nil
		}
		var errs []string
		for _, e := range service.UninstallAll() {
			ui.Error("%v", e)
			errs = append(errs, e.Error())
		}
		if len(errs) > 0 {
			return fmt.Errorf("卸载过程有错误: %s", strings.Join(errs, "; "))
		}
		return nil
	}

	if len(args) > 0 {
		name := strings.ToLower(args[0])
		if err := service.UninstallComponent(name); err != nil {
			return err
		}
		return nil
	}

	if err := requireTTY("uninstall"); err != nil {
		return err
	}
	return tui.RunUninstallWizard()
}

var mirrorCmd = &cobra.Command{
	Use:   "mirror <language> [mirror-name]",
	Short: "单独配置某个语言的镜像源",
	RunE:  runMirrorCmd,
}

func runMirrorCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println("用法: envkit mirror <language> [mirror-name]")
		fmt.Println()
		fmt.Println("支持的语言:")
		fmt.Println("  npm, pip, go, rust, gradle, android")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  envkit mirror npm npmmirror")
		fmt.Println("  envkit mirror pip tsinghua")
		return fmt.Errorf("缺少 language 参数")
	}

	if len(args) == 1 {
		if err := requireTTY("mirror"); err != nil {
			return err
		}
		return tui.RunMirrorWizard(args[0])
	}

	language := args[0]
	mirrorName := args[1]
	if err := service.ConfigureMirror(language, mirrorName); err != nil {
		return fmt.Errorf("配置失败: %w", err)
	}
	ui.Success("镜像源配置成功！")
	return nil
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "管理 Docker 容器",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if err := requireTTY("docker"); err != nil {
				printDockerUsage()
				return err
			}
			return tui.RunDockerMenu()
		}
		return runDockerSubcommand(args)
	},
}

func printDockerUsage() {
	fmt.Println("用法: envkit docker <subcommand>")
	fmt.Println()
	fmt.Println("子命令:")
	fmt.Println("  start <database> <version>  启动数据库容器")
	fmt.Println("  stop <container>            停止容器")
	fmt.Println("  list                        列出所有容器")
	fmt.Println("  remove <container>          删除容器")
}

func runDockerSubcommand(args []string) error {
	sub := args[0]
	switch sub {
	case "start":
		if len(args) < 3 {
			return fmt.Errorf("用法: envkit docker start <database> <version>")
		}
		if err := service.DockerStart(args[1], args[2]); err != nil {
			return err
		}
	case "stop":
		if len(args) < 2 {
			return fmt.Errorf("用法: envkit docker stop <container>")
		}
		if err := service.DockerStop(args[1]); err != nil {
			return err
		}
	case "list", "ls":
		if err := service.DockerList(); err != nil {
			return fmt.Errorf("获取容器列表失败: %w", err)
		}
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("用法: envkit docker remove <container>")
		}
		fmt.Print("是否同时删除数据卷? (y/N): ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		removeVolume := answer == "y" || answer == "Y"
		if err := service.DockerRemove(args[1], removeVolume); err != nil {
			return err
		}
	default:
		return fmt.Errorf("未知子命令: %s", sub)
	}
	return nil
}
