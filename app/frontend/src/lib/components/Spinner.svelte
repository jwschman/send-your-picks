<script lang="ts">
  export let size: 'sm' | 'md' | 'lg' = 'md'
  export let color: string = 'var(--custom-color-brand)'

  const sizes = {
    sm: '20px',
    md: '40px',
    lg: '60px'
  }
</script>

<div class="spinner-container">
  <div
    class="spinner"
    style="width: {sizes[size]}; height: {sizes[size]}; border-color: {color}33; border-top-color: {color};"
  ></div>
</div>

<style>
  .spinner-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 20px;
  }

  .spinner {
    border: 3px solid;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>
