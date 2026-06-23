package mirror

import (
	"fmt"
	"os/exec"
)

// NPMConfigurator npm镜像源配置器
type NPMConfigurator struct {
	registry *Registry
}

// NewNPMConfigurator 创建npm配置器
func NewNPMConfigurator(registry *Registry) *NPMConfigurator {
	return &NPMConfigurator{registry: registry}
}

// Configure 配置npm镜像源
func (n *NPMConfigurator) Configure(mirror string) error {
	if mirror == "" {
		mirror = "npmmirror"
	}

	url, ok := n.registry.GetMirror("npm", mirror)
	if !ok {
		return fmt.Errorf("未找到镜像源: %s", mirror)
	}

	// 设置npm镜像源
	cmd := exec.Command("npm", "config", "set", "registry", url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置npm镜像源失败: %w", err)
	}

	// 设置pnpm镜像源
	if exec.Command("pnpm", "--version").Run() == nil {
		cmd = exec.Command("pnpm", "config", "set", "registry", url)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  设置 pnpm 镜像源失败: %v\n", err)
		}
	}

	// 设置yarn镜像源
	if exec.Command("yarn", "--version").Run() == nil {
		cmd = exec.Command("yarn", "config", "set", "registry", url)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  设置 yarn 镜像源失败: %v\n", err)
		}
	}

	fmt.Printf("✓ 已配置npm镜像源: %s\n", url)
	return nil
}

// Verify 验证配置
func (n *NPMConfigurator) Verify() error {
	cmd := exec.Command("npm", "config", "get", "registry")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Printf("当前npm镜像源: %s", string(output))
	return nil
}
