<script lang="ts">
  export let title: string
  export let message: string = ''
  export let icon: string = '📭'
  export let actionText: string = ''
  export let actionHref: string = ''
</script>

<div class="empty-state">
  <div class="empty-icon">{icon}</div>
  <h3 class="empty-title">{title}</h3>
  {#if message}
    <p class="empty-message">{message}</p>
  {/if}
  {#if actionText && actionHref}
    <a href={actionHref} class="button primary empty-action">
      {actionText}
    </a>
  {/if}
  <slot />
</div>

<style>
  .empty-state {
    padding: 60px 40px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .empty-icon {
    font-size: 4rem;
    opacity: 0.5;
    margin-bottom: 8px;
  }

  .empty-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
    color: var(--custom-color-text);
  }

  .empty-message {
    font-size: 0.95rem;
    color: var(--custom-color-secondary);
    max-width: 400px;
    margin: 0;
    line-height: 1.6;
  }

  .empty-action {
    margin-top: 8px;
  }

  @media (max-width: 640px) {
    .empty-state {
      padding: 40px 20px;
    }

    .empty-icon {
      font-size: 3rem;
    }
  }
</style>
