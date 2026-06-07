package mirror

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// PipConfigurator pip镜像源配置器
type PipConfigurator struct {
	registry *Registry
}

// NewPipConfigurator 创建pip配置器
func NewPipConfigurator(registry *Registry) *PipConfigurator {
	return &PipConfigurator{registry: registry}
}

// Configure 配置pip镜像源
func (p *PipConfigurator) Configure(mirror string) error {
	if mirror == "" {
		mirror = "tsinghua"
	}

	url, ok := p.registry.GetMirror("pip", mirror)
	if !ok {
		return fmt.Errorf("未找到镜像源: %s", mirror)
	}

	// 获取pip配置文件路径
	configPath, err := p.getConfigPath()
	if err != nil {
		return err
	}

	// 创建配置目录
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入配置文件
	content := fmt.Sprintf(`[global]
index-url = %s
[install]
trusted-host = %s
`, url, p.extractHost(url))

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✓ 已配置pip镜像源: %s\n", url)
	return nil
}

// getConfigPath 获取pip配置文件路径
func (p *PipConfigurator) getConfigPath() (string, error) {
	var configPath string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("无法获取APPDATA环境变量")
		}
		configPath = filepath.Join(appData, "pip", "pip.ini")
	case "darwin", "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法获取用户目录: %w", err)
		}
		configPath = filepath.Join(home, ".pip", "pip.conf")
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return configPath, nil
}

// extractHost 从URL中提取主机名
func (p *PipConfigurator) extractHost(url string) string {
	// 简单提取主机名（去掉协议和路径）
	host := url
	if len(url) > 8 && url[:8] == "https://" {
		host = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		host = url[7:]
	}

	// 去掉路径部分
	for i, c := range host {
		if c == '/' {
			return host[:i]
		}
	}
	return host
}

// Verify 验证配置
func (p *PipConfigurator) Verify() error {
	cmd := exec.Command("pip", "config", "get", "global.index-url")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Printf("当前pip镜像源: %s", string(output))
	return nil
}
