<script lang="ts">
  import { onMount } from 'svelte';
  import { GetTools, InstallTool, UninstallTool } from '../../wailsjs/go/main/App';

  interface Tool {
    name: string;
    displayName: string;
    version: string;
    installed: boolean;
  }

  let tools: Tool[] = [];
  let loading = false;

  onMount(async () => {
    await loadTools();
  });

  async function loadTools() {
    loading = true;
    try {
      tools = await GetTools();
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

  function getToolIcon(name: string): string {
    const icons: Record<string, string> = {
      git: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
      docker: 'M4 6h2v2H4zm3 0h2v2H7zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm-6 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2z',
      code: 'M16 18l6-6-6-6M8 6l-6 6 6 6',
      conda: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z',
      kubectl: 'M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3z',
      minikube: 'M12 2L2 7l10 5 10-5-10-5z'
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
</style>
