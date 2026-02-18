<script lang="ts">
  import type { Badge } from '$lib/types'

  export let badges: Badge[] = []
  export let compact: boolean = false
</script>

{#if badges.length > 0}
  <span class="badge-list" class:compact>
    {#each badges as badge}
      <span class="user-badge {badge.type}">{badge.label}</span>
    {/each}
  </span>
{/if}

<style>
  .badge-list {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .user-badge {
    display: inline-block;
    padding: 3px 10px;
    border-radius: var(--custom-border-radius);
    font-size: 0.7rem;
    font-weight: bold;
    text-transform: uppercase;
  }

  .compact .user-badge {
    padding: 2px 7px;
    font-size: 0.6rem;
  }

  .user-badge.previous_week_winner {
    background: #ca8a04;
    color: white;
  }

  .user-badge.previous_season_winner {
    background: #b45309;
    color: white;
  }

  .user-badge.previous_season_loser {
    background: #7f1d1d;
    color: white;
  }
</style>
