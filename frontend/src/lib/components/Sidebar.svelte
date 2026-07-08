<script lang="ts">
  import { currentPage } from '../stores/app';
  import BrandIcon from './icons/BrandIcon.svelte';

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
      <BrandIcon name="languages" size={20} color="currentColor" />
      Languages
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'tools'}
      on:click={() => navigate('tools')}
    >
      <BrandIcon name="tools" size={20} color="currentColor" />
      Tools
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'database'}
      on:click={() => navigate('database')}
    >
      <BrandIcon name="database" size={20} color="currentColor" />
      Database
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'environment'}
      on:click={() => navigate('environment')}
    >
      <BrandIcon name="environment" size={20} color="currentColor" />
      Environment
    </button>

    <button
      class="nav-item"
      class:active={$currentPage === 'settings'}
      on:click={() => navigate('settings')}
    >
      <BrandIcon name="settings" size={20} color="currentColor" />
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
