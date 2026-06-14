<script lang="ts">
  import { currentPage } from '../stores/app';

  function navigate(page: string) {
    currentPage.set(page);
  }
</script>

<aside class="sidebar">
  <div class="logo">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
      <rect x="3" y="3" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="14" y="3" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="3" y="14" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="14" y="14" width="7" height="7" rx="1" fill="#007aff"/>
    </svg>
    <span>EnvKit</span>
  </div>

  <nav>
    <button
      class="nav-item"
      class:active={$currentPage === 'languages'}
      on:click={() => navigate('languages')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
      </svg>
      Languages
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'tools'}
      on:click={() => navigate('tools')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
      </svg>
      Tools
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'database'}
      on:click={() => navigate('database')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <ellipse cx="12" cy="5" rx="9" ry="3"/>
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
      </svg>
      Database
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'environment'}
      on:click={() => navigate('environment')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="2" y="3" width="20" height="5" rx="2"/>
        <rect x="2" y="10" width="20" height="5" rx="2"/>
        <rect x="2" y="17" width="20" height="5" rx="2"/>
      </svg>
      Environment
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'settings'}
      on:click={() => navigate('settings')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="3"/>
        <path d="M12 1v6m0 6v6m9-9h-6m-6 0H3"/>
      </svg>
      Settings
    </button>
  </nav>
</aside>

<style>
  .sidebar {
    width: 200px;
    background-color: #f5f5f7;
    border-right: 1px solid #d1d1d6;
    display: flex;
    flex-direction: column;
    padding: 16px 0;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px 16px 16px;
    font-weight: 600;
    font-size: 15px;
    color: #1d1d1f;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 0 8px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: none;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    color: #1d1d1f;
    font-size: 13px;
    text-align: left;
    transition: background-color 0.1s;
  }

  .nav-item:hover {
    background-color: #ebebeb;
  }

  .nav-item.active {
    background-color: #d1d1d6;
    font-weight: 500;
  }
</style>
