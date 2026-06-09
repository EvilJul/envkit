package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fusheng/envkit/internal/ui"
)

// ManifestItem 记录单项安装信息
type ManifestItem struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "language" or "tool"
	Version     string    `json:"version"`
	Paths       []string  `json:"paths"`         // 创建的文件夹/文件路径
	ShellLines  []string  `json:"shell_lines"`   // 写入 shell 配置文件中的关键字特征
	InstalledAt time.Time `json:"installed_at"`
}

// Manifest 完整的安装清单
type Manifest struct {
	Items map[string]ManifestItem `json:"items"`
}

// getManifestPath 获取清单 JSON 文件路径 (~/.envkit/manifest.json)
func getManifestPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".envkit")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "manifest.json"), nil
}

// LoadManifest 加载现有清单
func LoadManifest() (*Manifest, error) {
	path, err := getManifestPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Manifest{Items: make(map[string]ManifestItem)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return &Manifest{Items: make(map[string]ManifestItem)}, nil
	}
	if m.Items == nil {
		m.Items = make(map[string]ManifestItem)
	}
	return &m, nil
}

// SaveManifest 保存清单
func SaveManifest(m *Manifest) error {
	path, err := getManifestPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RecordInstallation 记录一次成功安装
func RecordInstallation(name string, itemType string, version string, paths []string, shellLines []string) {
	m, err := LoadManifest()
	if err != nil {
		return
	}

	home, _ := os.UserHomeDir()
	var resolvedPaths []string
	for _, p := range paths {
		if strings.HasPrefix(p, "~") {
			p = filepath.Join(home, p[1:])
		}
		resolvedPaths = append(resolvedPaths, filepath.Clean(p))
	}

	m.Items[name] = ManifestItem{
		Name:        name,
		Type:        itemType,
		Version:     version,
		Paths:       resolvedPaths,
		ShellLines:  shellLines,
		InstalledAt: time.Now(),
	}

	_ = SaveManifest(m)
}

// CleanShellConfigs 清理 shell 配置文件中写入的环境变量行
func CleanShellConfigs(keywords []string) error {
	if len(keywords) == 0 {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	files := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		contentBytes, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(contentBytes), "\n")
		var newLines []string

		for i := 0; i < len(lines); i++ {
			line := lines[i]

			// 特殊处理 conda initialize 块
			if strings.Contains(line, "# >>> conda initialize >>>") {
				// 寻找闭合行并跳过整个块
				for i+1 < len(lines) {
					i++
					if strings.Contains(lines[i], "# <<< conda initialize <<<") {
						break
					}
				}
				continue
			}

			shouldSkip := false

			// 检查是否命中任何关键字
			for _, kw := range keywords {
				if strings.Contains(line, kw) {
					shouldSkip = true
					break
				}
			}

			if shouldSkip {
				// 如果是注释行如 "# added by envkit" 或 "# fnm integration"，我们通常还会额外跳过它后面的相关指令行和空行
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					for i+1 < len(lines) {
						trimmedNext := strings.TrimSpace(lines[i+1])
						if strings.HasPrefix(trimmedNext, "export ") || strings.HasPrefix(trimmedNext, "eval ") || trimmedNext == "" {
							i++
						} else {
							break
						}
					}
				}
				continue
			}

			newLines = append(newLines, line)
		}

		// 写回文件并去除多余的空行
		newContent := strings.Join(newLines, "\n")
		_ = os.WriteFile(file, []byte(newContent), 0644)
	}

	return nil
}

// UninstallComponent 卸载单个组件
func UninstallComponent(name string) error {
	m, err := LoadManifest()
	if err != nil {
		return fmt.Errorf("加载清单失败: %w", err)
	}

	item, exists := m.Items[name]
	if !exists {
		return fmt.Errorf("清单中未记录组件 '%s' 的安装信息，无法通过 EnvKit 自动卸载", name)
	}

	ui.Info("正在卸载组件: %s...", name)

	// 1. 删除安装目录和相关文件
	for _, path := range item.Paths {
		if path == "" || path == "/" {
			continue // 绝对安全保证，禁止删除根目录
		}
		if _, err := os.Stat(path); err == nil {
			ui.Info("  清理路径: %s", path)
			_ = os.RemoveAll(path)
		}
	}

	// 2. 清理环境变量配置
	if len(item.ShellLines) > 0 {
		ui.Info("  清理 Shell 环境变量配置...")
		_ = CleanShellConfigs(item.ShellLines)
	}

	// 3. 从清单中移除并保存
	delete(m.Items, name)
	_ = SaveManifest(m)

	ui.Success("组件 '%s' 已彻底卸载完成！", name)
	return nil
}
