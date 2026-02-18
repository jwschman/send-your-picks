<script lang="ts">
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import EmptyState from '$lib/components/EmptyState.svelte'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import type { Season } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let seasons: Season[] = []
  let loading = true
  let error = ''

  onMount(() => {
    fetchSeasons()
  })

  async function fetchSeasons() {
    try {
      loading = true
      error = ''

      const result = await api.get<{ seasons: Season[] | null }>('/api/seasons', session?.access_token)
      seasons = result.seasons || []
    } catch (err) {
      error = handleApiError(err, 'Failed to load seasons')
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Commissioner Seasons</title>
</svelte:head>

<div class="page">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/commissioner" class="button">← Back</a>
    </div>
    <h1 class="header-title">Seasons</h1>
    <div class="header-right">
      <a href="/commissioner/seasons/new" class="button primary">+ Create Season</a>
    </div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {:else if error}
    <div class="card error-card">
      <p class="error">{error}</p>
    </div>
  {:else if seasons.length === 0}
    <div class="card">
      <EmptyState
        icon="🏈"
        title="No Seasons Yet"
        message="Get started by creating your first season."
        actionText="Create Season"
        actionHref="/commissioner/seasons/new"
      />
    </div>
  {:else}
    <div class="card">
      <table class="seasons-table">
        <thead>
          <tr>
            <th>Year</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each seasons as season}
            <tr>
              <td>{season.year}{season.is_postseason ? ' Postseason' : ''}</td>
              <td>
                <span class="status-badge {season.is_active ? 'active' : 'inactive'}">
                  {season.is_active ? 'Active' : 'Inactive'}
                </span>
              </td>
              <td>
                <a
                  href={`/commissioner/seasons/${season.id}`}
                  class="button small"
                >
                  View Season
                </a>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .page {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .seasons-table {
    width: 100%;
    border-collapse: collapse;
  }

  .seasons-table th,
  .seasons-table td {
    border: 1px solid var(--custom-border-color, #333);
    padding: 10px;
    text-align: left;
  }

  .seasons-table th {
    background: var(--custom-panel-color);
  }

  .seasons-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .button.small {
    padding: 6px 10px;
    font-size: 0.85em;
  }
</style>
