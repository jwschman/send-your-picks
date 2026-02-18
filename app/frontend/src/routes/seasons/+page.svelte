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

  let loading = false
  let error = ''

  let seasons: Season[] = []
  let activeSeason: Season | null = null

  onMount(() => {
    fetchSeasons()
  })

  async function fetchSeasons() {
    try {
      loading = true
      error = ''

      const responseData = await api.get<{ seasons: Season[] | null }>('/api/seasons', session?.access_token)
      seasons = responseData.seasons || []
      activeSeason = seasons.find(s => s.is_active) || null
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch seasons')
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Seasons</title>
</svelte:head>

<div class="seasons-container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/dashboard" class="button">← Back</a>
    </div>
    <h1 class="header-title">Seasons</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {/if}

  {#if error}
    <div class="card error-card">
      <p class="error">{error}</p>
    </div>
  {/if}

  {#if !loading && seasons.length === 0}
    <div class="card">
      <EmptyState
        icon="🏈"
        title="No Seasons Yet"
        message="No seasons have been created. Check back later or contact your commissioner."
      />
    </div>
  {/if}

  {#if seasons.length > 0}
    <!-- Active Season Highlight -->
    {#if activeSeason}
      <div class="card active-season-card">
        <div class="active-season-header">
          <div>
            <h2>Active Season</h2>
            <p class="season-year-large">{activeSeason.year}{activeSeason.is_postseason ? ' Postseason' : ''}</p>
          </div>
          <a href="/seasons/{activeSeason.id}" class="button primary">
            View Season
          </a>
        </div>
      </div>
    {/if}

    <!-- All Seasons -->
    <div class="seasons-section">
      <h2>All Seasons</h2>
      <div class="seasons-grid">
        {#each seasons as season}
          <a href="/seasons/{season.id}" class="season-card card">
            <div class="season-card-header">
              <span class="season-year">{season.year}</span>
              <div class="badge-group">
                {#if season.is_postseason}
                  <span class="status-badge postseason">Postseason</span>
                {:else}
                  <span class="status-badge regular">Regular</span>
                {/if}
                {#if season.is_active}
                  <span class="status-badge active">Active</span>
                {:else}
                  <span class="status-badge inactive">Archived</span>
                {/if}
              </div>
            </div>
          </a>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .seasons-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .active-season-card {
    padding: 25px;
    margin-bottom: 30px;
    border: 2px solid var(--custom-color-brand);
  }

  .active-season-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    flex-wrap: wrap;
  }

  .active-season-header h2 {
    margin: 0 0 8px 0;
    font-size: 1.1rem;
    font-weight: 500;
  }

  .season-year-large {
    font-size: 2.5rem;
    font-weight: 700;
    color: var(--custom-color-brand);
    margin: 0;
  }

  .seasons-section {
    margin-bottom: 30px;
  }

  .seasons-section h2 {
    font-size: 1.3rem;
    margin-bottom: 20px;
  }

  .seasons-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 15px;
  }

  .season-card {
    text-decoration: none;
    transition: all 0.2s ease;
    cursor: pointer;
    padding: 20px;
  }

  .season-card:hover {
    transform: translateY(-2px);
    border-color: var(--custom-color-brand);
  }

  .season-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 15px;
  }

  .season-year {
    font-size: 1.8rem;
    font-weight: 600;
    color: var(--custom-color);
  }

  .status-badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-badge.active {
    background: var(--custom-color-brand);
    color: #0a0a0a;
  }

  .status-badge.inactive {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.6);
  }

  .status-badge.postseason {
    background: #7c3aed;
    color: white;
  }

  .status-badge.regular {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.6);
  }

  .badge-group {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  @media only screen and (max-width: 45em) {
    .seasons-grid {
      grid-template-columns: 1fr;
    }

    .active-season-header {
      flex-direction: column;
      text-align: center;
    }

    .season-year-large {
      font-size: 2rem;
    }
  }
</style>