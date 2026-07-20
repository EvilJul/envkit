<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { GetDatabases, StartDatabase, StopDatabase, RemoveDatabase } from '../../wailsjs/go/main/App';
  import ProgressBar from '../lib/components/ProgressBar.svelte';
  import LoadingSkeleton from '../lib/components/LoadingSkeleton.svelte';
  import { subscribeTaskProgress } from '../lib/taskProgress';

  interface Database {
    name: string;
    type: string;
    version: string;
    status: string;
  }

  interface ProgressInfo {
    percent: number;
    stage: string;
    message: string;
    status: 'running' | 'success' | 'error';
  }

  let databases: Database[] = [];
  let loading = false;
  let showStartModal = false;
  let selectedDb = { name: '', version: '' };
  let progressMap: Record<string, ProgressInfo> = {};
  const progressTimers: Record<string, ReturnType<typeof setTimeout>> = {};
  let unsubProgress: (() => void) | null = null;

  const SUCCESS_HOLD_MS = 1200;
  const ERROR_HOLD_MS = 3000;

  // 仅列出后端 StartDatabase 实际支持的数据库，避免假按钮
  const availableDatabases = [
    { name: 'postgres', displayName: 'PostgreSQL', defaultVersion: '16', description: '强大的开源关系型数据库' },
    { name: 'redis', displayName: 'Redis', defaultVersion: '7', description: '高性能内存键值数据库' },
    { name: 'mysql', displayName: 'MySQL', defaultVersion: '8.0', description: '最流行的开源关系型数据库' },
    { name: 'mongodb', displayName: 'MongoDB', defaultVersion: '6.0', description: '文档型 NoSQL 数据库' }
  ];

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
    await loadDatabases();
  });

  onDestroy(() => {
    unsubProgress?.();
  });

  async function loadDatabases() {
    loading = true;
    try {
      databases = await GetDatabases();
    } catch (err) {
      console.error('Failed to load databases:', err);
    }
    loading = false;
  }

  function openStartModal(db: any) {
    selectedDb = { name: db.name, version: db.defaultVersion };
    showStartModal = true;
  }

  function clearProgressLater(name: string, delay: number) {
    if (progressTimers[name]) {
      clearTimeout(progressTimers[name]);
    }
    progressTimers[name] = setTimeout(() => {
      const next = { ...progressMap };
      delete next[name];
      progressMap = next;
      delete progressTimers[name];
    }, delay);
  }

  function setProgress(name: string, info: ProgressInfo) {
    progressMap = { ...progressMap, [name]: info };
  }

  async function startDatabase() {
    if (!selectedDb.name || !selectedDb.version) return;

    const name = selectedDb.name;
    showStartModal = false;
    setProgress(name, { percent: 0, stage: 'preparing', message: '准备启动...', status: 'running' });
    try {
      await StartDatabase(name, selectedDb.version);
      await loadDatabases();
      setProgress(name, { percent: 100, stage: 'done', message: '启动完成', status: 'success' });
      clearProgressLater(name, SUCCESS_HOLD_MS);
    } catch (err) {
      setProgress(name, { percent: 100, stage: 'error', message: String(err), status: 'error' });
      clearProgressLater(name, ERROR_HOLD_MS);
    }
  }

  async function stopDatabase(containerName: string) {
    if (!confirm(`确定要停止容器 ${containerName} 吗？`)) return;

    setProgress(containerName, { percent: 0, stage: 'notifying', message: '准备停止...', status: 'running' });
    try {
      await StopDatabase(containerName);
      await loadDatabases();
      setProgress(containerName, { percent: 100, stage: 'done', message: '已停止', status: 'success' });
      clearProgressLater(containerName, SUCCESS_HOLD_MS);
    } catch (err) {
      setProgress(containerName, { percent: 100, stage: 'error', message: String(err), status: 'error' });
      clearProgressLater(containerName, ERROR_HOLD_MS);
    }
  }

  async function removeDatabase(containerName: string) {
    if (!confirm(`确定要删除容器 ${containerName} 吗？\n\n数据将会丢失！`)) return;

    setProgress(containerName, { percent: 0, stage: 'stopping', message: '准备移除...', status: 'running' });
    try {
      await RemoveDatabase(containerName, true);
      await loadDatabases();
      setProgress(containerName, { percent: 100, stage: 'done', message: '移除完成', status: 'success' });
      clearProgressLater(containerName, SUCCESS_HOLD_MS);
    } catch (err) {
      setProgress(containerName, { percent: 100, stage: 'error', message: String(err), status: 'error' });
      clearProgressLater(containerName, ERROR_HOLD_MS);
    }
  }
</script>

<div class="page">
  <div class="header">
    <h1>Database Containers</h1>
    <button class="btn-refresh" on:click={loadDatabases} disabled={loading}>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
      </svg>
      {loading ? '加载中...' : '刷新'}
    </button>
  </div>

  {#if loading && databases.length === 0}
    <LoadingSkeleton count={4} columns="repeat(auto-fill, minmax(280px, 1fr))" />
  {:else if databases.length === 0}
    <div class="empty-state">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#86868b" stroke-width="2">
        <ellipse cx="12" cy="5" rx="9" ry="3"/>
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
      </svg>
      <h2>没有运行中的容器</h2>
      <p>启动一个数据库容器来开始使用</p>

      <div class="db-options">
        {#each availableDatabases as db}
          <button class="btn-start-db" on:click={() => openStartModal(db)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            启动 {db.displayName}
          </button>
        {/each}
      </div>
    </div>
  {:else}
    <div class="database-table">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>版本</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {#each databases as db}
            <tr>
              <td class="name">{db.name}</td>
              <td>{db.type}</td>
              <td>{db.version}</td>
              <td>
                <span class="status" class:running={db.status === 'running'}>
                  {#if db.status === 'running'}
                    ● 运行中
                  {:else}
                    ○ 已停止
                  {/if}
                </span>
              </td>
              <td>
                <div class="actions">
                  {#if db.status === 'running'}
                    <button class="btn-small btn-secondary" on:click={() => stopDatabase(db.name)} disabled={!!progressMap[db.name]}>
                      停止
                    </button>
                  {/if}
                  <button class="btn-small btn-danger" on:click={() => removeDatabase(db.name)} disabled={!!progressMap[db.name]}>
                    删除
                  </button>
                </div>
                {#if progressMap[db.name]}
                  <div class="row-progress">
                    <ProgressBar
                      percent={progressMap[db.name].percent}
                      stage={progressMap[db.name].stage}
                      message={progressMap[db.name].message}
                      status={progressMap[db.name].status}
                    />
                  </div>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <div class="add-container">
      <button class="btn-primary" on:click={() => openStartModal(availableDatabases[0])} disabled={loading}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        启动新容器
      </button>
    </div>

    {#if Object.keys(progressMap).filter(n => !databases.some(d => d.name === n)).length > 0}
      <div class="operation-status">
        <div class="operation-status-title">操作状态</div>
        <div class="operation-status-list">
          {#each Object.keys(progressMap).filter(n => !databases.some(d => d.name === n)) as name (name)}
            <ProgressBar
              percent={progressMap[name].percent}
              stage={progressMap[name].stage}
              message={progressMap[name].message}
              status={progressMap[name].status}
            />
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

{#if showStartModal}
  <div class="modal-overlay" on:click={() => showStartModal = false}>
    <div class="modal" on:click|stopPropagation>
      <h2>启动数据库容器</h2>

      <div class="form-group">
        <label>数据库类型</label>
        <select bind:value={selectedDb.name}>
          {#each availableDatabases as db}
            <option value={db.name}>{db.displayName}</option>
          {/each}
        </select>
      </div>

      <div class="form-group">
        <label>版本</label>
        <input type="text" bind:value={selectedDb.version} placeholder="例如: 16" />
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" on:click={() => showStartModal = false}>
          取消
        </button>
        <button class="btn-primary" on:click={startDatabase}>
          启动
        </button>
      </div>
    </div>
  </div>
{/if}

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

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
    color: #86868b;
  }

  .empty-state h2 {
    font-size: 18px;
    font-weight: 500;
    margin: 16px 0 8px 0;
    color: #1d1d1f;
  }

  .empty-state p {
    margin: 0 0 24px 0;
    font-size: 14px;
  }

  .db-options {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
    width: 100%;
    max-width: 400px;
  }

  .btn-start-db {
    padding: 12px 16px;
    background: #007aff;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: opacity 0.1s;
  }

  .btn-start-db:hover {
    opacity: 0.9;
  }

  .database-table {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background: #ebebeb;
  }

  th {
    text-align: left;
    padding: 12px 16px;
    font-size: 11px;
    font-weight: 600;
    color: #6e6e73;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  td {
    padding: 12px 16px;
    font-size: 13px;
    border-top: 1px solid #e5e5e5;
  }

  td.name {
    font-weight: 500;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #86868b;
  }

  .status.running {
    color: #28a745;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .add-container {
    margin-top: 16px;
  }

  .row-progress {
    margin-top: 8px;
    min-width: 220px;
  }

  .operation-status {
    margin-top: 16px;
    padding: 12px 16px;
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
  }

  .operation-status-title {
    font-size: 11px;
    font-weight: 600;
    color: #6e6e73;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
  }

  .operation-status-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
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

  .btn-danger {
    background: #dc3545;
    color: white;
  }

  .btn-small {
    padding: 4px 12px;
    font-size: 12px;
  }

  .btn-refresh {
    background: #f5f5f5;
    color: #1d1d1f;
    border: 1px solid #d1d1d6;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    border-radius: 8px;
    padding: 24px;
    width: 90%;
    max-width: 400px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }

  .modal h2 {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 20px 0;
    color: #1d1d1f;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    font-size: 13px;
    font-weight: 500;
    margin-bottom: 6px;
    color: #1d1d1f;
  }

  .form-group input,
  .form-group select {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    font-size: 13px;
    background: white;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #007aff;
    box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    margin-top: 24px;
  }

  .modal-actions button {
    flex: 1;
    justify-content: center;
  }
</style>
