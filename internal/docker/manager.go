package docker

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fusheng/envkit/internal/ui"
)

// ContainerManager Docker 容器管理器
type ContainerManager struct{}

// NewContainerManager 创建容器管理器
func NewContainerManager() *ContainerManager {
	return &ContainerManager{}
}

// IsDockerRunning 检查 Docker 是否运行
func (m *ContainerManager) IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// StartPostgreSQL 启动 PostgreSQL 容器
func (m *ContainerManager) StartPostgreSQL(version, password string) error {
	if !m.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	containerName := "envkit-postgres"

	spinner := ui.NewSpinner(fmt.Sprintf("正在启动 PostgreSQL %s", version))
	spinner.Start()
	defer spinner.Stop()

	// 检查容器是否已存在
	if m.containerExists(containerName) {
		ui.Info("PostgreSQL 容器已存在，正在启动...")
		return m.startContainer(containerName)
	}

	// 创建并启动新容器
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-e", "POSTGRES_PASSWORD="+password,
		"-p", "5432:5432",
		"-v", "envkit-postgres-data:/var/lib/postgresql/data",
		"postgres:"+version)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动失败: %s", string(output))
	}

	ui.Success("PostgreSQL 容器启动成功！")
	ui.Info("连接信息:")
	ui.Info("  主机: localhost")
	ui.Info("  端口: 5432")
	ui.Info("  用户: postgres")
	ui.Info("  密码: %s", password)
	ui.Info("  数据卷: envkit-postgres-data")

	return nil
}

// StartRedis 启动 Redis 容器
func (m *ContainerManager) StartRedis(version string) error {
	if !m.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	containerName := "envkit-redis"

	spinner := ui.NewSpinner(fmt.Sprintf("正在启动 Redis %s", version))
	spinner.Start()
	defer spinner.Stop()

	// 检查容器是否已存在
	if m.containerExists(containerName) {
		ui.Info("Redis 容器已存在，正在启动...")
		return m.startContainer(containerName)
	}

	// 创建并启动新容器
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-p", "6379:6379",
		"-v", "envkit-redis-data:/data",
		"redis:"+version,
		"redis-server", "--appendonly", "yes")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动失败: %s", string(output))
	}

	ui.Success("Redis 容器启动成功！")
	ui.Info("连接信息:")
	ui.Info("  主机: localhost")
	ui.Info("  端口: 6379")
	ui.Info("  数据卷: envkit-redis-data")

	return nil
}

// StartMySQL 启动 MySQL 容器
func (m *ContainerManager) StartMySQL(version, password string) error {
	if !m.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	containerName := "envkit-mysql"

	spinner := ui.NewSpinner(fmt.Sprintf("正在启动 MySQL %s", version))
	spinner.Start()
	defer spinner.Stop()

	// 检查容器是否已存在
	if m.containerExists(containerName) {
		ui.Info("MySQL 容器已存在，正在启动...")
		return m.startContainer(containerName)
	}

	// 创建并启动新容器
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-e", "MYSQL_ROOT_PASSWORD="+password,
		"-p", "3306:3306",
		"-v", "envkit-mysql-data:/var/lib/mysql",
		"mysql:"+version)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动失败: %s", string(output))
	}

	ui.Success("MySQL 容器启动成功！")
	ui.Info("连接信息:")
	ui.Info("  主机: localhost")
	ui.Info("  端口: 3306")
	ui.Info("  用户: root")
	ui.Info("  密码: %s", password)
	ui.Info("  数据卷: envkit-mysql-data")

	return nil
}

// StartMongoDB 启动 MongoDB 容器
func (m *ContainerManager) StartMongoDB(version string) error {
	if !m.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}

	containerName := "envkit-mongodb"

	spinner := ui.NewSpinner(fmt.Sprintf("正在启动 MongoDB %s", version))
	spinner.Start()
	defer spinner.Stop()

	// 检查容器是否已存在
	if m.containerExists(containerName) {
		ui.Info("MongoDB 容器已存在，正在启动...")
		return m.startContainer(containerName)
	}

	// 创建并启动新容器
	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-p", "27017:27017",
		"-v", "envkit-mongodb-data:/data/db",
		"mongo:"+version)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动失败: %s", string(output))
	}

	ui.Success("MongoDB 容器启动成功！")
	ui.Info("连接信息:")
	ui.Info("  主机: localhost")
	ui.Info("  端口: 27017")
	ui.Info("  数据卷: envkit-mongodb-data")

	return nil
}

// ContainerInfo 容器摘要（供 CLI 表格与 TUI 展示）
type ContainerInfo struct {
	Name   string
	Status string
	Ports  string
}

// ListContainersData 返回 envkit 管理的容器列表
func (m *ContainerManager) ListContainersData() ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "-a",
		"--filter", "name=envkit-",
		"--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}
	var rows []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		info := ContainerInfo{}
		if len(parts) > 0 {
			info.Name = parts[0]
		}
		if len(parts) > 1 {
			info.Status = parts[1]
		}
		if len(parts) > 2 {
			info.Ports = parts[2]
		}
		rows = append(rows, info)
	}
	return rows, nil
}

// ListContainers 列出所有 envkit 管理的容器（打印到 stdout，CLI 用）
func (m *ContainerManager) ListContainers() error {
	rows, err := m.ListContainersData()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("（没有 envkit-* 容器）")
		return nil
	}
	fmt.Printf("%-24s %-24s %s\n", "NAMES", "STATUS", "PORTS")
	for _, r := range rows {
		fmt.Printf("%-24s %-24s %s\n", r.Name, r.Status, r.Ports)
	}
	return nil
}

// StopContainer 停止容器
func (m *ContainerManager) StopContainer(name string) error {
	spinner := ui.NewSpinner(fmt.Sprintf("正在停止容器 %s", name))
	spinner.Start()
	defer spinner.Stop()

	cmd := exec.Command("docker", "stop", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}

	ui.Success("容器已停止: %s", name)
	return nil
}

// RemoveContainer 删除容器
func (m *ContainerManager) RemoveContainer(name string, removeVolume bool) error {
	// 先停止容器
	stopCmd := exec.Command("docker", "stop", name)
	stopCmd.Run() // 忽略错误，容器可能已经停止

	spinner := ui.NewSpinner(fmt.Sprintf("正在删除容器 %s", name))
	spinner.Start()
	defer spinner.Stop()

	// 删除容器
	cmd := exec.Command("docker", "rm", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("删除容器失败: %w", err)
	}

	// 删除数据卷
	if removeVolume {
		volumeName := name + "-data"
		volumeCmd := exec.Command("docker", "volume", "rm", volumeName)
		if err := volumeCmd.Run(); err != nil {
			ui.Warning("删除数据卷失败: %v", err)
		}
	}

	ui.Success("容器已删除: %s", name)
	return nil
}

// 辅助方法

func (m *ContainerManager) containerExists(name string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name="+name, "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == name
}

func (m *ContainerManager) startContainer(name string) error {
	cmd := exec.Command("docker", "start", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动容器失败: %w", err)
	}
	ui.Success("容器启动成功: %s", name)
	return nil
}
