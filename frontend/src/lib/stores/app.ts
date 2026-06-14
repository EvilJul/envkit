import { writable } from 'svelte/store';

export const currentPage = writable('languages');
export const loading = writable(false);
export const systemInfo = writable({
  os: '',
  architecture: '',
  distribution: ''
});
