<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import ProgressBar from '../lib/components/ProgressBar.svelte';
  import LoadingSkeleton from '../lib/components/LoadingSkeleton.svelte';
  import BrandIcon from '../lib/components/icons/BrandIcon.svelte';
  import { subscribeTaskProgress } from '../lib/taskProgress';
  import {
    GetTools,
    InstallTool,
    UninstallTool,
    InstallAndroid,
    UninstallAndroid,
    GetAndroidInfo,
    GetAndroidMirrors,
    GetGradleMirrors,
    ConfigureAndroidMirror,
    ConfigureGradleMirror,
    GetSettings,
    SaveSettings
  } from '../../wailsjs/go/main/App';

  interface Tool {
    name: string;
    icon?: string;
    displayName: string;
    version: string;
    installed: boolean;
  }

  interface AndroidInfo {
    installed: boolean;
    version: string;
    sdkPath: string;
    adbPath: string;
    hasSdk: boolean;
  }

  interface Mirror { value: string; label: string; url: string; }

  interface ProgressInfo {
    percent: number;
    stage: string;
    message: string;
    status: 'running' | 'success' | 'error';
  }

  let tools: Tool[] = [];
  let androidInfo: AndroidInfo = {
    installed: false, version: '', sdkPath: '', adbPath: '', hasSdk: false
  };
  let androidMirrors: Mirror[] = [];
  let gradleMirrors: Mirror[] = [];
  let androidMirror = 'aliyun';
  let gradleMirror = 'aliyun';
  let loading = false;
  let androidApplying = false;
  let gradleApplying = false;
  let progressMap: Record<string, ProgressInfo> = {};
  let androidMirrorProgress: ProgressInfo | null = null;
  let gradleMirrorProgress: ProgressInfo | null = null;
  let unsubProgress: (() => void) | null = null;

  const SUCCESS_HOLD_MS = 1200;
  const ERROR_HOLD_MS = 3000;

  const TOOL_DESCRIPTIONS: Record<string, string> = {
    git: '分布式版本控制系统',
    docker: '容器化应用打包与运行时',
    make: '构建自动化工具',
    cmake: '跨平台构建系统生成器',
    curl: '命令行数据传输工具',
    android: 'Android 应用开发 SDK'
  };

  onMount(async () => {
    unsubProgress = subscribeTaskProgress((evt) => {
      const id = evt.taskId;
      if (!id) return;
      const status: ProgressInfo['status'] =
        evt.stage === 'error' ? 'error' :
        evt.stage === 'done' ? 'success' : 'running';
      const info: ProgressInfo = {
        percent: evt.percent,
        stage: evt.stage,
        message: evt.message || '',
        status
      };
      if (id === 'android-mirror') {
        androidMirrorProgress = info;
      } else if (id === 'gradle-mirror') {
        gradleMirrorProgress = info;
      } else {
        progressMap = { ...progressMap, [id]: info };
      }
    });
    await loadAll();
  });

  onDestroy(() => {
    unsubProgress?.();
  });

  async function loadAll() {
    loading = true;
    try {
      tools = await GetTools();
      androidInfo = await GetAndroidInfo();
      [androidMirrors, gradleMirrors] = await Promise.all([GetAndroidMirrors(), GetGradleMirrors()]);
      const settings = await GetSettings();
      androidMirror = settings.androidMirror;
      gradleMirror = settings.gradleMirror;
    } catch (err) {
      console.error('Failed to load tools:', err);
      alert(`加载失败: ${err}`);
    }
    loading = false;
  }

  function clearProgressLater(key: string, ms: number) {
    setTimeout(() => {
      if (key === 'android-mirror') {
        androidMirrorProgress = null;
      } else if (key === 'gradle-mirror') {
        gradleMirrorProgress = null;
      } else {
        const { [key]: _, ...rest } = progressMap;
        progressMap = rest;
      }
    }, ms);
  }

  async function install(tool: Tool) {
    if (!confirm(`确定要安装 ${tool.displayName} 吗？\n\n这可能需要几分钟时间。`)) return;

    loading = true;
    const key = tool.name;
    progressMap = {
      ...progressMap,
      [key]: { percent: 0, stage: 'preparing', message: '准备安装...', status: 'running' }
    };
    try {
      if (tool.name === 'android') {
        await InstallAndroid();
      } else {
        await InstallTool(tool.name);
      }
      progressMap = {
        ...progressMap,
        [key]: { percent: 100, stage: 'done', message: '安装完成', status: 'success' }
      };
      await loadAll();
      alert(`${tool.displayName} 安装成功！`);
      clearProgressLater(key, SUCCESS_HOLD_MS);
    } catch (err) {
      progressMap = {
        ...progressMap,
        [key]: { percent: 0, stage: 'error', message: String(err), status: 'error' }
      };
      clearProgressLater(key, ERROR_HOLD_MS);
      alert(`安装失败: ${err}`);
    }
    loading = false;
  }

  async function uninstall(tool: Tool) {
    if (!confirm(`确定要卸载 ${tool.displayName} 吗？\n\n这将删除所有文件和配置。`)) return;

    loading = true;
    const key = tool.name;
    progressMap = {
      ...progressMap,
      [key]: { percent: 0, stage: 'preparing', message: '准备卸载...', status: 'running' }
    };
    try {
      if (tool.name === 'android') {
        await UninstallAndroid();
      } else {
        await UninstallTool(tool.name);
      }
      progressMap = {
        ...progressMap,
        [key]: { percent: 100, stage: 'done', message: '卸载完成', status: 'success' }
      };
      await loadAll();
      alert(`${tool.displayName} 卸载成功！`);
      clearProgressLater(key, SUCCESS_HOLD_MS);
    } catch (err) {
      progressMap = {
        ...progressMap,
        [key]: { percent: 0, stage: 'error', message: String(err), status: 'error' }
      };
      clearProgressLater(key, ERROR_HOLD_MS);
      alert(`卸载失败: ${err}`);
    }
    loading = false;
  }

  async function applyAndroidMirror() {
    if (!confirm(`确定要应用 Android SDK 镜像源 ${androidMirror} 吗？\n\n将写入 ~/.android/repositories.cfg、~/.gradle/init.d/ 与 shell 配置文件。`)) return;
    androidApplying = true;
    androidMirrorProgress = { percent: 0, stage: 'configuring', message: '准备配置...', status: 'running' };
    try {
      await ConfigureAndroidMirror(androidMirror);
      const settings = await GetSettings();
      await SaveSettings({ ...settings, androidMirror });
      androidMirrorProgress = { percent: 100, stage: 'done', message: '配置完成', status: 'success' };
      alert('Android SDK 镜像源配置成功！\n请重新打开终端或执行 source ~/.zshrc 使环境变量生效。');
      clearProgressLater('android-mirror', SUCCESS_HOLD_MS);
    } catch (err) {
      androidMirrorProgress = { percent: 0, stage: 'error', message: String(err), status: 'error' };
      clearProgressLater('android-mirror', ERROR_HOLD_MS);
      alert(`Android 镜像源配置失败: ${err}`);
    }
    androidApplying = false;
  }

  async function applyGradleMirror() {
    if (!confirm(`确定要应用 Gradle 镜像源 ${gradleMirror} 吗？\n\n将写入 ~/.gradle/init.d/aliyun.gradle 与 ~/.gradle/init.gradle。`)) return;
    gradleApplying = true;
    gradleMirrorProgress = { percent: 0, stage: 'configuring', message: '准备配置...', status: 'running' };
    try {
      await ConfigureGradleMirror(gradleMirror);
      const settings = await GetSettings();
      await SaveSettings({ ...settings, gradleMirror });
      gradleMirrorProgress = { percent: 100, stage: 'done', message: '配置完成', status: 'success' };
      alert('Gradle 镜像源配置成功！');
      clearProgressLater('gradle-mirror', SUCCESS_HOLD_MS);
    } catch (err) {
      gradleMirrorProgress = { percent: 0, stage: 'error', message: String(err), status: 'error' };
      clearProgressLater('gradle-mirror', ERROR_HOLD_MS);
      alert(`Gradle 镜像源配置失败: ${err}`);
    }
    gradleApplying = false;
  }

</script>

<div class="page">
  <div class="header">
    <h1>Tools</h1>
    <button class="btn-refresh" on:click={loadAll} disabled={loading}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
      </svg>
      {loading ? '加载中...' : '刷新'}
    </button>
  </div>

  {#if loading && tools.length === 0}
    <LoadingSkeleton count={4} columns="repeat(auto-fill, minmax(280px, 1fr))" />
  {:else}
  <div class="tools-grid">
    {#each tools as tool (tool.name)}
      <div class="tool-card" class:android-card={tool.name === 'android'}>
        <div class="tool-icon">
          <BrandIcon name={tool.icon || tool.name} size={32} />
        </div>
        <div class="tool-info">
          <h3>
            {tool.displayName}
            {#if tool.name === 'android'}<span class="badge">开发环境</span>{/if}
          </h3>
          <p class="description">{TOOL_DESCRIPTIONS[tool.name] || ''}</p>
          {#if tool.installed}
            <span class="status installed">✓ 已安装</span>
            {#if tool.version}<span class="version">{tool.version}</span>{/if}
            {#if tool.name === 'android' && androidInfo.sdkPath}
              <div class="sdk-path" title={androidInfo.sdkPath}>SDK: {androidInfo.sdkPath}</div>
              <div class="sdk-path" title={androidInfo.adbPath}>ADB: {androidInfo.adbPath}</div>
            {/if}
          {:else}
            <span class="status not-installed">✗ 未安装</span>
            {#if tool.name === 'android'}
              <div class="sdk-desc">包含 cmdline-tools / platform-tools / build-tools / platforms</div>
            {/if}
          {/if}
        </div>
        <div class="tool-actions">
          {#if tool.installed}
            <button class="btn-secondary" on:click={() => uninstall(tool)} disabled={loading}>卸载</button>
          {:else}
            <button class="btn-primary" on:click={() => install(tool)} disabled={loading}>安装</button>
          {/if}
        </div>
        {#if progressMap[tool.name]}
          <div class="tool-progress">
            <ProgressBar
              percent={progressMap[tool.name].percent}
              stage={progressMap[tool.name].stage}
              message={progressMap[tool.name].message}
              status={progressMap[tool.name].status}
            />
          </div>
        {/if}
      </div>
    {/each}
  </div>

  <section class="mirror-section">
    <h2>Android / Gradle 镜像源</h2>
    <p class="hint">配置国内镜像以加速 SDK 与依赖下载。配置后立即生效。</p>

    <div class="mirror-grid">
      <div class="mirror-card">
        <label class="mirror-label">Android SDK 镜像源</label>
        <select bind:value={androidMirror} disabled={androidApplying || gradleApplying}>
          {#each androidMirrors as m}
            <option value={m.value}>{m.label}</option>
          {/each}
        </select>
        <div class="mirror-url">{androidMirrors.find(m => m.value === androidMirror)?.url || ''}</div>
        <button class="btn-primary" on:click={applyAndroidMirror} disabled={androidApplying || gradleApplying}>
          {androidApplying ? '配置中...' : '应用 Android 镜像源'}
        </button>
        {#if androidMirrorProgress}
          <ProgressBar
            percent={androidMirrorProgress.percent}
            stage={androidMirrorProgress.stage}
            message={androidMirrorProgress.message}
            status={androidMirrorProgress.status}
          />
        {/if}
      </div>

      <div class="mirror-card">
        <label class="mirror-label">Gradle 镜像源</label>
        <select bind:value={gradleMirror} disabled={androidApplying || gradleApplying}>
          {#each gradleMirrors as m}
            <option value={m.value}>{m.label}</option>
          {/each}
        </select>
        <div class="mirror-url">{gradleMirrors.find(m => m.value === gradleMirror)?.url || ''}</div>
        <button class="btn-primary" on:click={applyGradleMirror} disabled={androidApplying || gradleApplying}>
          {gradleApplying ? '配置中...' : '应用 Gradle 镜像源'}
        </button>
        {#if gradleMirrorProgress}
          <ProgressBar
            percent={gradleMirrorProgress.percent}
            stage={gradleMirrorProgress.stage}
            message={gradleMirrorProgress.message}
            status={gradleMirrorProgress.status}
          />
        {/if}
      </div>
    </div>
  </section>
</div>

<style>
  .page { max-width: 1200px; }

  .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
  h1 { font-size: 24px; font-weight: 600; margin: 0; color: #1d1d1f; }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .tool-card {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 8px;
    padding: 20px;
    display: flex; flex-direction: column; gap: 12px;
    transition: border-color 0.15s;
  }
  .tool-card:hover { border-color: #d1d1d6; }

  .tool-icon {
    width: 48px; height: 48px;
    display: flex; align-items: center; justify-content: center;
    background: #ffffff;
    border-radius: 10px;
  }

  .tool-info { flex: 1; }
  .tool-info h3 { font-size: 15px; font-weight: 600; margin: 0 0 4px 0; color: #1d1d1f; display: flex; align-items: center; gap: 6px; }
  .tool-info .description { font-size: 12px; color: #6e6e73; margin: 0 0 8px 0; }

  .status {
    display: inline-block; padding: 2px 8px; border-radius: 4px;
    font-size: 12px; margin-right: 8px;
  }
  .status.installed { background: #28a74520; color: #28a745; }
  .status.not-installed { background: #86868b20; color: #86868b; }
  .version { font-size: 12px; color: #6e6e73; }

  .tool-actions { display: flex; gap: 8px; }

  button {
    padding: 6px 16px; border: none; border-radius: 4px; font-size: 13px;
    cursor: pointer; transition: opacity 0.1s;
    display: flex; align-items: center; gap: 6px;
    flex: 1; justify-content: center;
  }
  button:hover:not(:disabled) { opacity: 0.8; }
  button:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-primary { background: #007aff; color: white; font-weight: 500; }
  .btn-secondary { background: #e5e5e5; color: #1d1d1f; }
  .btn-refresh { background: #f5f5f5; color: #1d1d1f; border: 1px solid #d1d1d6; flex: 0 0 auto; }

  .android-card {
    background: linear-gradient(135deg, #e8f4ff 0%, #f5f9ff 100%);
    border-color: #a3d3ff;
  }
  .android-card:hover { border-color: #007aff; }

  .badge {
    display: inline-block;
    padding: 1px 6px;
    background: #a4c639;
    color: white;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 500;
  }

  .sdk-path {
    margin-top: 4px;
    font-size: 11px;
    color: #6e6e73;
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sdk-desc { margin-top: 4px; font-size: 11px; color: #6e6e73; }

  .mirror-section {
    margin-top: 32px;
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 8px;
    padding: 24px;
  }
  .mirror-section h2 { font-size: 16px; font-weight: 600; margin: 0 0 4px 0; color: #1d1d1f; }
  .mirror-section .hint { font-size: 12px; color: #6e6e73; margin: 0 0 20px 0; }

  .mirror-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
  }
  .mirror-card {
    background: white;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .mirror-label { font-size: 12px; font-weight: 600; color: #6e6e73; text-transform: uppercase; letter-spacing: 0.5px; }
  .mirror-card select {
    padding: 8px 12px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    font-size: 13px;
    background: white;
  }
  .mirror-card select:focus { outline: none; border-color: #007aff; }
  .mirror-url {
    font-size: 11px;
    color: #007aff;
    font-family: monospace;
    word-break: break-all;
  }
  .mirror-card .btn-primary { margin-top: 8px; }
</style>