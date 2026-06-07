package mirror

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// GoConfigurator Go镜像源配置器
type GoConfigurator struct {
	registry *Registry
}

// NewGoConfigurator 创建Go配置器
func NewGoConfigurator(registry *Registry) *GoConfigurator {
	return &GoConfigurator{registry: registry}
}

// Configure 配置Go镜像源
func (g *GoConfigurator) Configure(mirror string) error {
	if mirror == "" {
		mirror = "goproxy"
	}

	url, ok := g.registry.GetMirror("go", mirror)
	if !ok {
		return fmt.Errorf("未找到镜像源: %s", mirror)
	}

	// 设置GOPROXY环境变量
	if err := g.setGoProxy(url); err != nil {
		return err
	}

	// 设置GOSUMDB（用于校验）
	if err := g.setGoSumDB(); err != nil {
		return err
	}

	fmt.Printf("✓ 已配置Go镜像源: %s\n", url)
	return nil
}

// setGoProxy 设置GOPROXY
func (g *GoConfigurator) setGoProxy(url string) error {
	// 使用go env命令设置
	cmd := exec.Command("go", "env", "-w", fmt.Sprintf("GOPROXY=%s,direct", url))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置GOPROXY失败: %w", err)
	}
	return nil
}

// setGoSumDB 设置GOSUMDB
func (g *GoConfigurator) setGoSumDB() error {
	// 国内环境建议使用goproxy.cn的sum.golang.org代理
	cmd := exec.Command("go", "env", "-w", "GOSUMDB=sum.golang.org")
	if err := cmd.Run(); err != nil {
		// 如果设置失败，可以关闭校验（不推荐，但某些情况下需要）
		cmd = exec.Command("go", "env", "-w", "GOSUMDB=off")
		_ = cmd.Run()
	}
	return nil
}

// ConfigurePrivate 配置私有仓库
func (g *GoConfigurator) ConfigurePrivate(pattern string) error {
	cmd := exec.Command("go", "env", "-w", fmt.Sprintf("GOPRIVATE=%s", pattern))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置GOPRIVATE失败: %w", err)
	}
	fmt.Printf("✓ 已配置Go私有仓库: %s\n", pattern)
	return nil
}

// SetToProfile 将Go环境变量写入shell配置文件
func (g *GoConfigurator) SetToProfile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var profilePath string
	switch runtime.GOOS {
	case "darwin", "linux":
		// 尝试写入.bashrc和.zshrc
		bashrc := home + "/.bashrc"
		zshrc := home + "/.zshrc"

		goproxy, _ := exec.Command("go", "env", "GOPROXY").Output()
		exportLine := fmt.Sprintf("\nexport GOPROXY=%s", string(goproxy))

		// 写入.bashrc
		if _, err := os.Stat(bashrc); err == nil {
			f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer f.Close()
				f.WriteString(exportLine)
			}
		}

		// 写入.zshrc
		if _, err := os.Stat(zshrc); err == nil {
			f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer f.Close()
				f.WriteString(exportLine)
			}
		}

		profilePath = bashrc
	case "windows":
		// Windows使用系统环境变量，已通过go env -w设置
		return nil
	}

	fmt.Printf("✓ 已将Go环境变量写入: %s\n", profilePath)
	return nil
}

// Verify 验证配置
func (g *GoConfigurator) Verify() error {
	cmd := exec.Command("go", "env", "GOPROXY")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Printf("当前GOPROXY: %s", string(output))
	return nil
}
