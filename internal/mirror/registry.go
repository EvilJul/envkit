package mirror

// Registry 国内镜像源注册表
type Registry struct {
	mirrors map[string]map[string]string
}

// NewRegistry 创建镜像源注册表
func NewRegistry() *Registry {
	r := &Registry{
		mirrors: make(map[string]map[string]string),
	}
	r.initMirrors()
	return r
}

// initMirrors 初始化所有镜像源
func (r *Registry) initMirrors() {
	// Node.js 镜像源
	r.mirrors["npm"] = map[string]string{
		"npmmirror": "https://registry.npmmirror.com",
		"taobao":    "https://registry.npmmirror.com",
		"tencent":   "https://mirrors.cloud.tencent.com/npm/",
		"huawei":    "https://repo.huaweicloud.com/repository/npm/",
	}

	// Python 镜像源
	r.mirrors["pip"] = map[string]string{
		"tsinghua": "https://pypi.tuna.tsinghua.edu.cn/simple",
		"aliyun":   "https://mirrors.aliyun.com/pypi/simple/",
		"ustc":     "https://mirrors.ustc.edu.cn/pypi/web/simple",
		"douban":   "https://pypi.douban.com/simple/",
	}

	// Go 镜像源
	r.mirrors["go"] = map[string]string{
		"goproxy":  "https://goproxy.cn",
		"aliyun":   "https://mirrors.aliyun.com/goproxy/",
		"qiniu":    "https://goproxy.qiniu.io",
		"official": "https://goproxy.io",
	}

	// Rust 镜像源
	r.mirrors["rust"] = map[string]string{
		"ustc":      "https://mirrors.ustc.edu.cn/crates.io-index",
		"sjtu":      "https://mirrors.sjtug.sjtu.edu.cn/git/crates.io-index",
		"tsinghua":  "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git",
		"bytedance": "https://rsproxy.cn",
	}

	// Ruby 镜像源
	r.mirrors["gem"] = map[string]string{
		"ruby-china": "https://gems.ruby-china.com",
		"tsinghua":   "https://mirrors.tuna.tsinghua.edu.cn/rubygems/",
		"ustc":       "https://mirrors.ustc.edu.cn/rubygems/",
	}

	// Maven 镜像源
	r.mirrors["maven"] = map[string]string{
		"aliyun":  "https://maven.aliyun.com/repository/public",
		"tencent": "https://mirrors.cloud.tencent.com/nexus/repository/maven-public/",
		"huawei":  "https://repo.huaweicloud.com/repository/maven/",
		"netease": "https://mirrors.163.com/maven/repository/maven-public/",
	}

	// Composer (PHP) 镜像源
	r.mirrors["composer"] = map[string]string{
		"aliyun":      "https://mirrors.aliyun.com/composer/",
		"tencent":     "https://mirrors.cloud.tencent.com/composer/",
		"phpcomposer": "https://packagist.phpcomposer.com",
	}

	// Docker 镜像源
	r.mirrors["docker"] = map[string]string{
		"aliyun":  "https://registry.cn-hangzhou.aliyuncs.com",
		"ustc":    "https://docker.mirrors.ustc.edu.cn",
		"163":     "https://hub-mirror.c.163.com",
		"tencent": "https://mirror.ccs.tencentyun.com",
	}

	// Homebrew 镜像源
	r.mirrors["homebrew"] = map[string]string{
		"tsinghua": "https://mirrors.tuna.tsinghua.edu.cn/git/homebrew",
		"ustc":     "https://mirrors.ustc.edu.cn/brew.git",
		"aliyun":   "https://mirrors.aliyun.com/homebrew",
	}

	// Android SDK 镜像源
	r.mirrors["android"] = map[string]string{
		"aliyun":   "https://mirrors.aliyun.com/android.googlesource.com/",
		"tencent":  "https://mirrors.cloud.tencent.com/AndroidSDK/",
		"huawei":   "https://repo.huaweicloud.com/android/",
		"tsinghua": "https://mirrors.tuna.tsinghua.edu.cn/android/",
	}

	// Gradle 镜像源
	r.mirrors["gradle"] = map[string]string{
		"aliyun":   "https://mirrors.aliyun.com/macports/distfiles/gradle/",
		"tencent":  "https://mirrors.cloud.tencent.com/gradle/",
		"huawei":   "https://repo.huaweicloud.com/gradle/",
		"tsinghua": "https://mirrors.tuna.tsinghua.edu.cn/gradle/",
	}
}

// GetMirror 获取指定类型和名称的镜像源URL
func (r *Registry) GetMirror(mirrorType, name string) (string, bool) {
	if mirrors, ok := r.mirrors[mirrorType]; ok {
		if url, exists := mirrors[name]; exists {
			return url, true
		}
	}
	return "", false
}

// GetDefaultMirror 获取指定类型的默认镜像源
func (r *Registry) GetDefaultMirror(mirrorType string) string {
	defaults := map[string]string{
		"npm":      "npmmirror",
		"pip":      "tsinghua",
		"go":       "goproxy",
		"rust":     "ustc",
		"gem":      "ruby-china",
		"maven":    "aliyun",
		"composer": "aliyun",
		"docker":   "aliyun",
		"homebrew": "tsinghua",
		"android":  "aliyun",
		"gradle":   "aliyun",
	}

	if defaultName, ok := defaults[mirrorType]; ok {
		if url, exists := r.GetMirror(mirrorType, defaultName); exists {
			return url
		}
	}
	return ""
}

// ListMirrors 列出指定类型的所有镜像源
func (r *Registry) ListMirrors(mirrorType string) map[string]string {
	if mirrors, ok := r.mirrors[mirrorType]; ok {
		return mirrors
	}
	return make(map[string]string)
}
