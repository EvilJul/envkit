<script lang="ts">
  import { onMount } from 'svelte';
  import { GetSystemInfo } from '../../wailsjs/go/main/App';

  interface SystemInfo {
    os: string;
    architecture: string;
    distribution: string;
  }

  let systemInfo: SystemInfo = {
    os: '',
    architecture: '',
    distribution: ''
  };

  let settings = {
    launchAtStartup: false,
    minimizeToTray: true,
    defaultNpmMirror: 'npmmirror',
    defaultPipMirror: 'tsinghua',
    defaultGoMirror: 'goproxy',
    showExpertOptions: false
  };

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

  onMount(async () => {
    try {
      systemInfo = await GetSystemInfo();
    } catch (err) {
      console.error('Failed to get system info:', err);
    }
  });

  function saveSettings() {
    // TODO: 实现设置保存
    alert('设置已保存！\n\n注意：部分设置需要重启应用生效。');
  }

  function resetSettings() {
    if (!confirm('确定要重置所有设置为默认值吗？')) return;

    settings = {
      launchAtStartup: false,
      minimizeToTray: true,
      defaultNpmMirror: 'npmmirror',
      defaultPipMirror: 'tsinghua',
      defaultGoMirror: 'goproxy',
      showExpertOptions: false
    };
    alert('设置已重置');
  }
</script>

<div class="page">
  <h1>Settings</h1>

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
        <input type="checkbox" bind:checked={settings.launchAtStartup} />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <label>最小化到托盘</label>
          <p>关闭窗口时最小化到系统托盘</p>
        </div>
        <input type="checkbox" bind:checked={settings.minimizeToTray} />
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
        <select bind:value={settings.defaultNpmMirror}>
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
        <select bind:value={settings.defaultPipMirror}>
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
        <select bind:value={settings.defaultGoMirror}>
          {#each mirrorOptions.go as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
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
        <input type="checkbox" bind:checked={settings.showExpertOptions} />
      </div>
      {#if settings.showExpertOptions}
        <div class="expert-options">
          <div class="setting-item">
            <div class="setting-info">
              <label>安装目录</label>
              <p>工具和语言环境的默认安装位置</p>
            </div>
            <input type="text" value="/usr/local" disabled />
          </div>
          <div class="setting-item">
            <div class="setting-info">
              <label>日志级别</label>
              <p>应用日志的详细程度</p>
            </div>
            <select>
              <option>Info</option>
              <option>Debug</option>
              <option>Warning</option>
              <option>Error</option>
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
          <a href="#" on:click|preventDefault={() => alert('https://github.com/fusheng/envkit')}>
            GitHub 仓库
          </a>
          <span>·</span>
          <a href="#" on:click|preventDefault={() => alert('查看许可证信息')}>
            MIT License
          </a>
          <span>·</span>
          <a href="#" on:click|preventDefault={() => alert('打开文档...')}>
            文档
          </a>
        </div>
      </div>
    </section>

    <!-- 操作按钮 -->
    <div class="actions">
      <button class="btn-secondary" on:click={resetSettings}>
        重置为默认
      </button>
      <button class="btn-primary" on:click={saveSettings}>
        保存设置
      </button>
    </div>
  </div>
</div>

<style>
  .page {
    max-width: 800px;
  }

  h1 {
    font-size: 24px;
    font-weight: 600;
    margin: 0 0 20px 0;
    color: #1d1d1f;
  }

  .settings-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

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

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .info-item .label {
    font-size: 11px;
    font-weight: 600;
    color: #6e6e73;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .info-item .value {
    font-size: 13px;
    color: #1d1d1f;
  }

  .setting-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #e5e5e5;
  }

  .setting-item:last-child {
    border-bottom: none;
  }

  .setting-info {
    flex: 1;
  }

  .setting-info label {
    display: block;
    font-size: 13px;
    font-weight: 500;
    color: #1d1d1f;
    margin-bottom: 2px;
  }

  .setting-info p {
    font-size: 12px;
    color: #6e6e73;
    margin: 0;
  }

  input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  input[type="text"],
  select {
    padding: 6px 12px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    font-size: 13px;
    background: white;
    min-width: 200px;
  }

  input[type="text"]:focus,
  select:focus {
    outline: none;
    border-color: #007aff;
    box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
  }

  input[type="text"]:disabled {
    background: #f5f5f5;
    color: #86868b;
    cursor: not-allowed;
  }

  .expert-options {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #e5e5e5;
  }

  .about-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 20px 0;
  }

  .app-icon {
    margin-bottom: 12px;
  }

  .about-info h3 {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #1d1d1f;
  }

  .about-info .version {
    font-size: 13px;
    color: #6e6e73;
    margin: 0 0 8px 0;
  }

  .about-info .description {
    font-size: 13px;
    color: #1d1d1f;
    margin: 0 0 16px 0;
  }

  .links {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }

  .links a {
    color: #007aff;
    text-decoration: none;
  }

  .links a:hover {
    text-decoration: underline;
  }

  .links span {
    color: #d1d1d6;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding-top: 8px;
  }

  button {
    padding: 8px 20px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.1s;
  }

  button:hover {
    opacity: 0.8;
  }

  .btn-primary {
    background: #007aff;
    color: white;
  }

  .btn-secondary {
    background: #e5e5e5;
    color: #1d1d1f;
  }
</style>
