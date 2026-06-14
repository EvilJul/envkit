<script lang="ts">
  import { onMount } from 'svelte';
  import { GetLanguages, InstallLanguage, UninstallLanguage } from '../../wailsjs/go/main/App';

  interface Language {
    name: string;
    displayName: string;
    version: string;
    installed: boolean;
    mirror: string;
  }

  let languages: Language[] = [];
  let loading = false;

  onMount(async () => {
    await loadLanguages();
  });

  async function loadLanguages() {
    loading = true;
    try {
      languages = await GetLanguages();
    } catch (err) {
      console.error('Failed to load languages:', err);
      alert(`加载失败: ${err}`);
    }
    loading = false;
  }

  async function install(lang: Language) {
    if (!confirm(`确定要安装 ${lang.displayName} 吗？\n\n这可能需要几分钟时间。`)) return;

    loading = true;
    try {
      await InstallLanguage(lang.name, lang.version || 'latest');
      await loadLanguages();
      alert(`${lang.displayName} 安装成功！`);
    } catch (err) {
      alert(`安装失败: ${err}`);
    }
    loading = false;
  }

  async function uninstall(lang: Language) {
    if (!confirm(`确定要卸载 ${lang.displayName} 吗？\n\n这将删除所有文件和环境变量配置。`)) return;

    loading = true;
    try {
      await UninstallLanguage(lang.name);
      await loadLanguages();
      alert(`${lang.displayName} 卸载成功！`);
    } catch (err) {
      alert(`卸载失败: ${err}`);
    }
    loading = false;
  }
</script>

<div class="page">
  <div class="header">
    <h1>Languages</h1>
    <button class="btn-refresh" on:click={loadLanguages} disabled={loading}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
      </svg>
      {loading ? '加载中...' : '刷新'}
    </button>
  </div>

  <div class="language-list">
    {#each languages as lang}
      <div class="language-card">
        <div class="card-header">
          <div class="language-info">
            <h3>{lang.displayName}</h3>
            {#if lang.installed}
              <span class="status installed">✓ 已安装</span>
              <span class="version">{lang.version}</span>
            {:else}
              <span class="status not-installed">✗ 未安装</span>
            {/if}
          </div>
          <div class="actions">
            {#if lang.installed}
              <button class="btn-secondary" on:click={() => uninstall(lang)} disabled={loading}>
                卸载
              </button>
            {:else}
              <button class="btn-primary" on:click={() => install(lang)} disabled={loading}>
                安装
              </button>
            {/if}
          </div>
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

  .language-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .language-card {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 16px;
    transition: border-color 0.1s;
  }

  .language-card:hover {
    border-color: #d1d1d6;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .language-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .language-info h3 {
    font-size: 15px;
    font-weight: 500;
    margin: 0;
    color: #1d1d1f;
  }

  .status {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
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

  .actions {
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
