<script lang="ts">
  export let width: string = '100%'
  export let height: string = '20px'
  export let borderRadius: string = '4px'
  export let count: number = 1
  export let gap: string = '10px'
</script>

<div class="skeleton-wrapper" style="gap: {gap};">
  {#each Array(count) as _, i}
    <div
      class="skeleton"
      style="width: {width}; height: {height}; border-radius: {borderRadius};"
    ></div>
  {/each}
</div>

<style>
  .skeleton-wrapper {
    display: flex;
    flex-direction: column;
  }

  .skeleton {
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0.05) 25%,
      rgba(255, 255, 255, 0.1) 50%,
      rgba(255, 255, 255, 0.05) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }

  @keyframes shimmer {
    0% {
      background-position: -200% 0;
    }
    100% {
      background-position: 200% 0;
    }
  }
</style>
