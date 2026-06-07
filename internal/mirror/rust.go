package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// RustConfigurator Rust镜像源配置器
type RustConfigurator struct {
	registry *Registry
}

// NewRustConfigurator 创建Rust配置器
func NewRustConfigurator(registry *Registry) *RustConfigurator {
	return &RustConfigurator{registry: registry}
}

// Configure 配置Rust镜像源
func (r *RustConfigurator) Configure(mirror string) error {
	if mirror == "" {
		mirror = "ustc"
	}

	url, ok := r.registry.GetMirror("rust", mirror)
	if !ok {
		return fmt.Errorf("未找到镜像源: %s", mirror)
	}

	// 获取cargo配置文件路径
	configPath, err := r.getConfigPath()
	if err != nil {
		return err
	}

	// 创建配置目录
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 根据不同的镜像源类型生成配置
	var content string
	if mirror == "bytedance" {
		// 字节跳动镜像使用不同的配置方式
		content = `[source.crates-io]
replace-with = 'rsproxy'

[source.rsproxy]
registry = "https://rsproxy.cn/crates.io-index"

[registries.rsproxy]
index = "https://rsproxy.cn/crates.io-index"

[net]
git-fetch-with-cli = true
`
	} else {
		// 其他镜像源使用标准配置
		content = fmt.Sprintf(`[source.crates-io]
replace-with = 'mirror'

[source.mirror]
registry = "%s"

[net]
git-fetch-with-cli = true
`, url)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✓ 已配置Rust镜像源: %s\n", url)
	return nil
}

// getConfigPath 获取cargo配置文件路径
func (r *RustConfigurator) getConfigPath() (string, error) {
	var configPath string

	switch runtime.GOOS {
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			return "", fmt.Errorf("无法获取USERPROFILE环境变量")
		}
		configPath = filepath.Join(userProfile, ".cargo", "config")
	case "darwin", "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法获取用户目录: %w", err)
		}
		configPath = filepath.Join(home, ".cargo", "config")
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return configPath, nil
}

// ConfigureRustup 配置rustup镜像源
func (r *RustConfigurator) ConfigureRustup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var profilePath string
	switch runtime.GOOS {
	case "darwin", "linux":
		// 配置rustup使用中科大镜像
		envVars := []string{
			"export RUSTUP_DIST_SERVER=https://mirrors.ustc.edu.cn/rust-static",
			"export RUSTUP_UPDATE_ROOT=https://mirrors.ustc.edu.cn/rust-static/rustup",
		}

		// 写入.bashrc
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			r.appendToFile(bashrc, envVars)
		}

		// 写入.zshrc
		zshrc := filepath.Join(home, ".zshrc")
		if _, err := os.Stat(zshrc); err == nil {
			r.appendToFile(zshrc, envVars)
		}

		profilePath = bashrc
	case "windows":
		// Windows使用系统环境变量
		os.Setenv("RUSTUP_DIST_SERVER", "https://mirrors.ustc.edu.cn/rust-static")
		os.Setenv("RUSTUP_UPDATE_ROOT", "https://mirrors.ustc.edu.cn/rust-static/rustup")
		return nil
	}

	fmt.Printf("✓ 已配置rustup镜像源，写入: %s\n", profilePath)
	return nil
}

// appendToFile 追加内容到文件
func (r *RustConfigurator) appendToFile(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("\n# Rust镜像源配置\n")
	for _, line := range lines {
		f.WriteString(line + "\n")
	}
	return nil
}

// Verify 验证配置
func (r *RustConfigurator) Verify() error {
	configPath, err := r.getConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在")
	}

	fmt.Printf("Cargo配置文件: %s\n", configPath)
	return nil
}
