<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import {
    GetLanguages,
    InstallLanguage,
    UninstallLanguage,
    SetLanguageMirror
  } from '../../wailsjs/go/main/App';
  import ProgressBar from '../lib/components/ProgressBar.svelte';
  import LoadingSkeleton from '../lib/components/LoadingSkeleton.svelte';
  import BrandIcon from '../lib/components/icons/BrandIcon.svelte';
  import { subscribeTaskProgress } from '../lib/taskProgress';

  interface Language {
    name: string;
    displayName: string;
    description: string;
    icon: string;
    installed: boolean;
    version: string;
    mirror: string;
    availableVersions: { version: string; size: string }[];
    mirrorOptions: string[];
  }

  interface ProgressInfo {
    percent: number;
    stage: string;
    message: string;
    status: 'running' | 'success' | 'error';
  }

  const MIRROR_LABELS: Record<string, string> = {
    official: 'Official',
    npmmirror: 'npmmirror',
    goproxy: 'goproxy.cn',
    aliyun: '阿里云',
    tsinghua: '清华大学',
    ustc: '中科大'
  };

  let languages: Language[] = [];
  let loading = false;

  // 安装模态框
  let showInstallModal = false;
  let selectedLang: Language | null = null;
  let installVersion = '';

  // 卸载确认模态框（替代原生 confirm，避免被跳过/被禁用）
  let showUninstallModal = false;
  let uninstallTarget: Language | null = null;
  let uninstallConfirmText = '';

  // 进度：每个语言卡片独立展示
  let progressMap: Record<string, ProgressInfo> = {};
  let progressTimers: Record<string, ReturnType<typeof setTimeout>> = {};
  let unsubProgress: (() => void) | null = null;

  const SUCCESS_HOLD_MS = 1500;
  const ERROR_HOLD_MS = 3000;

  onMount(async () => {
    unsubProgress = subscribeTaskProgress((evt) => {
      const id = evt.taskId;
      if (!id) return;
      const status: ProgressInfo['status'] =
        evt.stage === 'error' ? 'error' :
        evt.stage === 'done' ? 'success' : 'running';
      setProgress(id, {
        percent: evt.percent,
        stage: evt.stage,
        message: evt.message || '',
        status
      });
    });
    await loadLanguages();
  });

  onDestroy(() => {
    unsubProgress?.();
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

  function setProgress(name: string, info: ProgressInfo) {
    progressMap = { ...progressMap, [name]: info };
  }

  function clearProgressLater(name: string, ms: number) {
    if (progressTimers[name]) clearTimeout(progressTimers[name]);
    progressTimers[name] = setTimeout(() => {
      const next = { ...progressMap };
      delete next[name];
      progressMap = next;
    }, ms);
  }

  function openInstallModal(lang: Language) {
    selectedLang = lang;
    installVersion = lang.availableVersions?.[0]?.version || 'latest';
    showInstallModal = true;
  }

  function openUninstallModal(lang: Language) {
    uninstallTarget = lang;
    uninstallConfirmText = '';
    showUninstallModal = true;
  }

  async function confirmInstall() {
    if (!selectedLang) return;
    const name = selectedLang.name;
    const version = installVersion;
    const displayName = selectedLang.displayName;
    showInstallModal = false;
    loading = true;

    setProgress(name, { percent: 0, stage: 'downloading', message: '准备下载...', status: 'running' });

    try {
      await InstallLanguage(name, version);
      setProgress(name, { percent: 100, stage: 'done', message: '安装完成', status: 'success' });
      await loadLanguages();
      clearProgressLater(name, SUCCESS_HOLD_MS);
      alert(`${displayName} ${version} 安装成功！`);
    } catch (err) {
      setProgress(name, { percent: 100, stage: 'error', message: String(err), status: 'error' });
      clearProgressLater(name, ERROR_HOLD_MS);
      alert(`安装失败: ${err}`);
    }
    loading = false;
    selectedLang = null;
  }

  async function confirmUninstall() {
    if (!uninstallTarget) return;
    if (uninstallConfirmText !== uninstallTarget.displayName) {
      alert(`请输入 "${uninstallTarget.displayName}" 以确认卸载`);
      return;
    }
    const lang = uninstallTarget;
    const name = lang.name;
    showUninstallModal = false;
    loading = true;

    setProgress(name, { percent: 0, stage: 'stopping', message: '停止进程...', status: 'running' });

    try {
      await UninstallLanguage(name);
      setProgress(name, { percent: 100, stage: 'done', message: '卸载完成', status: 'success' });
      await loadLanguages();
      clearProgressLater(name, SUCCESS_HOLD_MS);
      alert(`${lang.displayName} 卸载成功！`);
    } catch (err) {
      setProgress(name, { percent: 100, stage: 'error', message: String(err), status: 'error' });
      clearProgressLater(name, ERROR_HOLD_MS);
      alert(`卸载失败: ${err}`);
    }
    loading = false;
    uninstallTarget = null;
    uninstallConfirmText = '';
  }

  async function changeMirror(lang: Language, mirror: string) {
    if (lang.mirror === mirror) return;
    loading = true;
    try {
      await SetLanguageMirror(lang.name, mirror);
      await loadLanguages();
    } catch (err) {
      alert(`切换镜像源失败: ${err}`);
    }
    loading = false;
  }

  function mirrorLabel(value: string): string {
    return MIRROR_LABELS[value] || value;
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

  {#if loading && languages.length === 0}
    <LoadingSkeleton count={6} columns="repeat(auto-fill, minmax(320px, 1fr))" />
  {:else}
  <div class="language-grid">
    {#each languages as lang (lang.name)}
      <div class="language-card" class:installed={lang.installed} class:busy={!!progressMap[lang.name]}>
        <div class="card-icon">
          <BrandIcon name={lang.icon || lang.name} size={32} />
        </div>

        <div class="card-body">
          <div class="card-top">
            <h3>{lang.displayName}</h3>
            {#if progressMap[lang.name]?.status === 'running'}
              <span class="badge running">
                <span class="spinner"></span>
                进行中…
              </span>
            {:else if lang.installed}
              <span class="badge installed">✓ 已安装</span>
            {:else}
              <span class="badge not-installed">未安装</span>
            {/if}
          </div>

          <p class="description">{lang.description}</p>

          {#if lang.installed}
            <div class="meta">
              <span class="meta-label">版本：</span>
              <span class="meta-value">{lang.version || 'latest'}</span>
            </div>
          {/if}

          <div class="mirror-row">
            <span class="meta-label">镜像源：</span>
            <select
              class="mirror-select"
              value={lang.mirror}
              disabled={loading || !!progressMap[lang.name]}
              on:change={(e) => changeMirror(lang, e.currentTarget.value)}
            >
              {#each lang.mirrorOptions as m}
                <option value={m}>{mirrorLabel(m)}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="card-actions">
          {#if lang.installed}
            <button
              class="btn-secondary"
              on:click={() => openUninstallModal(lang)}
              disabled={loading || !!progressMap[lang.name]}
            >
              {progressMap[lang.name]?.status === 'running' ? '卸载中…' : '卸载'}
            </button>
          {:else}
            <button
              class="btn-primary"
              on:click={() => openInstallModal(lang)}
              disabled={loading || !!progressMap[lang.name]}
            >
              安装
            </button>
          {/if}
        </div>

        {#if progressMap[lang.name]}
          <div class="card-progress">
            <ProgressBar
              percent={progressMap[lang.name].percent}
              stage={progressMap[lang.name].stage}
              message={progressMap[lang.name].message}
              status={progressMap[lang.name].status}
            />
          </div>
        {/if}
      </div>
    {/each}
  </div>
  {/if}
</div>

<!-- 安装模态框：版本选择 -->
{#if showInstallModal && selectedLang}
  <div class="modal-overlay" on:click={() => showInstallModal = false}>
    <div class="modal" on:click|stopPropagation>
      <h2>安装 {selectedLang.displayName}</h2>

      <div class="form-group">
        <label for="install-version">选择版本</label>
        <select id="install-version" bind:value={installVersion}>
          {#each selectedLang.availableVersions as v}
            <option value={v.version}>{v.version}（{v.size}）</option>
          {/each}
        </select>
      </div>

      <div class="info-note">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#6e6e73" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <span>将使用镜像源 {mirrorLabel(selectedLang.mirror)} 进行下载。安装过程可能需要数分钟。</span>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" on:click={() => showInstallModal = false}>取消</button>
        <button class="btn-primary" on:click={confirmInstall}>开始安装</button>
      </div>
    </div>
  </div>
{/if}

<!-- 卸载确认模态框：必须输入名称确认 -->
{#if showUninstallModal && uninstallTarget}
  <div class="modal-overlay" on:click={() => (showUninstallModal = false)}>
    <div class="modal modal-danger" on:click|stopPropagation>
      <h2>⚠️ 卸载 {uninstallTarget.displayName}</h2>

      <p class="danger-desc">
        此操作将删除 {uninstallTarget.displayName} 的所有文件与环境变量配置（{uninstallTarget.description ? uninstallTarget.description + ')' : ''}。<br/>
        请输入 <strong>{uninstallTarget.displayName}</strong> 以确认卸载：
      </p>

      <div class="form-group">
        <input
          type="text"
          class="confirm-input"
          bind:value={uninstallConfirmText}
          placeholder={uninstallTarget.displayName}
          autocomplete="off"
        />
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" on:click={() => (showUninstallModal = false)}>取消</button>
        <button
          class="btn-danger"
          disabled={uninstallConfirmText !== uninstallTarget.displayName}
          on:click={confirmUninstall}
        >
          确认卸载
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page { max-width: 1200px; }

  .header {
    display: flex; justify-content: space-between; align-items: center;
    margin-bottom: 20px;
  }

  h1 { font-size: 24px; font-weight: 600; margin: 0; color: #1d1d1f; }

  .language-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
  }

  .language-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 8px;
    padding: 16px;
    transition: border-color 0.15s;
    min-height: 180px;
  }
  .language-card:hover { border-color: #d1d1d6; }
  .language-card.installed { border-color: #b9e3c4; background: #f4faf5; }
  .language-card.busy { border-color: #007aff; box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.08); }

  .card-icon {
    width: 48px; height: 48px;
    display: flex; align-items: center; justify-content: center;
    background: white;
    border-radius: 10px;
  }

  .card-body { flex: 1; min-width: 0; }

  .card-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
  .card-top h3 { margin: 0; font-size: 15px; font-weight: 600; color: #1d1d1f; }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
  }
  .badge.installed { background: #28a74520; color: #28a745; }
  .badge.not-installed { background: #86868b20; color: #86868b; }
  .badge.running { background: #007aff20; color: #007aff; }

  .spinner {
    width: 10px; height: 10px;
    border: 2px solid #007aff40;
    border-top-color: #007aff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    display: inline-block;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  .description {
    margin: 6px 0 8px 0;
    font-size: 12px;
    color: #6e6e73;
    line-height: 1.5;
  }

  .meta, .mirror-row {
    display: flex; align-items: center; gap: 6px;
    margin-top: 4px;
    font-size: 12px;
  }
  .meta-label { color: #86868b; }
  .meta-value { color: #1d1d1f; font-weight: 500; }

  .mirror-select {
    padding: 3px 8px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    background: white;
    font-size: 12px;
    color: #1d1d1f;
    cursor: pointer;
  }
  .mirror-select:disabled { opacity: 0.5; cursor: not-allowed; }

  .card-actions {
    display: flex;
    gap: 8px;
  }

  .card-progress {
    margin-top: 4px;
    padding-top: 8px;
    border-top: 1px dashed #d1d1d6;
  }

  button {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    cursor: pointer;
    transition: opacity 0.1s;
    display: flex; align-items: center; gap: 6px;
  }
  button:hover:not(:disabled) { opacity: 0.8; }
  button:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-primary { background: #007aff; color: white; font-weight: 500; flex: 1; justify-content: center; }
  .btn-secondary { background: #e5e5e5; color: #1d1d1f; flex: 1; justify-content: center; }
  .btn-danger { background: #dc3545; color: white; font-weight: 500; flex: 1; justify-content: center; }
  .btn-danger:disabled { opacity: 0.4; }
  .btn-refresh { background: #f5f5f5; color: #1d1d1f; border: 1px solid #d1d1d6; }

  .modal-overlay {
    position: fixed; inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex; align-items: center; justify-content: center;
    z-index: 1000;
  }
  .modal {
    background: white;
    border-radius: 8px;
    padding: 24px;
    width: 90%;
    max-width: 440px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }
  .modal-danger { border-top: 4px solid #dc3545; }
  .modal h2 { font-size: 18px; font-weight: 600; margin: 0 0 20px 0; color: #1d1d1f; }
  .modal-danger h2 { color: #dc3545; }
  .danger-desc { font-size: 13px; color: #6e6e73; margin: 0 0 16px 0; line-height: 1.6; }
  .danger-desc strong { color: #1d1d1f; font-weight: 600; }

  .form-group { margin-bottom: 16px; }
  .form-group label { display: block; font-size: 13px; font-weight: 500; margin-bottom: 6px; color: #1d1d1f; }
  .form-group select, .form-group input {
    width: 100%; padding: 8px 12px;
    border: 1px solid #d1d1d6; border-radius: 4px;
    font-size: 13px; background: white;
    font-family: inherit;
  }
  .form-group select:focus, .form-group input:focus { outline: none; border-color: #007aff; }
  .confirm-input { font-weight: 500; }

  .info-note {
    display: flex; align-items: flex-start; gap: 8px;
    padding: 12px;
    background: #f5f5f5; border-radius: 4px;
    margin-bottom: 16px;
  }
  .info-note span { font-size: 12px; color: #6e6e73; line-height: 1.5; }

  .modal-actions { display: flex; gap: 12px; margin-top: 20px; }
  .modal-actions button { flex: 1; justify-content: center; }
</style>