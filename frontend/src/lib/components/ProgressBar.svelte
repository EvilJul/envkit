<script lang="ts">
  export let percent: number = 0;
  export let stage: string = '';
  export let message: string = '';
  export let status: 'running' | 'success' | 'error' = 'running';

  $: clamped = Math.max(0, Math.min(100, Math.round(percent)));
  $: fillColor =
    status === 'success' ? '#28a745' :
    status === 'error'   ? '#dc3545' :
                            '#007aff';
  $: statusIcon =
    status === 'success' ? '✓' :
    status === 'error'   ? '✕' :
                            '';

  // 阶段名 -> 中文显示
  const STAGE_LABELS: Record<string, string> = {
    downloading: '下载中…',
    extracting: '解压中…',
    installing: '安装中…',
    configuring: '配置中…',
    verifying: '校验中…',
    preparing: '准备中…',
    initializing: '初始化中…',
    starting: '启动中…',
    stopping: '停止中…',
    cleaning: '清理中…',
    removing: '移除中…',
    notifying: '通知中…',
    validating: '校验中…',
    writing: '写入中…',
    done: '完成',
    error: '错误'
  };

  function stageLabel(s: string): string {
    if (!s) return '';
    return STAGE_LABELS[s] || s;
  }
</script>

<div class="progress" class:success={status === 'success'} class:error={status === 'error'}>
  <div class="progress-header">
    <span class="stage">
      {#if stage === 'done' || status === 'success'}
        {stageLabel(stage) || '完成'} {statusIcon}
      {:else if stage === 'error' || status === 'error'}
        {stageLabel(stage) || '错误'} {statusIcon}
      {:else}
        {stageLabel(stage) || '处理中…'}
      {/if}
    </span>
    <span class="percent">{clamped}%</span>
  </div>

  <div class="track">
    <div
      class="fill"
      style="width: {clamped}%; background: {fillColor};"
    ></div>
  </div>

  {#if message}
    <div class="message">{message}</div>
  {/if}
</div>

<style>
  .progress {
    width: 100%;
    font-size: 12px;
    color: #1d1d1f;
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .stage {
    color: #1d1d1f;
    font-weight: 500;
    font-size: 12px;
  }

  .progress:not(.success):not(.error) .stage {
    animation: pulse 1.6s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.55; }
  }

  .percent {
    color: #6e6e73;
    font-variant-numeric: tabular-nums;
    font-size: 11px;
  }

  .track {
    position: relative;
    width: 100%;
    height: 8px;
    background: #e5e5e5;
    border-radius: 4px;
    overflow: hidden;
  }

  .fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.35s ease-out, background 0.35s ease-out;
  }

  .message {
    margin-top: 4px;
    font-size: 11px;
    color: #86868b;
    line-height: 1.4;
  }

  .progress.success .stage { color: #28a745; }
  .progress.error   .stage { color: #dc3545; }
  .progress.success .percent,
  .progress.error   .percent { color: inherit; }
</style>