package appapi

import (
	"fmt"
	"strings"

	"github.com/fusheng/envkit/internal/docker"
)

// Database GUI 数据库容器信息
type Database struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"` // running | stopped
}

// 容器名 → 展示类型
var envkitDBTypes = map[string]string{
	"envkit-postgres": "PostgreSQL",
	"envkit-redis":    "Redis",
	"envkit-mysql":    "MySQL",
	"envkit-mongodb":  "MongoDB",
}

// GetDatabases 列出 envkit 管理的数据库容器
func GetDatabases() []Database {
	mgr := docker.NewContainerManager()
	if !mgr.IsDockerRunning() {
		return []Database{}
	}
	rows, err := mgr.ListContainersData()
	if err != nil {
		return []Database{}
	}
	out := make([]Database, 0, len(rows))
	for _, r := range rows {
		dbType := envkitDBTypes[r.Name]
		if dbType == "" {
			// 非标准命名：去掉 envkit- 前缀作为类型
			dbType = strings.TrimPrefix(r.Name, "envkit-")
			if dbType == r.Name {
				dbType = r.Name
			}
		}
		status := "stopped"
		// docker 状态如 "Up 2 hours" / "Exited (0) 3 days ago"
		if strings.HasPrefix(strings.ToLower(r.Status), "up") {
			status = "running"
		}
		out = append(out, Database{
			Name:    r.Name,
			Type:    dbType,
			Version: "",
			Status:  status,
		})
	}
	return out
}

// SupportedDatabase 是否支持启动的数据库名
func SupportedDatabase(name string) bool {
	switch strings.ToLower(name) {
	case "postgres", "postgresql", "redis", "mysql", "mongodb", "mongo":
		return true
	default:
		return false
	}
}

// ValidateDatabaseName 校验数据库名，不支持时返回错误
func ValidateDatabaseName(name string) error {
	if SupportedDatabase(name) {
		return nil
	}
	return fmt.Errorf("不支持的数据库: %s（支持: postgres, redis, mysql, mongodb）", name)
}
