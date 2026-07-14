package service

import (
	"fmt"

	"github.com/fusheng/envkit/internal/docker"
)

// DockerStart 启动数据库容器
func DockerStart(database, version string) error {
	dockerMgr := docker.NewContainerManager()
	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行，请先启动 Docker")
	}
	switch database {
	case "postgresql", "postgres":
		return dockerMgr.StartPostgreSQL(version, "postgres")
	case "redis":
		return dockerMgr.StartRedis(version)
	case "mysql":
		return dockerMgr.StartMySQL(version, "mysql")
	case "mongodb", "mongo":
		return dockerMgr.StartMongoDB(version)
	default:
		return fmt.Errorf("不支持的数据库: %s", database)
	}
}

// DockerStop 停止容器
func DockerStop(containerName string) error {
	return docker.NewContainerManager().StopContainer(containerName)
}

// DockerList 列出容器（打印）
func DockerList() error {
	return docker.NewContainerManager().ListContainers()
}

// DockerListData 列出容器（结构化，供 TUI）
func DockerListData() ([]docker.ContainerInfo, error) {
	return docker.NewContainerManager().ListContainersData()
}

// DockerRemove 删除容器
func DockerRemove(containerName string, removeVolume bool) error {
	return docker.NewContainerManager().RemoveContainer(containerName, removeVolume)
}

// DockerDatabases 支持的数据库类型
func DockerDatabases() []string {
	return []string{"postgres", "redis", "mysql", "mongodb"}
}
