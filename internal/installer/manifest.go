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
	Paths       []string  `json:"paths"`
	ShellLines  []string  `json:"shell_lines"`
	InstalledAt time.Time `json:"installed_at"`
}

// Manifest 完整的安装清单
type Manifest struct {
	Items map[string]ManifestItem `json:"items"`
}

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

// LoadManifest 加载现有清单；损坏的 JSON 返回错误（避免静默清空）
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
	if len(strings.TrimSpace(string(data))) == 0 {
		return &Manifest{Items: make(map[string]ManifestItem)}, nil
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("解析清单失败 (%s): %w", path, err)
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

// RecordInstallation 记录一次成功安装；保存失败返回 error
func RecordInstallation(name string, itemType string, version string, paths []string, shellLines []string) error {
	name = NormalizeComponentName(name)

	m, err := LoadManifest()
	if err != nil {
		return err
	}

	home, _ := os.UserHomeDir()
	var resolvedPaths []string
	for _, p := range paths {
		p = filepath.FromSlash(p)
		if strings.HasPrefix(p, "~") {
			// 不能用 filepath.Join(home, p[1:])：Windows 上 "~/foo" 的 p[1:] 是 "/foo"，Join 会当成绝对路径丢弃 home
			rel := strings.TrimPrefix(p, "~")
			rel = strings.TrimPrefix(rel, "/")
			rel = strings.TrimPrefix(rel, `\`)
			p = filepath.Join(home, filepath.FromSlash(rel))
		}
		resolvedPaths = append(resolvedPaths, filepath.Clean(p))
	}

	// 默认带组件标签，便于精确清理 shell
	if len(shellLines) == 0 {
		shellLines = []string{"# envkit:" + name}
	}

	m.Items[name] = ManifestItem{
		Name:        name,
		Type:        itemType,
		Version:     version,
		Paths:       resolvedPaths,
		ShellLines:  shellLines,
		InstalledAt: time.Now(),
	}

	return SaveManifest(m)
}

// recordInstall 安装成功后写入清单；失败时返回包装错误，避免静默丢失导致无法卸载
func recordInstall(name, itemType, version string, paths, shellLines []string) error {
	if err := RecordInstallation(name, itemType, version, paths, shellLines); err != nil {
		return fmt.Errorf("%s 已安装，但写入清单失败（后续卸载可能不可用）: %w", name, err)
	}
	return nil
}

// CleanShellConfigs 清理 shell 配置文件中与 keywords 匹配的行
// stripConda: 仅卸载 miniconda 时为 true
func CleanShellConfigs(keywords []string, stripConda bool) error {
	if len(keywords) == 0 && !stripConda {
		return nil
	}

	files := getShellConfigFiles()
	if files == nil {
		return nil
	}

	var lastErr error
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		contentBytes, err := os.ReadFile(file)
		if err != nil {
			lastErr = err
			continue
		}

		lines := strings.Split(string(contentBytes), "\n")
		var newLines []string

		for i := 0; i < len(lines); i++ {
			line := lines[i]

			if stripConda && strings.Contains(line, "# >>> conda initialize >>>") {
				for i+1 < len(lines) {
					i++
					if strings.Contains(lines[i], "# <<< conda initialize <<<") {
						break
					}
				}
				continue
			}

			shouldSkip := false
			for _, kw := range keywords {
				if kw != "" && strings.Contains(line, kw) {
					shouldSkip = true
					break
				}
			}

			if shouldSkip {
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					for i+1 < len(lines) {
						trimmedNext := strings.TrimSpace(lines[i+1])
						if strings.HasPrefix(trimmedNext, "export ") || strings.HasPrefix(trimmedNext, "eval ") ||
							strings.HasPrefix(trimmedNext, "[[") || strings.HasPrefix(trimmedNext, "[ -s") ||
							trimmedNext == "" {
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

		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// NormalizeComponentName 统一组件别名到清单键（安装/卸载对齐）
func NormalizeComponentName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "code":
		return "vscode"
	case "nodejs":
		return "node"
	case "python3":
		return "python"
	case "golang":
		return "go"
	case "jdk":
		return "java"
	case "conda":
		return "miniconda"
	case "android-sdk":
		return "android"
	case "esp-idf":
		return "espidf"
	default:
		return name
	}
}

// UninstallComponent 卸载单个组件；任一步失败则返回 error
func UninstallComponent(name string) error {
	name = NormalizeComponentName(name)

	m, err := LoadManifest()
	if err != nil {
		return fmt.Errorf("加载清单失败: %w", err)
	}

	item, exists := m.Items[name]
	if !exists {
		// 兼容历史清单可能用 code 记录 VS Code
		if alt, ok := m.Items["code"]; ok && name == "vscode" {
			item = alt
			exists = true
			name = "code"
		}
	}
	if !exists {
		return fmt.Errorf("清单中未记录组件 '%s' 的安装信息，无法通过 EnvKit 自动卸载", name)
	}

	ui.Info("正在卸载组件: %s...", name)

	var errs []string
	for _, path := range item.Paths {
		if path == "" || path == "/" || path == filepath.Dir("/") {
			continue
		}

		info, err := os.Lstat(path)
		if err != nil {
			if !os.IsNotExist(err) {
				errs = append(errs, fmt.Sprintf("检查 %s: %v", path, err))
			}
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			ui.Info("  清理符号链接: %s", path)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				errs = append(errs, fmt.Sprintf("删除链接 %s: %v", path, err))
			}
		} else {
			ui.Info("  清理路径: %s", path)
			if err := os.RemoveAll(path); err != nil {
				errs = append(errs, fmt.Sprintf("删除 %s: %v", path, err))
			}
		}
	}

	if len(item.ShellLines) > 0 {
		ui.Info("  清理 Shell 环境变量配置...")
		stripConda := name == "miniconda" || name == "conda"
		if err := CleanShellConfigs(item.ShellLines, stripConda); err != nil {
			errs = append(errs, fmt.Sprintf("清理 shell 配置: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("卸载 %s 未完全成功: %s", name, strings.Join(errs, "; "))
	}

	delete(m.Items, name)
	if err := SaveManifest(m); err != nil {
		return fmt.Errorf("更新清单失败: %w", err)
	}

	ui.Success("组件 '%s' 已卸载完成", name)
	return nil
}
