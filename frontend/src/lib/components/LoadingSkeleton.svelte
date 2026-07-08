<script lang="ts">
  /** 页面加载骨架屏，count 为占位卡片数量 */
  export let count: number = 3;
  export let columns: string = 'repeat(auto-fill, minmax(280px, 1fr))';
</script>

<div class="skeleton-grid" style="grid-template-columns: {columns}">
  {#each Array(count) as _, i (i)}
    <div class="skeleton-card">
      <div class="skeleton-line title shimmer"></div>
      <div class="skeleton-line short shimmer"></div>
      <div class="skeleton-line shimmer"></div>
      <div class="skeleton-line medium shimmer"></div>
    </div>
  {/each}
</div>

<style>
  .skeleton-grid {
    display: grid;
    gap: 16px;
    padding: 8px 0;
  }

  .skeleton-card {
    background: #f5f5f7;
    border: 1px solid #e5e5ea;
    border-radius: 12px;
    padding: 20px;
    min-height: 120px;
  }

  .skeleton-line {
    height: 12px;
    border-radius: 6px;
    background: #e8e8ed;
    margin-bottom: 12px;
  }

  .skeleton-line.title {
    height: 18px;
    width: 45%;
    margin-bottom: 16px;
  }

  .skeleton-line.short {
    width: 30%;
  }

  .skeleton-line.medium {
    width: 70%;
    margin-bottom: 0;
  }

  .shimmer {
    position: relative;
    overflow: hidden;
  }

  .shimmer::after {
    content: '';
    position: absolute;
    inset: 0;
    transform: translateX(-100%);
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.65),
      transparent
    );
    animation: shimmer 1.4s ease-in-out infinite;
  }

  @keyframes shimmer {
    100% {
      transform: translateX(100%);
    }
  }
</style>
