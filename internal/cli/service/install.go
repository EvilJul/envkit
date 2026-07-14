package service

import (
	"fmt"
	"strings"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/progress"
	"github.com/fusheng/envkit/internal/ui"
)

// InstallResult 安装结果
type InstallResult struct {
	FailedComponents []string
	Succeeded        []string
}

// Error 实现 error，便于 CLI 非 0 退出
func (r InstallResult) Error() string {
	if len(r.FailedComponents) == 0 {
		return ""
	}
	return fmt.Sprintf("安装失败: %s", strings.Join(r.FailedComponents, ", "))
}

// RunInstallation 执行安装流程（使用当前 ui.Renderer 输出）
func RunInstallation(cfg *config.Config) InstallResult {
	var failed []string
	var succeeded []string
	total := CountInstallSteps(cfg)
	if total == 0 {
		total = 1
	}
	step := 0

	progress.Report(progress.Event{
		TaskID:  "install",
		Stage:   "preparing",
		Percent: 0,
		Message: "准备安装环境...",
	})

	sysInfo := detector.DetectSystem()
	ui.PrintSection("系统信息")
	ui.Info("操作系统: %s", string(sysInfo.OS))
	ui.Info("架构: %s", string(sysInfo.Architecture))
	if sysInfo.IsLinux() {
		ui.Info("发行版: %s", sysInfo.Distribution)
	}

	if len(cfg.Languages) > 0 {
		ui.PrintSection("安装编程语言")
		for _, lang := range cfg.Languages {
			step++
			langInstaller := installer.GetInstaller(lang.Name)
			if langInstaller == nil {
				ui.Warning("不支持的语言: %s", lang.Name)
				reportInstallError(lang.Name, fmt.Sprintf("不支持的语言: %s", lang.Name))
				failed = append(failed, lang.Name+"(不支持)")
				continue
			}
			if langInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", lang.Name, strings.TrimSpace(langInstaller.GetVersion()))
				reportInstallDone(lang.Name, fmt.Sprintf("%s 已安装", lang.Name), step, total)
				succeeded = append(succeeded, lang.Name)
			} else {
				reportInstallStart(lang.Name, fmt.Sprintf("正在安装 %s %s...", lang.Name, lang.Version), step, total)
				ui.Info("正在安装 %s %s...", lang.Name, lang.Version)
				if err := langInstaller.Install(lang.Version); err != nil {
					ui.Error("安装 %s 失败: %v", lang.Name, err)
					reportInstallError(lang.Name, fmt.Sprintf("安装 %s 失败: %v", lang.Name, err))
					failed = append(failed, lang.Name)
					continue
				}
				// 强制校验：脚本返回 0 但二进制不在 PATH 时不算成功
				if !langInstaller.IsInstalled() {
					msg := fmt.Sprintf("%s 安装脚本完成，但当前环境检测不到可执行文件", lang.Name)
					ui.Error("%s", msg)
					reportInstallError(lang.Name, msg)
					failed = append(failed, lang.Name)
					continue
				}
				ui.Success("%s 安装成功: %s", lang.Name, strings.TrimSpace(langInstaller.GetVersion()))
				reportInstallDone(lang.Name, fmt.Sprintf("%s 安装成功", lang.Name), step, total)
				succeeded = append(succeeded, lang.Name)
			}
			if lang.Mirror != "" {
				ui.Info("配置 %s 镜像源: %s", lang.Name, lang.Mirror)
				if err := configureMirror(lang.Name, lang.Mirror); err != nil {
					ui.Warning("配置镜像源失败: %v", err)
				} else {
					ui.Success("镜像源配置成功")
				}
			}
		}
	}

	if len(cfg.Tools) > 0 {
		ui.PrintSection("安装开发工具")
		for _, tool := range cfg.Tools {
			step++
			toolInstaller := installer.GetToolInstaller(tool)
			if toolInstaller == nil {
				ui.Warning("不支持的工具: %s", tool)
				reportInstallError(tool, fmt.Sprintf("不支持的工具: %s", tool))
				failed = append(failed, tool+"(不支持)")
				continue
			}
			if toolInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", tool, strings.TrimSpace(toolInstaller.GetVersion()))
				reportInstallDone(tool, fmt.Sprintf("%s 已安装", tool), step, total)
				succeeded = append(succeeded, tool)
			} else {
				reportInstallStart(tool, fmt.Sprintf("正在安装 %s...", tool), step, total)
				ui.Info("正在安装 %s...", tool)
				if err := toolInstaller.Install(); err != nil {
					ui.Error("安装 %s 失败: %v", tool, err)
					reportInstallError(tool, fmt.Sprintf("安装 %s 失败: %v", tool, err))
					failed = append(failed, tool)
					continue
				}
				if !toolInstaller.IsInstalled() {
					msg := fmt.Sprintf("%s 安装流程完成，但当前环境检测不到可执行文件", tool)
					ui.Error("%s", msg)
					reportInstallError(tool, msg)
					failed = append(failed, tool)
					continue
				}
				ui.Success("%s 安装成功！", tool)
				reportInstallDone(tool, fmt.Sprintf("%s 安装成功", tool), step, total)
				succeeded = append(succeeded, tool)
			}
		}
	}

	if len(cfg.Databases) > 0 {
		ui.PrintSection("启动数据库容器")
		dockerMgr := docker.NewContainerManager()
		if !dockerMgr.IsDockerRunning() {
			ui.Warning("Docker 未运行，跳过数据库容器启动")
			for _, db := range cfg.Databases {
				if db.Docker {
					failed = append(failed, db.Name+"(Docker未运行)")
					reportInstallError(db.Name, "Docker 未运行")
				}
			}
		} else {
			for _, db := range cfg.Databases {
				if !db.Docker {
					continue
				}
				step++
				reportInstallStart(db.Name, fmt.Sprintf("正在启动 %s %s...", db.Name, db.Version), step, total)
				ui.Info("正在启动 %s %s...", db.Name, db.Version)
				var err error
				switch db.Name {
				case "postgresql", "postgres":
					err = dockerMgr.StartPostgreSQL(db.Version, "postgres")
				case "redis":
					err = dockerMgr.StartRedis(db.Version)
				case "mysql":
					err = dockerMgr.StartMySQL(db.Version, "mysql")
				case "mongodb", "mongo":
					err = dockerMgr.StartMongoDB(db.Version)
				default:
					ui.Warning("不支持的数据库: %s", db.Name)
					failed = append(failed, db.Name+"(不支持)")
					continue
				}
				if err != nil {
					ui.Error("启动失败: %v", err)
					reportInstallError(db.Name, fmt.Sprintf("启动 %s 失败: %v", db.Name, err))
					failed = append(failed, db.Name)
				} else {
					reportInstallDone(db.Name, fmt.Sprintf("%s 已启动", db.Name), step, total)
					succeeded = append(succeeded, db.Name)
				}
			}
		}
	}

	ui.PrintSection("安装完成")
	if len(failed) > 0 {
		ui.Warning("部分组件安装或启动失败: %s", strings.Join(failed, ", "))
		if len(succeeded) > 0 {
			ui.Info("已成功: %s", strings.Join(succeeded, ", "))
		}
	} else {
		ui.Success("开发环境配置完成！")
	}
	ui.Info("提示: 若新开命令找不到，请执行 source ~/.bashrc 或重开终端（PATH 已写入配置）")

	finalStage := "done"
	finalMsg := "安装流程结束"
	if len(failed) > 0 {
		finalStage = "error"
		finalMsg = "安装完成（含失败项）: " + strings.Join(failed, ", ")
	}
	progress.Report(progress.Event{
		TaskID:  "install",
		Stage:   finalStage,
		Percent: 100,
		Message: finalMsg,
	})

	return InstallResult{FailedComponents: failed, Succeeded: succeeded}
}

func configureMirror(language, mirrorName string) error {
	// 安装配置里 language 名 → mirror 子命令名
	key := language
	switch language {
	case "node", "nodejs":
		key = "npm"
	case "python", "python3":
		key = "pip"
	case "golang":
		key = "go"
	}
	return ConfigureMirror(key, mirrorName)
}
