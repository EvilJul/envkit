package service

import (
	"fmt"

	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/detector"
	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/mirror"
	"github.com/fusheng/envkit/internal/progress"
	"github.com/fusheng/envkit/internal/ui"
)

// InstallResult 安装结果
type InstallResult struct {
	FailedComponents []string
}

// RunInstallation 执行安装流程（使用当前 ui.Renderer 输出）
func RunInstallation(cfg *config.Config) InstallResult {
	var failed []string
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
				continue
			}
			if langInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", lang.Name, langInstaller.GetVersion())
				reportInstallDone(lang.Name, fmt.Sprintf("%s 已安装", lang.Name), step, total)
			} else {
				reportInstallStart(lang.Name, fmt.Sprintf("正在安装 %s %s...", lang.Name, lang.Version), step, total)
				ui.Info("正在安装 %s %s...", lang.Name, lang.Version)
				if err := langInstaller.Install(lang.Version); err != nil {
					ui.Error("安装 %s 失败: %v", lang.Name, err)
					reportInstallError(lang.Name, fmt.Sprintf("安装 %s 失败: %v", lang.Name, err))
					failed = append(failed, lang.Name)
					continue
				}
				ui.Success("%s 安装成功！", lang.Name)
				reportInstallDone(lang.Name, fmt.Sprintf("%s 安装成功", lang.Name), step, total)
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
				continue
			}
			if toolInstaller.IsInstalled() {
				ui.Info("%s 已安装: %s", tool, toolInstaller.GetVersion())
				reportInstallDone(tool, fmt.Sprintf("%s 已安装", tool), step, total)
			} else {
				reportInstallStart(tool, fmt.Sprintf("正在安装 %s...", tool), step, total)
				ui.Info("正在安装 %s...", tool)
				if err := toolInstaller.Install(); err != nil {
					ui.Error("安装 %s 失败: %v", tool, err)
					reportInstallError(tool, fmt.Sprintf("安装 %s 失败: %v", tool, err))
					failed = append(failed, tool)
					continue
				}
				ui.Success("%s 安装成功！", tool)
				reportInstallDone(tool, fmt.Sprintf("%s 安装成功", tool), step, total)
			}
		}
	}

	if len(cfg.Databases) > 0 {
		ui.PrintSection("启动数据库容器")
		dockerMgr := docker.NewContainerManager()
		if !dockerMgr.IsDockerRunning() {
			ui.Warning("Docker 未运行，跳过数据库容器启动")
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
					continue
				}
				if err != nil {
					ui.Error("启动失败: %v", err)
					reportInstallError(db.Name, fmt.Sprintf("启动 %s 失败: %v", db.Name, err))
					failed = append(failed, db.Name)
				} else {
					reportInstallDone(db.Name, fmt.Sprintf("%s 已启动", db.Name), step, total)
				}
			}
		}
	}

	ui.PrintSection("安装完成")
	if len(failed) > 0 {
		ui.Warning("部分组件安装或启动失败")
	} else {
		ui.Success("开发环境配置完成！")
	}

	progress.Report(progress.Event{
		TaskID:  "install",
		Stage:   "done",
		Percent: 100,
		Message: "安装流程结束",
	})

	return InstallResult{FailedComponents: failed}
}

func configureMirror(language, mirrorName string) error {
	registry := mirror.NewRegistry()
	switch language {
	case "node", "nodejs":
		return mirror.NewNPMConfigurator(registry).Configure(mirrorName)
	case "python":
		return mirror.NewPipConfigurator(registry).Configure(mirrorName)
	case "go", "golang":
		return mirror.NewGoConfigurator(registry).Configure(mirrorName)
	case "rust":
		return mirror.NewRustConfigurator(registry).Configure(mirrorName)
	default:
		return nil
	}
}
