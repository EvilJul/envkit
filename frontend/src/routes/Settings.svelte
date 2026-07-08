<script lang="ts">
  import { onMount } from 'svelte';
  import {
    GetSystemInfo,
    GetSettings,
    SaveSettings,
    ResetSettings,
    GetAndroidMirrors,
    GetGradleMirrors,
    ConfigureAndroidMirror,
    ConfigureGradleMirror
  } from '../../wailsjs/go/main/App';

  interface SystemInfo {
    os: string;
    architecture: string;
    distribution: string;
  }

  interface Mirror { value: string; label: string; url: string; }

  let systemInfo: SystemInfo = { os: '', architecture: '', distribution: '' };
  let androidMirrors: Mirror[] = [];
  let gradleMirrors: Mirror[] = [];

  let settings: any = {
    launchAtStartup: false,
    minimizeToTray: true,
    defaultNpmMirror: 'npmmirror',
    defaultPipMirror: 'tsinghua',
    defaultGoMirror: 'goproxy',
    androidMirror: 'aliyun',
    gradleMirror: 'aliyun',
    showExpertOptions: false,
    installDir: '/usr/local',
    logLevel: 'Info'
  };

  let saving = false;
  let androidApplying = false;
  let gradleApplying = false;
  let dirty = false;
  let lastSavedAt = '';

  const mirrorOptions = {
    npm: [
      { value: 'npmmirror', label: 'npmmirror (淘宝镜像)' },
      { value: 'official', label: 'Official (官方源)' }
    ],
    pip: [
      { value: 'tsinghua', label: '清华大学' },
      { value: 'aliyun', label: '阿里云' },
      { value: 'ustc', label: '中科大' },
      { value: 'official', label: 'Official (官方源)' }
    ],
    go: [
      { value: 'goproxy', label: 'goproxy.cn' },
      { value: 'aliyun', label: '阿里云' },
      { value: 'official', label: 'Official (官方源)' }
    ]
  };

  const logLevels = ['Info', 'Debug', 'Warning', 'Error'];

  onMount(async () => {
    try {
      systemInfo = await GetSystemInfo();
      [androidMirrors, gradleMirrors] = await Promise.all([GetAndroidMirrors(), GetGradleMirrors()]);
      const saved = await GetSettings();
      settings = { ...settings, ...saved };
    } catch (err) {
      console.error('Failed to init settings:', err);
    }
  });

  function markDirty() { dirty = true; }

  async function saveSettings() {
    saving = true;
    try {
      await SaveSettings(settings);
      dirty = false;
      lastSavedAt = new Date().toLocaleTimeString();
    } catch (err) {
      alert(`保存失败: ${err}`);
    }
    saving = false;
  }

  async function resetSettings() {
    if (!confirm('确定要重置所有设置为默认值吗？')) return;
    try {
      const fresh = await ResetSettings();
      settings = { ...settings, ...fresh };
      dirty = false;
      lastSavedAt = new Date().toLocaleTimeString();
      alert('设置已重置');
    } catch (err) {
      alert(`重置失败: ${err}`);
    }
  }

  async function applyAndroidMirror() {
    if (!confirm(`确定要应用 Android SDK 镜像源 ${settings.androidMirror} 吗？\n\n将写入 ~/.android/repositories.cfg、~/.gradle/init.d/ 与 shell 配置文件。`)) return;
    androidApplying = true;
    try {
      await ConfigureAndroidMirror(settings.androidMirror);
      await SaveSettings({ ...settings });
      dirty = false;
      lastSavedAt = new Date().toLocaleTimeString();
      alert('Android SDK 镜像源配置成功！\n请重新打开终端或执行 source ~/.zshrc 使环境变量生效。');
    } catch (err) {
      alert(`Android 镜像源配置失败: ${err}`);
    }
    androidApplying = false;
  }

  async function applyGradleMirror() {
    if (!confirm(`确定要应用 Gradle 镜像源 ${settings.gradleMirror} 吗？\n\n将写入 ~/.gradle/init.d/ 与 ~/.gradle/init.gradle。`)) return;
    gradleApplying = true;
    try {
      await ConfigureGradleMirror(settings.gradleMirror);
      await SaveSettings({ ...settings });
      dirty = false;
      lastSavedAt = new Date().toLocaleTimeString();
      alert('Gradle 镜像源配置成功！');
    } catch (err) {
      alert(`Gradle 镜像源配置失败: ${err}`);
    }
    gradleApplying = false;
  }

  function androidMirrorUrl(): string {
    return androidMirrors.find(m => m.value === settings.androidMirror)?.url || '';
  }
  function gradleMirrorUrl(): string {
    return gradleMirrors.find(m => m.value === settings.gradleMirror)?.url || '';
  }
</script>

<div class="page">
  <div class="header">
    <h1>Settings</h1>
    <div class="status">
      {#if dirty}<span class="dirty">● 未保存</span>
      {:else if lastSavedAt}<span class="saved">✓ 已保存于 {lastSavedAt}</span>
      {/if}
    </div>
  </div>

  <div class="settings-container">
    <!-- 系统信息 -->
    <section class="settings-section">
      <h2>系统信息</h2>
      <div class="info-grid">
        <div class="info-item">
          <span class="label">操作系统</span>
          <span class="value">{systemInfo.os || '加载中...'}</span>
        </div>
        <div class="info-item">
          <span class="label">架构</span>
          <span class="value">{systemInfo.architecture || '加载中...'}</span>
        </div>
        {#if systemInfo.distribution}
          <div class="info-item">
            <span class="label">发行版</span>
            <span class="value">{systemInfo.distribution}</span>
          </div>
        {/if}
        <div class="info-item">
          <span class="label">版本</span>
          <span class="value">v0.2.0</span>
        </div>
      </div>
    </section>

    <!-- 通用设置 -->
    <section class="settings-section">
      <h2>通用</h2>
      <div class="setting-item">
        <div class="setting-info">
          <label>开机自启动</label>
          <p>在系统启动时自动启动 EnvKit</p>
        </div>
        <input type="checkbox" bind:checked={settings.launchAtStartup} on:change={markDirty} />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <label>最小化到托盘</label>
          <p>关闭窗口时最小化到系统托盘</p>
        </div>
        <input type="checkbox" bind:checked={settings.minimizeToTray} on:change={markDirty} />
      </div>
    </section>

    <!-- 镜像源设置 -->
    <section class="settings-section">
      <h2>默认镜像源</h2>
      <div class="setting-item">
        <div class="setting-info">
          <label>npm 镜像源</label>
          <p>安装 Node.js 时使用的默认镜像</p>
        </div>
        <select bind:value={settings.defaultNpmMirror} on:change={markDirty}>
          {#each mirrorOptions.npm as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <label>pip 镜像源</label>
          <p>安装 Python 时使用的默认镜像</p>
        </div>
        <select bind:value={settings.defaultPipMirror} on:change={markDirty}>
          {#each mirrorOptions.pip as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <label>Go 镜像源</label>
          <p>安装 Go 时使用的默认镜像</p>
        </div>
        <select bind:value={settings.defaultGoMirror} on:change={markDirty}>
          {#each mirrorOptions.go as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
    </section>

    <!-- Android / Gradle 镜像源配置 -->
    <section class="settings-section">
      <h2>Android / Gradle 镜像源</h2>
      <div class="setting-item">
        <div class="setting-info">
          <label>Android SDK 镜像源</label>
          <p>用于加速下载 Android SDK 组件（cmdline-tools / platform-tools / build-tools / platforms）</p>
        </div>
        <select bind:value={settings.androidMirror} on:change={markDirty}>
          {#each androidMirrors as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
      <div class="url-line">URL：{androidMirrorUrl()}</div>

      <div class="setting-item">
        <div class="setting-info">
          <label>Gradle 镜像源</label>
          <p>用于加速 Gradle 构建时的依赖下载（替换所有 Maven 仓库为镜像源）</p>
        </div>
        <select bind:value={settings.gradleMirror} on:change={markDirty}>
          {#each gradleMirrors as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>
      <div class="url-line">URL：{gradleMirrorUrl()}</div>

      <div class="action-buttons">
        <button class="btn-primary" on:click={applyAndroidMirror} disabled={androidApplying || gradleApplying || saving}>
          {androidApplying ? '配置中...' : '应用 Android 镜像源'}
        </button>
        <button class="btn-secondary" on:click={applyGradleMirror} disabled={androidApplying || gradleApplying || saving}>
          {gradleApplying ? '配置中...' : '应用 Gradle 镜像源'}
        </button>
      </div>
    </section>

    <!-- 高级设置 -->
    <section class="settings-section">
      <h2>高级</h2>
      <div class="setting-item">
        <div class="setting-info">
          <label>显示专家选项</label>
          <p>显示高级用户选项和调试信息</p>
        </div>
        <input type="checkbox" bind:checked={settings.showExpertOptions} on:change={markDirty} />
      </div>
      {#if settings.showExpertOptions}
        <div class="expert-options">
          <div class="setting-item">
            <div class="setting-info">
              <label>安装目录</label>
              <p>工具和语言环境的默认安装位置</p>
            </div>
            <input type="text" bind:value={settings.installDir} on:change={markDirty} />
          </div>
          <div class="setting-item">
            <div class="setting-info">
              <label>日志级别</label>
              <p>应用日志的详细程度</p>
            </div>
            <select bind:value={settings.logLevel} on:change={markDirty}>
              {#each logLevels as lv}<option value={lv}>{lv}</option>{/each}
            </select>
          </div>
        </div>
      {/if}
    </section>

    <!-- 关于 -->
    <section class="settings-section">
      <h2>关于</h2>
      <div class="about-info">
        <div class="app-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none">
            <rect x="3" y="3" width="7" height="7" rx="1" fill="#007aff"/>
            <rect x="14" y="3" width="7" height="7" rx="1" fill="#007aff"/>
            <rect x="3" y="14" width="7" height="7" rx="1" fill="#007aff"/>
            <rect x="14" y="14" width="7" height="7" rx="1" fill="#007aff"/>
          </svg>
        </div>
        <h3>EnvKit</h3>
        <p class="version">版本 0.2.0</p>
        <p class="description">轻量级跨平台开发环境管理工具</p>
        <div class="links">
          <a href="#" on:click|preventDefault={() => alert('https://github.com/fusheng/envkit')}>GitHub 仓库</a>
          <span>·</span>
          <a href="#" on:click|preventDefault={() => alert('查看许可证信息')}>MIT License</a>
          <span>·</span>
          <a href="#" on:click|preventDefault={() => alert('打开文档...')}>文档</a>
        </div>
      </div>
    </section>

    <!-- 操作按钮 -->
    <div class="actions">
      <button class="btn-secondary" on:click={resetSettings} disabled={saving}>重置为默认</button>
      <button class="btn-primary" on:click={saveSettings} disabled={saving || !dirty}>
        {saving ? '保存中...' : (dirty ? '保存设置' : '已是最新')}
      </button>
    </div>
  </div>
</div>

<style>
  .page { max-width: 800px; }
  .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
  h1 { font-size: 24px; font-weight: 600; margin: 0; color: #1d1d1f; }

  .status .dirty { color: #ff9500; font-size: 12px; }
  .status .saved { color: #28a745; font-size: 12px; }

  .settings-container { display: flex; flex-direction: column; gap: 24px; }

  .settings-section {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 20px;
  }
  .settings-section h2 {
    font-size: 15px;
    font-weight: 600;
    margin: 0 0 16px 0;
    color: #1d1d1f;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  .info-item { display: flex; flex-direction: column; gap: 4px; }
  .info-item .label {
    font-size: 11px; font-weight: 600;
    color: #6e6e73;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .info-item .value { font-size: 13px; color: #1d1d1f; }

  .setting-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #e5e5e5;
  }
  .setting-item:last-child { border-bottom: none; }

  .setting-info { flex: 1; }
  .setting-info label {
    display: block; font-size: 13px; font-weight: 500;
    color: #1d1d1f; margin-bottom: 2px;
  }
  .setting-info p { font-size: 12px; color: #6e6e73; margin: 0; }

  input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; }

  input[type="text"], select {
    padding: 6px 12px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    font-size: 13px;
    background: white;
    min-width: 200px;
  }
  input[type="text"]:focus, select:focus {
    outline: none; border-color: #007aff;
    box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
  }

  .url-line {
    margin: 0 0 12px 0;
    font-size: 11px;
    color: #007aff;
    font-family: monospace;
    word-break: break-all;
  }

  .expert-options {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #e5e5e5;
  }

  .about-info {
    display: flex; flex-direction: column; align-items: center;
    text-align: center; padding: 20px 0;
  }
  .app-icon { margin-bottom: 12px; }
  .about-info h3 { font-size: 18px; font-weight: 600; margin: 0 0 4px 0; color: #1d1d1f; }
  .about-info .version { font-size: 13px; color: #6e6e73; margin: 0 0 8px 0; }
  .about-info .description { font-size: 13px; color: #1d1d1f; margin: 0 0 16px 0; }

  .links { display: flex; align-items: center; gap: 8px; font-size: 13px; }
  .links a { color: #007aff; text-decoration: none; }
  .links a:hover { text-decoration: underline; }
  .links span { color: #d1d1d6; }

  .actions { display: flex; justify-content: flex-end; gap: 12px; padding-top: 8px; }

  button {
    padding: 8px 20px;
    border: none; border-radius: 4px;
    font-size: 13px; font-weight: 500;
    cursor: pointer; transition: opacity 0.1s;
  }
  button:hover:not(:disabled) { opacity: 0.8; }
  button:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-primary { background: #007aff; color: white; }
  .btn-secondary { background: #e5e5e5; color: #1d1d1f; }

  .action-buttons { display: flex; gap: 8px; margin-top: 12px; }
  .action-buttons button { flex: 1; }
</style>