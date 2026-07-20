package cli

import (
	"fmt"
	"strings"

	"github.com/fusheng/envkit/internal/cli/service"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/ui"
)

// PrintDetectReport 打印检测报告
func PrintDetectReport(report service.DetectReport) {
	ui.PrintHeader("系统环境检测")
	ui.PrintSection("系统信息")
	fmt.Printf("操作系统: %s\n", ui.Cyan(string(report.System.OS)))
	fmt.Printf("架构: %s\n", ui.Cyan(string(report.System.Architecture)))
	if report.System.IsLinux() {
		fmt.Printf("发行版: %s\n", ui.Cyan(report.System.Distribution))
	}

	ui.PrintSection("已安装的编程语言")
	langTable := ui.NewTable("语言", "版本", "状态")
	for name, tool := range report.Languages {
		if tool.Installed {
			langTable.AddRow(name, tool.Version, ui.Green("✓ 已安装"))
		}
	}
	langTable.Render()

	ui.PrintSection("已安装的开发工具")
	toolTable := ui.NewTable("工具", "版本", "状态")
	for name, tool := range report.Tools {
		if tool.Installed {
			toolTable.AddRow(name, tool.Version, ui.Green("✓ 已安装"))
		}
	}
	toolTable.Render()

	ui.PrintSection("可用的包管理器")
	for name, tool := range report.Managers {
		if tool.Installed {
			ui.Success("%s", name)
		}
	}
}

// PrintCatalog 打印组件目录
func PrintCatalog() {
	languages := detector.DetectLanguages()
	tools := detector.DetectTools()

	ui.PrintHeader("支持的一键安装环境与开发工具列表")

	ui.PrintSection("编程语言 (Languages)")
	langTable := ui.NewTable("名称", "默认版本", "当前状态", "安装源 / 说明")
	langTable.AddRow("node", "20.11.1", statusANSI(languages["node"]), "Fast Node Manager (fnm)")
	langTable.AddRow("python", "3.10.11", statusANSI(languages["python"]), "Astral uv")
	langTable.AddRow("go", "1.22.0", statusANSI(languages["go"]), "官方包或 Homebrew")
	langTable.AddRow("rust", "stable", statusANSI(languages["rustc"]), "rustup 官方安装器")
	langTable.AddRow("java", "21", statusANSI(languages["java"]), "SDKMAN! 或 Microsoft OpenJDK")
	langTable.AddRow("bun", "latest", statusANSI(languages["bun"]), "Bun 官方安装器")
	langTable.Render()

	ui.PrintSection("开发工具 (Tools)")
	toolTable := ui.NewTable("名称", "当前状态", "说明")
	toolTable.AddRow("git", statusANSI(tools["git"]), "版本控制系统")
	toolTable.AddRow("docker", statusANSI(tools["docker"]), "应用容器引擎")
	toolTable.AddRow("vscode", statusANSI(tools["code"]), "Visual Studio Code 编辑器")
	toolTable.AddRow("uv", statusANSI(tools["uv"]), "Astral uv（Python 包/项目管理器）")
	toolTable.AddRow("miniconda", statusANSI(tools["conda"]), "Conda 虚拟环境与包管理器 (清华源)")
	toolTable.AddRow("kubectl", statusANSI(tools["kubectl"]), "Kubernetes 命令行控制工具")
	toolTable.AddRow("minikube", statusANSI(tools["minikube"]), "本地单节点 Kubernetes 集群运行工具")
	toolTable.AddRow("espidf", statusANSI(tools["espidf"]), "ESP-IDF (EIM) 乐鑫物联网开发框架")
	toolTable.AddRow("android", statusANSI(tools["android"]), "Android SDK (cmdline-tools/platform-tools/build-tools)")
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

func statusANSI(tool *detector.Tool) string {
	if tool != nil && tool.Installed {
		v := tool.Version
		if len(v) > 40 {
			v = v[:37] + "..."
		}
		return ui.Green("✓ 已安装 (" + v + ")")
	}
	return ui.Yellow("✗ 未安装")
}

// JoinErrors 合并错误信息
func JoinErrors(errs []error) string {
	var parts []string
	for _, e := range errs {
		if e != nil {
			parts = append(parts, e.Error())
		}
	}
	return strings.Join(parts, "; ")
}
