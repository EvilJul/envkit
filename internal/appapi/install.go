package appapi

import (
	"fmt"

	"github.com/fusheng/envkit/internal/docker"
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/progress"
)

// InstallLanguageWithProgress 安装语言环境并上报进度
func InstallLanguageWithProgress(name, version string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{
		TaskID: name, Stage: "preparing", Percent: 0, Message: "准备安装 " + name + "...",
	})

	langInstaller := installer.GetInstaller(name)
	if langInstaller == nil {
		err := fmt.Errorf("不支持的语言: %s", name)
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}

	progress.Report(progress.Event{
		TaskID: name, Stage: "installing", Percent: 30, Message: "正在安装 " + name + " " + version + "...",
	})

	if err := langInstaller.Install(version); err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}

	progress.Report(progress.Event{
		TaskID: name, Stage: "done", Percent: 100, Message: name + " 安装完成",
	})
	return nil
}

// UninstallLanguageWithProgress 卸载语言环境
func UninstallLanguageWithProgress(name string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: name, Stage: "stopping", Percent: 10, Message: "准备卸载 " + name + "..."})
	if err := installer.UninstallComponent(name); err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: name, Stage: "done", Percent: 100, Message: name + " 卸载完成"})
	return nil
}

// InstallToolWithProgress 安装工具
func InstallToolWithProgress(name string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: name, Stage: "preparing", Percent: 0, Message: "准备安装 " + name + "..."})

	toolInstaller := installer.GetToolInstaller(name)
	if toolInstaller == nil {
		err := fmt.Errorf("不支持的工具: %s", name)
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}

	progress.Report(progress.Event{TaskID: name, Stage: "installing", Percent: 40, Message: "正在安装 " + name + "..."})
	if err := toolInstaller.Install(); err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: name, Stage: "done", Percent: 100, Message: name + " 安装完成"})
	return nil
}

// UninstallToolWithProgress 卸载工具
func UninstallToolWithProgress(name string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: name, Stage: "removing", Percent: 10, Message: "准备卸载 " + name + "..."})
	if err := installer.UninstallComponent(name); err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: name, Stage: "done", Percent: 100, Message: name + " 卸载完成"})
	return nil
}

// InstallAndroidWithProgress 安装 Android SDK
func InstallAndroidWithProgress(rep progress.Reporter) error {
	return InstallToolWithProgress("android", rep)
}

// UninstallAndroidWithProgress 卸载 Android SDK
func UninstallAndroidWithProgress(rep progress.Reporter) error {
	return UninstallToolWithProgress("android", rep)
}

// ConfigureAndroidMirrorWithProgress 配置 Android 镜像
func ConfigureAndroidMirrorWithProgress(mirrorName string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	taskID := "android-mirror"
	progress.Report(progress.Event{TaskID: taskID, Stage: "configuring", Percent: 20, Message: "配置 Android 镜像源..."})
	if err := ConfigureAndroidMirror(mirrorName); err != nil {
		progress.Report(progress.Event{TaskID: taskID, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: taskID, Stage: "done", Percent: 100, Message: "Android 镜像配置完成"})
	return nil
}

// ConfigureGradleMirrorWithProgress 配置 Gradle 镜像
func ConfigureGradleMirrorWithProgress(mirrorName string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	taskID := "gradle-mirror"
	progress.Report(progress.Event{TaskID: taskID, Stage: "configuring", Percent: 20, Message: "配置 Gradle 镜像源..."})
	if err := ConfigureGradleMirror(mirrorName); err != nil {
		progress.Report(progress.Event{TaskID: taskID, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: taskID, Stage: "done", Percent: 100, Message: "Gradle 镜像配置完成"})
	return nil
}

// StartDatabaseWithProgress 启动数据库容器
func StartDatabaseWithProgress(name, version string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: name, Stage: "preparing", Percent: 10, Message: "准备启动 " + name + "..."})

	dockerMgr := docker.NewContainerManager()
	if !dockerMgr.IsDockerRunning() {
		err := fmt.Errorf("Docker 未运行，请先启动 Docker")
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}

	if err := ValidateDatabaseName(name); err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	if version == "" {
		version = "latest"
	}

	progress.Report(progress.Event{TaskID: name, Stage: "starting", Percent: 50, Message: "正在启动容器..."})
	var err error
	switch name {
	case "postgres", "postgresql":
		err = dockerMgr.StartPostgreSQL(version, "postgres")
	case "redis":
		err = dockerMgr.StartRedis(version)
	case "mysql":
		err = dockerMgr.StartMySQL(version, "mysql")
	case "mongodb", "mongo":
		err = dockerMgr.StartMongoDB(version)
	default:
		err = fmt.Errorf("不支持的数据库: %s", name)
	}
	if err != nil {
		progress.Report(progress.Event{TaskID: name, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: name, Stage: "done", Percent: 100, Message: name + " 启动完成"})
	return nil
}

// StopDatabaseWithProgress 停止数据库容器
func StopDatabaseWithProgress(containerName string, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: containerName, Stage: "stopping", Percent: 30, Message: "正在停止容器..."})
	if err := docker.NewContainerManager().StopContainer(containerName); err != nil {
		progress.Report(progress.Event{TaskID: containerName, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: containerName, Stage: "done", Percent: 100, Message: "容器已停止"})
	return nil
}

// RemoveDatabaseWithProgress 删除数据库容器
func RemoveDatabaseWithProgress(containerName string, removeVolume bool, rep progress.Reporter) error {
	old := progress.SetReporter(rep)
	defer progress.SetReporter(old)

	progress.Report(progress.Event{TaskID: containerName, Stage: "removing", Percent: 30, Message: "正在删除容器..."})
	if err := docker.NewContainerManager().RemoveContainer(containerName, removeVolume); err != nil {
		progress.Report(progress.Event{TaskID: containerName, Stage: "error", Percent: 0, Message: err.Error()})
		return err
	}
	progress.Report(progress.Event{TaskID: containerName, Stage: "done", Percent: 100, Message: "容器已删除"})
	return nil
}
