<script lang="ts">
  import { toasts } from '$lib/stores/toast'
  import { fade, fly } from 'svelte/transition'
</script>

<div class="toast-container">
  {#each $toasts as toast (toast.id)}
    <div
      class="toast toast-{toast.type}"
      in:fly={{ y: -20, duration: 300 }}
      out:fade={{ duration: 200 }}
    >
      <div class="toast-content">
        <span class="toast-icon">
          {#if toast.type === 'success'}
            ✓
          {:else if toast.type === 'error'}
            ✗
          {:else if toast.type === 'warning'}
            ⚠
          {:else}
            ℹ
          {/if}
        </span>
        <span class="toast-message">{toast.message}</span>
      </div>
      <button class="toast-close" on:click={() => toasts.remove(toast.id)}>
        ✕
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    top: 70px;
    right: 20px;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 400px;
  }

  .toast {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .toast-content {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
  }

  .toast-icon {
    font-size: 1.1rem;
    font-weight: bold;
    flex-shrink: 0;
  }

  .toast-message {
    font-size: 0.9rem;
    line-height: 1.4;
  }

  .toast-close {
    background: none;
    border: none;
    color: inherit;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0 4px;
    margin-left: 8px;
    opacity: 0.6;
    transition: opacity 0.2s;
  }

  .toast-close:hover {
    opacity: 1;
  }

  .toast-success {
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.3);
  }

  .toast-success .toast-icon {
    color: #10b981;
  }

  .toast-error {
    background: rgba(239, 68, 68, 0.15);
    border-color: rgba(239, 68, 68, 0.3);
  }

  .toast-error .toast-icon {
    color: #ef4444;
  }

  .toast-warning {
    background: rgba(245, 158, 11, 0.15);
    border-color: rgba(245, 158, 11, 0.3);
  }

  .toast-warning .toast-icon {
    color: #f59e0b;
  }

  .toast-info {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.3);
  }

  .toast-info .toast-icon {
    color: #3b82f6;
  }

  @media (max-width: 640px) {
    .toast-container {
      left: 10px;
      right: 10px;
      max-width: none;
    }
  }
</style>
