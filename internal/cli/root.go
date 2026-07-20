package cli

import (
	"fmt"
	"os"

	"github.com/fusheng/envkit/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Version 由发布构建通过 -ldflags -X 注入；本地默认与最新 release 对齐
var Version = "0.2.2"

var launchTUI bool

// Execute 运行 CLI
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "envkit",
	Short: "EnvKit - 开发环境快速配置工具",
	Long:  "轻量级跨平台开发环境管理工具，支持语言/工具一键安装与检测。",
	RunE: func(cmd *cobra.Command, args []string) error {
		if shouldLaunchTUI(cmd) {
			return tui.Run()
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&launchTUI, "tui", false, "启动 Bubble Tea 交互界面")
	rootCmd.AddCommand(
		initCmd,
		listCmd,
		detectCmd,
		installCmd,
		uninstallCmd,
		mirrorCmd,
		dockerCmd,
		versionCmd,
	)
}

func shouldLaunchTUI(cmd *cobra.Command) bool {
	if launchTUI {
		return isTTY()
	}
	if len(os.Args) == 1 && isTTY() {
		return true
	}
	return false
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func requireTTY(action string) error {
	if !isTTY() {
		return fmt.Errorf("%s 需要交互式终端，请使用 TTY 运行或指定命令行参数", action)
	}
	return nil
}
