<script lang="ts">
  import { goto } from '$app/navigation'
  import { env } from '$env/dynamic/public'
</script>

<svelte:head>
  <title>Authentication Error - {env.PUBLIC_APP_NAME || 'Send Your Picks'}</title>
</svelte:head>

<div class="error-container">
  <div class="error-card card">
    <div class="error-icon">!</div>
    <h1>Authentication Error</h1>
    <p>There was a problem verifying your account. The link may have expired or already been used.</p>
    <button class="button primary" on:click={() => goto('/login')}>
      Back to Login
    </button>
  </div>
</div>

<style>
  .error-container {
    min-height: 100vh;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 80px 20px 20px;
  }

  .error-card {
    width: 100%;
    max-width: 400px;
    padding: 40px;
    text-align: center;
  }

  .error-icon {
    width: 60px;
    height: 60px;
    margin: 0 auto 20px;
    background: rgba(239, 68, 68, 0.2);
    border: 2px solid #ef4444;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    font-weight: bold;
    color: #ef4444;
  }

  h1 {
    margin: 0 0 16px 0;
    font-size: 1.4rem;
  }

  p {
    color: rgba(255, 255, 255, 0.7);
    margin: 0 0 24px 0;
    font-size: 0.95rem;
    line-height: 1.5;
  }
</style>
