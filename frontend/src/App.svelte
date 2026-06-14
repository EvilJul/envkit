<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import Languages from './routes/Languages.svelte';
  import Tools from './routes/Tools.svelte';
  import Database from './routes/Database.svelte';
  import Environment from './routes/Environment.svelte';
  import Settings from './routes/Settings.svelte';
  import { currentPage } from './lib/stores/app';
  import { GetSystemInfo } from '../wailsjs/go/main/App';

  let page = 'languages';

  currentPage.subscribe(value => {
    page = value;
  });

  onMount(async () => {
    try {
      const sysInfo = await GetSystemInfo();
      console.log('System Info:', sysInfo);
    } catch (err) {
      console.error('Failed to get system info:', err);
    }
  });
</script>

<main>
  <div class="app-container">
    <Sidebar />
    <div class="content">
      {#if page === 'languages'}
        <Languages />
      {:else if page === 'tools'}
        <Tools />
      {:else if page === 'database'}
        <Database />
      {:else if page === 'environment'}
        <Environment />
      {:else if page === 'settings'}
        <Settings />
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', sans-serif;
    font-size: 13px;
    color: #1d1d1f;
    background-color: #ffffff;
  }

  :global(*) {
    box-sizing: border-box;
  }

  main {
    width: 100%;
    height: 100vh;
    overflow: hidden;
  }

  .app-container {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
    background-color: #ffffff;
  }

  .placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #86868b;
  }

  .placeholder h2 {
    font-size: 24px;
    font-weight: 600;
    margin: 0 0 8px 0;
    color: #1d1d1f;
  }

  .placeholder p {
    margin: 0;
    font-size: 14px;
  }
</style>
