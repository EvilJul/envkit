<script lang="ts">
  import { onMount } from 'svelte';
  import { GetEnvVariables, SetEnvVariable, DeleteEnvVariable, GetShellConfig } from '../../wailsjs/go/main/App';

  interface EnvVariable {
    key: string;
    value: string;
    scope: string; // 'user' | 'system'
    source: string; // 来源文件，如 ~/.zshrc
  }

  let userVariables: EnvVariable[] = [];
  let systemVariables: EnvVariable[] = [];
  let pathEntries: string[] = [];
  let loading = false;
  let activeTab: 'user' | 'system' | 'path' = 'user';
  let showAddModal = false;
  let showEditModal = false;
  let editingVariable: EnvVariable | null = null;
  let newVariable = { key: '', value: '', scope: 'user' };
  let shellType = 'zsh';

  onMount(async () => {
    await loadVariables();
  });

  async function loadVariables() {
    loading = true;
    try {
      const result = await GetEnvVariables();
      userVariables = result.user || [];
      systemVariables = result.system || [];
      pathEntries = result.path || [];
      shellType = result.shell || 'zsh';
    } catch (err) {
      console.error('Failed to load variables:', err);
      alert(`加载失败: ${err}`);
    }
    loading = false;
  }

  function openAddModal(scope: string) {
    newVariable = { key: '', value: '', scope };
    showAddModal = true;
  }

  function openEditModal(variable: EnvVariable) {
    editingVariable = { ...variable };
    showEditModal = true;
  }

  async function addVariable() {
    if (!newVariable.key || !newVariable.value) {
      alert('请填写完整信息');
      return;
    }

    loading = true;
    showAddModal = false;
    try {
      await SetEnvVariable(newVariable.key, newVariable.value, newVariable.scope);
      await loadVariables();
      alert('添加成功！\n\n注意：需要重新打开终端或重启应用才能生效。');
    } catch (err) {
      alert(`添加失败: ${err}`);
    }
    loading = false;
  }

  async function updateVariable() {
    if (!editingVariable || !editingVariable.key || !editingVariable.value) {
      alert('请填写完整信息');
      return;
    }

    loading = true;
    showEditModal = false;
    try {
      await SetEnvVariable(editingVariable.key, editingVariable.value, editingVariable.scope);
      await loadVariables();
      alert('更新成功！\n\n注意：需要重新打开终端或重启应用才能生效。');
    } catch (err) {
      alert(`更新失败: ${err}`);
    }
    loading = false;
    editingVariable = null;
  }

  async function deleteVariable(variable: EnvVariable) {
    if (!confirm(`确定要删除环境变量 "${variable.key}" 吗？\n\n这将从 ${variable.source} 文件中删除该变量。`)) {
      return;
    }

    loading = true;
    try {
      await DeleteEnvVariable(variable.key, variable.scope);
      await loadVariables();
      alert('删除成功！\n\n注意：需要重新打开终端或重启应用才能生效。');
    } catch (err) {
      alert(`删除失败: ${err}`);
    }
    loading = false;
  }

  async function addToPath() {
    const path = prompt('输入要添加到 PATH 的目录路径：');
    if (!path) return;

    loading = true;
    try {
      const currentPath = pathEntries.join(':');
      const newPath = `${currentPath}:${path}`;
      await SetEnvVariable('PATH', newPath, 'user');
      await loadVariables();
      alert('添加成功！\n\n注意：需要重新打开终端才能生效。');
    } catch (err) {
      alert(`添加失败: ${err}`);
    }
    loading = false;
  }

  async function removeFromPath(index: number) {
    if (!confirm(`确定要从 PATH 中删除该路径吗？\n\n${pathEntries[index]}`)) {
      return;
    }

    loading = true;
    try {
      const newEntries = pathEntries.filter((_, i) => i !== index);
      const newPath = newEntries.join(':');
      await SetEnvVariable('PATH', newPath, 'user');
      await loadVariables();
      alert('删除成功！\n\n注意：需要重新打开终端才能生效。');
    } catch (err) {
      alert(`删除失败: ${err}`);
    }
    loading = false;
  }
</script>

<div class="page">
  <div class="header">
    <h1>Environment Variables</h1>
    <div class="shell-info">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="4 17 10 11 4 5"/>
        <line x1="12" y1="19" x2="20" y2="19"/>
      </svg>
      当前 Shell: {shellType}
    </div>
  </div>

  <div class="tabs">
    <button
      class="tab"
      class:active={activeTab === 'user'}
      on:click={() => activeTab = 'user'}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
        <circle cx="12" cy="7" r="4"/>
      </svg>
      用户变量
    </button>
    <button
      class="tab"
      class:active={activeTab === 'system'}
      on:click={() => activeTab = 'system'}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="2" y="3" width="20" height="14" rx="2"/>
        <line x1="8" y1="21" x2="16" y2="21"/>
        <line x1="12" y1="17" x2="12" y2="21"/>
      </svg>
      系统变量
    </button>
    <button
      class="tab"
      class:active={activeTab === 'path'}
      on:click={() => activeTab = 'path'}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/>
        <polyline points="13 2 13 9 20 9"/>
      </svg>
      PATH 管理
    </button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'user'}
      <div class="section-header">
        <p class="section-description">当前用户的环境变量（修改 ~/.zshrc 或 ~/.bash_profile）</p>
        <button class="btn-primary" on:click={() => openAddModal('user')} disabled={loading}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新增变量
        </button>
      </div>

      <div class="variables-table">
        <div class="table-header">
          <div class="col-key">变量名</div>
          <div class="col-value">值</div>
          <div class="col-source">来源</div>
          <div class="col-actions">操作</div>
        </div>

        {#if userVariables.length === 0}
          <div class="empty-state">
            <p>没有找到用户环境变量</p>
            <button class="btn-link" on:click={() => openAddModal('user')}>添加第一个变量</button>
          </div>
        {:else}
          {#each userVariables as variable}
            <div class="table-row">
              <div class="col-key">{variable.key}</div>
              <div class="col-value" title={variable.value}>
                {variable.value.length > 60 ? variable.value.substring(0, 60) + '...' : variable.value}
              </div>
              <div class="col-source">{variable.source}</div>
              <div class="col-actions">
                <button class="btn-icon" on:click={() => openEditModal(variable)} title="编辑">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                </button>
                <button class="btn-icon btn-danger" on:click={() => deleteVariable(variable)} title="删除">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>

    {:else if activeTab === 'system'}
      <div class="section-header">
        <p class="section-description">系统级环境变量（修改 /etc/paths 或 /etc/profile，需要 sudo 权限）</p>
        <button class="btn-primary" on:click={() => openAddModal('system')} disabled={loading}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新增变量
        </button>
      </div>

      <div class="variables-table">
        <div class="table-header">
          <div class="col-key">变量名</div>
          <div class="col-value">值</div>
          <div class="col-source">来源</div>
          <div class="col-actions">操作</div>
        </div>

        {#if systemVariables.length === 0}
          <div class="empty-state">
            <p>没有找到系统环境变量</p>
            <p class="hint">或没有权限读取系统变量</p>
          </div>
        {:else}
          {#each systemVariables as variable}
            <div class="table-row">
              <div class="col-key">{variable.key}</div>
              <div class="col-value" title={variable.value}>
                {variable.value.length > 60 ? variable.value.substring(0, 60) + '...' : variable.value}
              </div>
              <div class="col-source">{variable.source}</div>
              <div class="col-actions">
                <button class="btn-icon" on:click={() => openEditModal(variable)} title="编辑">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                </button>
                <button class="btn-icon btn-danger" on:click={() => deleteVariable(variable)} title="删除">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>

    {:else}
      <div class="section-header">
        <p class="section-description">管理 PATH 环境变量中的目录列表</p>
        <button class="btn-primary" on:click={addToPath} disabled={loading}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          添加路径
        </button>
      </div>

      <div class="path-list">
        {#if pathEntries.length === 0}
          <div class="empty-state">
            <p>PATH 为空</p>
          </div>
        {:else}
          {#each pathEntries as entry, index}
            <div class="path-item">
              <div class="path-index">{index + 1}</div>
              <div class="path-value">{entry}</div>
              <button class="btn-icon btn-danger" on:click={() => removeFromPath(index)} title="删除">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
          {/each}
        {/if}
      </div>

      <div class="info-box">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#007aff" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <div>
          <p><strong>PATH 变量说明：</strong></p>
          <p>PATH 决定了系统在哪些目录中查找可执行命令。顺序越靠前的目录优先级越高。</p>
        </div>
      </div>
    {/if}
  </div>
</div>

{#if showAddModal}
  <div class="modal-overlay" on:click={() => showAddModal = false}>
    <div class="modal" on:click|stopPropagation>
      <h2>新增环境变量</h2>

      <div class="form-group">
        <label>变量名</label>
        <input type="text" bind:value={newVariable.key} placeholder="例如: JAVA_HOME" />
      </div>

      <div class="form-group">
        <label>值</label>
        <input type="text" bind:value={newVariable.value} placeholder="例如: /usr/local/java" />
      </div>

      <div class="info-note">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#6e6e73" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <span>修改将写入 ~/.{shellType}rc 文件，需要重新打开终端才能生效</span>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" on:click={() => showAddModal = false}>
          取消
        </button>
        <button class="btn-primary" on:click={addVariable}>
          添加
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showEditModal && editingVariable}
  <div class="modal-overlay" on:click={() => showEditModal = false}>
    <div class="modal" on:click|stopPropagation>
      <h2>编辑环境变量</h2>

      <div class="form-group">
        <label>变量名</label>
        <input type="text" bind:value={editingVariable.key} disabled />
      </div>

      <div class="form-group">
        <label>值</label>
        <input type="text" bind:value={editingVariable.value} />
      </div>

      <div class="info-note">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#6e6e73" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <span>修改将更新 {editingVariable.source} 文件，需要重新打开终端才能生效</span>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" on:click={() => showEditModal = false}>
          取消
        </button>
        <button class="btn-primary" on:click={updateVariable}>
          保存
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

  .shell-info {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    font-size: 12px;
    color: #6e6e73;
  }

  .tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid #e5e5e5;
    margin-bottom: 24px;
  }

  .tab {
    padding: 10px 20px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    font-size: 13px;
    font-weight: 500;
    color: #6e6e73;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .tab.active {
    color: #007aff;
    border-bottom-color: #007aff;
  }

  .tab:hover {
    color: #1d1d1f;
  }

  .tab-content {
    animation: fadeIn 0.2s;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .section-description {
    font-size: 13px;
    color: #6e6e73;
    margin: 0;
  }

  .variables-table {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 16px;
  }

  .table-header {
    display: grid;
    grid-template-columns: 200px 1fr 180px 100px;
    gap: 12px;
    padding: 8px 12px;
    font-size: 11px;
    font-weight: 600;
    color: #6e6e73;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 8px;
  }

  .table-row {
    display: grid;
    grid-template-columns: 200px 1fr 180px 100px;
    gap: 12px;
    padding: 12px;
    background: white;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    margin-bottom: 8px;
    align-items: center;
    font-size: 13px;
  }

  .col-key {
    font-weight: 500;
    color: #1d1d1f;
  }

  .col-value {
    color: #1d1d1f;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .col-source {
    font-size: 12px;
    color: #6e6e73;
  }

  .col-actions {
    display: flex;
    justify-content: flex-end;
    gap: 4px;
  }

  .empty-state {
    text-align: center;
    padding: 60px 20px;
    color: #86868b;
  }

  .empty-state p {
    margin: 0 0 8px 0;
    font-size: 14px;
  }

  .empty-state .hint {
    font-size: 12px;
    margin-top: 4px;
  }

  .path-list {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .path-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: white;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    margin-bottom: 8px;
  }

  .path-index {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #007aff;
    color: white;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
  }

  .path-value {
    flex: 1;
    font-size: 13px;
    color: #1d1d1f;
    font-family: 'Monaco', 'Menlo', monospace;
  }

  .info-box {
    display: flex;
    gap: 12px;
    padding: 16px;
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
  }

  .info-box p {
    margin: 0 0 4px 0;
    font-size: 13px;
    color: #1d1d1f;
  }

  .info-box strong {
    font-weight: 600;
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

  .btn-icon {
    padding: 6px;
    background: none;
    color: #1d1d1f;
  }

  .btn-icon:hover {
    background: #f5f5f5;
  }

  .btn-icon.btn-danger {
    color: #dc3545;
  }

  .btn-icon.btn-danger:hover {
    background: #dc354520;
  }

  .btn-link {
    background: none;
    color: #007aff;
    padding: 0;
    text-decoration: underline;
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
    max-width: 500px;
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

  .form-group input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #d1d1d6;
    border-radius: 4px;
    font-size: 13px;
    background: white;
  }

  .form-group input:focus {
    outline: none;
    border-color: #007aff;
    box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
  }

  .form-group input:disabled {
    background: #f5f5f5;
    color: #86868b;
    cursor: not-allowed;
  }

  .info-note {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 12px;
    background: #f5f5f5;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .info-note span {
    font-size: 12px;
    color: #6e6e73;
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    margin-top: 20px;
  }

  .modal-actions button {
    flex: 1;
    justify-content: center;
  }
</style>
