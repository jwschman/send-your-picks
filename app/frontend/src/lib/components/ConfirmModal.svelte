<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { fade, scale } from 'svelte/transition'

  export let open = false
  export let title = 'Confirm Action'
  export let message = 'Are you sure you want to proceed?'
  export let confirmText = 'Confirm'
  export let cancelText = 'Cancel'
  export let confirmVariant: 'primary' | 'danger' = 'danger'

  const dispatch = createEventDispatcher()

  function handleConfirm() {
    dispatch('confirm')
    open = false
  }

  function handleCancel() {
    dispatch('cancel')
    open = false
  }

  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      handleCancel()
    }
  }
</script>

{#if open}
  <div
    class="modal-backdrop"
    on:click={handleBackdropClick}
    on:keydown={(e) => e.key === 'Escape' && handleCancel()}
    transition:fade={{ duration: 200 }}
    role="button"
    tabindex="0"
  >
    <div class="modal" transition:scale={{ duration: 200, start: 0.95 }}>
      <div class="modal-header">
        <h2>{title}</h2>
      </div>

      <div class="modal-body">
        <p>{message}</p>
      </div>

      <div class="modal-footer">
        <button class="button" on:click={handleCancel}>
          {cancelText}
        </button>
        <button
          class="button {confirmVariant === 'danger' ? 'danger' : 'primary'}"
          on:click={handleConfirm}
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 10000;
    padding: 20px;
  }

  .modal {
    background: var(--custom-panel-color);
    border: 1px solid var(--custom-border-color);
    border-radius: var(--custom-border-radius);
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
    max-width: 500px;
    width: 100%;
  }

  .modal-header {
    padding: 20px 24px;
    border-bottom: var(--custom-border);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .modal-body {
    padding: 24px;
  }

  .modal-body p {
    margin: 0;
    line-height: 1.6;
    color: var(--custom-color-text);
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: var(--custom-border);
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .button.danger {
    background-color: #ef4444;
    color: white;
  }

  .button.danger:hover {
    background-color: #dc2626;
  }
</style>
