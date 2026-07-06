<script lang="ts">
  import { onMount } from 'svelte';
  import { GetTools, InstallTool, UninstallTool, InstallAndroid, UninstallAndroid, GetAndroidInfo } from '../../wailsjs/go/main/App';

  interface Tool {
    name: string;
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

  let tools: Tool[] = [];
  let androidInfo: AndroidInfo = {
    installed: false,
    version: '',
    sdkPath: '',
    adbPath: '',
    hasSdk: false
  };
  let loading = false;

  onMount(async () => {
    await loadTools();
  });

  async function loadTools() {
    loading = true;
    try {
      tools = await GetTools();
      // 加载 Android 详细信息
      try {
        androidInfo = await GetAndroidInfo();
      } catch (err) {
        console.error('Failed to load android info:', err);
      }
    } catch (err) {
      console.error('Failed to load tools:', err);
      alert(`加载失败: ${err}`);
    }
    loading = false;
  }

  async function install(tool: Tool) {
    if (!confirm(`确定要安装 ${tool.displayName} 吗？\n\n这可能需要几分钟时间。`)) return;

    loading = true;
    try {
      await InstallTool(tool.name);
      await loadTools();
      alert(`${tool.displayName} 安装成功！`);
    } catch (err) {
      alert(`安装失败: ${err}`);
    }
    loading = false;
  }

  async function uninstall(tool: Tool) {
    if (!confirm(`确定要卸载 ${tool.displayName} 吗？\n\n这将删除所有文件和配置。`)) return;

    loading = true;
    try {
      await UninstallTool(tool.name);
      await loadTools();
      alert(`${tool.displayName} 卸载成功！`);
    } catch (err) {
      alert(`卸载失败: ${err}`);
    }
    loading = false;
  }

  async function installAndroid() {
    if (!confirm('确定要安装 Android SDK 吗？\n\nAndroid SDK 是一个完整的开发环境套件，包含 cmdline-tools / platform-tools / build-tools / platforms 等组件。\n\n这可能需要较长时间。')) return;

    loading = true;
    try {
      await InstallAndroid();
      await loadTools();
      alert('Android SDK 安装成功！');
    } catch (err) {
      alert(`安装失败: ${err}`);
    }
    loading = false;
  }

  async function uninstallAndroid() {
    if (!confirm('确定要卸载 Android SDK 吗？\n\n这将删除 SDK 文件、清理 ANDROID_HOME 等环境变量配置。')) return;

    loading = true;
    try {
      await UninstallAndroid();
      await loadTools();
      alert('Android SDK 卸载成功！');
    } catch (err) {
      alert(`卸载失败: ${err}`);
    }
    loading = false;
  }

  function getToolIcon(name: string): string {
    const icons: Record<string, string> = {
      git: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
      docker: 'M4 6h2v2H4zm3 0h2v2H7zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm-6 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2z',
      code: 'M16 18l6-6-6-6M8 6l-6 6 6 6',
      conda: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z',
      kubectl: 'M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3z',
      minikube: 'M12 2L2 7l10 5 10-5-10-5z',
      // Android 机器人图标（head + body + 触角）
      android: 'M5 16V8a7 7 0 0 1 14 0v8M5 16h14M5 16a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM19 16a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM8 8V6m8 2V6M9 11h.01M15 11h.01'
    };
    return icons[name] || 'M12 2L2 7l10 5 10-5-10-5z';
  }
</script>

<div class="page">
  <div class="header">
    <h1>Tools</h1>
    <button class="btn-refresh" on:click={loadTools} disabled={loading}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
      </svg>
      {loading ? '加载中...' : '刷新'}
    </button>
  </div>

  <div class="tools-grid">
    {#each tools as tool}
      {#if tool.name === 'android'}
        <div class="tool-card android-card">
          <div class="tool-icon android-icon">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d={getToolIcon(tool.name)}/>
            </svg>
          </div>
          <div class="tool-info">
            <h3>
              {tool.displayName}
              <span class="badge">开发环境</span>
            </h3>
            {#if tool.installed}
              <span class="status installed">✓ 已安装</span>
              {#if androidInfo.version}
                <span class="version">{androidInfo.version}</span>
              {/if}
              {#if androidInfo.sdkPath}
                <div class="sdk-path" title={androidInfo.sdkPath}>
                  SDK: {androidInfo.sdkPath}
                </div>
              {/if}
            {:else}
              <span class="status not-installed">✗ 未安装</span>
              <div class="sdk-desc">包含 cmdline-tools / platform-tools / build-tools / platforms</div>
            {/if}
          </div>
          <div class="tool-actions">
            {#if tool.installed}
              <button class="btn-secondary" on:click={uninstallAndroid} disabled={loading}>
                卸载
              </button>
            {:else}
              <button class="btn-primary" on:click={installAndroid} disabled={loading}>
                安装
              </button>
            {/if}
          </div>
        </div>
      {:else}
        <div class="tool-card">
          <div class="tool-icon">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d={getToolIcon(tool.name)}/>
            </svg>
          </div>
          <div class="tool-info">
            <h3>{tool.displayName}</h3>
            {#if tool.installed}
              <span class="status installed">✓ 已安装</span>
              <span class="version">{tool.version}</span>
            {:else}
              <span class="status not-installed">✗ 未安装</span>
            {/if}
          </div>
          <div class="tool-actions">
            {#if tool.installed}
              <button class="btn-secondary" on:click={() => uninstall(tool)} disabled={loading}>
                卸载
              </button>
            {:else}
              <button class="btn-primary" on:click={() => install(tool)} disabled={loading}>
                安装
              </button>
            {/if}
          </div>
        </div>
      {/if}
    {/each}
  </div>
</div>

<style>
  .page {
    max-width: 1200px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  h1 {
    font-size: 24px;
    font-weight: 600;
    margin: 0;
    color: #1d1d1f;
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .tool-card {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    transition: border-color 0.1s;
  }

  .tool-card:hover {
    border-color: #d1d1d6;
  }

  .tool-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 48px;
    background: #ffffff;
    border-radius: 8px;
    color: #007aff;
  }

  .tool-info {
    flex: 1;
  }

  .tool-info h3 {
    font-size: 15px;
    font-weight: 500;
    margin: 0 0 8px 0;
    color: #1d1d1f;
  }

  .status {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
    margin-right: 8px;
  }

  .status.installed {
    background: #28a74520;
    color: #28a745;
  }

  .status.not-installed {
    background: #86868b20;
    color: #86868b;
  }

  .version {
    font-size: 12px;
    color: #6e6e73;
  }

  .tool-actions {
    display: flex;
    gap: 8px;
  }

  button {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    cursor: pointer;
    transition: opacity 0.1s;
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    justify-content: center;
  }

  button:hover:not(:disabled) {
    opacity: 0.8;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #007aff;
    color: white;
    font-weight: 500;
  }

  .btn-secondary {
    background: #e5e5e5;
    color: #1d1d1f;
  }

  .btn-refresh {
    background: #f5f5f5;
    color: #1d1d1f;
    border: 1px solid #d1d1d6;
  }

  /* Android SDK 卡片特殊样式：突出显示"开发环境套件"特征 */
  .tool-card.android-card {
    background: linear-gradient(135deg, #e8f4ff 0%, #f5f9ff 100%);
    border-color: #a3d3ff;
  }

  .tool-card.android-card:hover {
    border-color: #007aff;
  }

  .android-icon {
    background: #a4c63920;
    color: #3d8b3a;
  }

  .badge {
    display: inline-block;
    padding: 1px 6px;
    margin-left: 6px;
    background: #a4c639;
    color: white;
    border-radius: 3px;
    font-size: 10px;
    font-weight: 500;
    vertical-align: middle;
  }

  .sdk-path {
    margin-top: 4px;
    font-size: 11px;
    color: #6e6e73;
    font-family: monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }

  .sdk-desc {
    margin-top: 4px;
    font-size: 11px;
    color: #6e6e73;
  }
</style>
