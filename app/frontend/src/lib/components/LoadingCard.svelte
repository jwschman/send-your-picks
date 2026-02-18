<script lang="ts">
  import Skeleton from './Skeleton.svelte'

  export let lines: number = 3
</script>

<div class="card loading-card">
  <Skeleton height="24px" width="60%" borderRadius="4px" />
  <div style="margin-top: 15px;">
    <Skeleton count={lines} height="16px" gap="12px" />
  </div>
</div>

<style>
  .loading-card {
    padding: 20px;
  }
</style>
