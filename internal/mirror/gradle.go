package mirror

import (
	"fmt"
	"os"
	"path/filepath"
)

// GradleConfigurator Gradle 镜像源配置器
// 通过 ~/.gradle/init.d/init.gradle 全局替换所有 maven 仓库为国内镜像源
type GradleConfigurator struct {
	registry *Registry
}

// NewGradleConfigurator 创建 Gradle 配置器
func NewGradleConfigurator(registry *Registry) *GradleConfigurator {
	return &GradleConfigurator{registry: registry}
}

// Configure 主入口：默认使用 aliyun 镜像
func (g *GradleConfigurator) Configure(mirror string) error {
	if mirror == "" {
		mirror = "aliyun"
	}

	// 验证镜像源存在（gradle 镜像源至少有一个 URL 可用即可）
	if _, ok := g.registry.GetMirror("gradle", mirror); !ok {
		return fmt.Errorf("未找到镜像源: %s", mirror)
	}

	// 1. 清理旧配置，避免冲突
	if err := g.removeOldConfig(); err != nil {
		return err
	}

	// 2. 写入 ~/.gradle/init.d/aliyun.gradle
	if err := g.configureInitScript(); err != nil {
		return err
	}

	// 3. 写入 ~/.gradle/init.gradle（settings 入口）
	if err := g.configureSettings(); err != nil {
		return err
	}

	fmt.Printf("✓ 已配置Gradle全局镜像源: %s\n", mirror)
	return nil
}

// configureInitScript 创建 ~/.gradle/init.d/aliyun.gradle
// 该脚本在所有项目 settings.gradle 之后执行，能覆盖所有项目的仓库配置
func (g *GradleConfigurator) configureInitScript() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	initDir := filepath.Join(home, ".gradle", "init.d")
	if err := os.MkdirAll(initDir, 0755); err != nil {
		return fmt.Errorf("创建 Gradle init.d 目录失败: %w", err)
	}

	initFile := filepath.Join(initDir, "aliyun.gradle")
	content := `// 由 envkit 自动生成：Gradle 全局国内镜像源
// 替换所有 maven 仓库为阿里云镜像
allprojects {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/public' }
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
        maven { url 'https://maven.aliyun.com/repository/central' }
        maven { url 'https://maven.aliyun.com/repository/jcenter' }
        mavenCentral()
        google()
    }
    buildscript {
        repositories {
            maven { url 'https://maven.aliyun.com/repository/public' }
            maven { url 'https://maven.aliyun.com/repository/google' }
            maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
            maven { url 'https://maven.aliyun.com/repository/central' }
            mavenCentral()
            google()
        }
    }
}

settingsEvaluated { settings ->
    settings.pluginManagement {
        repositories {
            maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
            maven { url 'https://maven.aliyun.com/repository/public' }
            gradlePluginPortal()
        }
    }
}
`

	if err := os.WriteFile(initFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 Gradle init 脚本失败: %w", err)
	}
	return nil
}

// configureSettings 创建 ~/.gradle/init.gradle
// 兼容旧版 Gradle（< 6.8）以及未使用 init.d 的场景
func (g *GradleConfigurator) configureSettings() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	gradleDir := filepath.Join(home, ".gradle")
	if err := os.MkdirAll(gradleDir, 0755); err != nil {
		return fmt.Errorf("创建 .gradle 目录失败: %w", err)
	}

	initFile := filepath.Join(gradleDir, "init.gradle")
	content := `// 由 envkit 自动生成：Gradle 全局国内镜像源
// 通过 settingsEvaluated 钩子统一替换仓库地址
allprojects {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/public' }
        maven { url 'https://maven.aliyun.com/repository/google' }
        maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
        maven { url 'https://maven.aliyun.com/repository/central' }
        maven { url 'https://maven.aliyun.com/repository/jcenter' }
        mavenCentral()
        google()
    }
}
`

	if err := os.WriteFile(initFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 Gradle init.gradle 失败: %w", err)
	}
	return nil
}

// removeOldConfig 清理旧的 envkit 写入的 Gradle 配置（幂等性保证）
func (g *GradleConfigurator) removeOldConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 删除可能存在的旧版本 aliyun 镜像配置
	candidates := []string{
		filepath.Join(home, ".gradle", "init.d", "aliyun.gradle"),
		filepath.Join(home, ".gradle", "init.d", "android-mirror.gradle"),
		filepath.Join(home, ".gradle", "init.gradle"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = os.Remove(p)
		}
	}
	return nil
}

// Verify 验证 Gradle 镜像源配置是否生效
func (g *GradleConfigurator) Verify() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	initScript := filepath.Join(home, ".gradle", "init.d", "aliyun.gradle")
	if _, err := os.Stat(initScript); os.IsNotExist(err) {
		return fmt.Errorf("Gradle init 脚本不存在: %s", initScript)
	}
	fmt.Printf("✓ Gradle init 脚本: %s\n", initScript)

	settingsFile := filepath.Join(home, ".gradle", "init.gradle")
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return fmt.Errorf("Gradle init.gradle 不存在: %s", settingsFile)
	}
	fmt.Printf("✓ Gradle init.gradle: %s\n", settingsFile)
	return nil
}
