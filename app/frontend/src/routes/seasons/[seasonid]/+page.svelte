<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { logData } from '$lib/utils/logger'
  import { formatShortDate } from '$lib/utils/formatters'
  import type { Season, Standing } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  type MyStandings = Standing & {
    rank: number | null
  }

  let season: Season | null = null
  let myStandings: MyStandings | null = null
  let seasonId: string = ''

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    if (!seasonId) {
      error = 'No season ID provided'
      return
    }

    fetchSeason()
    fetchMyStandings()
  })

  async function fetchSeason() {
    try {
      loading = true
      error = ''

      const responseData = await api.get<{ season: Season }>(`/api/seasons/${seasonId}`, session?.access_token)
      season = responseData.season
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        error = 'Season not found'
      } else {
        error = handleApiError(err, 'Failed to fetch season')
      }
    } finally {
      loading = false
    }
  }

  async function fetchMyStandings() {
    try {
      const responseData = await api.get<{ my_standings: MyStandings }>(
        `/api/seasons/${seasonId}/standings/me`,
        session?.access_token
      )
      myStandings = responseData.my_standings
      logData('STANDINGS_LOAD_SUCCESS', { seasonId })
    } catch (err) {
      // Silently fail for standings - not critical, just means no data yet
      logData('STANDINGS_UNAVAILABLE', { seasonId, reason: 'not_yet_available' })
    }
  }

  function getStatusBadgeClass(status: string): string {
    switch (status) {
      case 'draft': return 'draft'
      case 'active': return 'active'
      case 'played': return 'open'
      case 'final': return 'closed'
      default: return 'draft'
    }
  }

  function getStatusLabel(status: string): string {
    switch (status) {
      case 'draft': return 'Draft'
      case 'active': return 'Active'
      case 'played': return 'Played'
      case 'final': return 'Final'
      default: return status
    }
  }
</script>

<svelte:head>
  <title>{season ? `${season.year} ${season.is_postseason ? 'Postseason' : 'Season'}` : 'Season'}</title>
</svelte:head>

<div class="container">
  <div class="row">
    <div class="col-12">
      {#if loading}
        <div class="card">
          <Spinner />
        </div>
      {/if}

      {#if error}
        <div class="card error-card">
          <p class="error">{error}</p>
          <div style="margin-top: 20px;">
            <a href="/seasons" class="button">← Back to Seasons</a>
          </div>
        </div>
      {/if}

      {#if season && !loading}
        <div class="page-container">
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons" class="button">← Back</a>
            </div>
            <h1 class="header-title">{season.year} {season.is_postseason ? 'Postseason' : 'Season'}</h1>
            <div class="header-right">
              {#if season.is_active}
                <span class="status-badge active">Active Season</span>
              {:else}
                <span class="status-badge inactive">Inactive</span>
              {/if}
            </div>
          </div>

          <div class="card info-card">
            <h2>Season Information</h2>
            <div class="info-row">
              <span class="info-label">Year:</span>
              <span class="info-value">{season.year}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Status:</span>
              <span class="info-value">{season.is_active ? 'Active' : 'Inactive'}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Total Weeks:</span>
              <span class="info-value">{season.weeks?.length ?? 0}</span>
            </div>
            {#if myStandings}
              <div class="info-row">
                <span class="info-label">Your Points:</span>
                <span class="info-value">{myStandings.points}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Your Rank:</span>
                <span class="info-value">
                  {#if myStandings.rank !== null}
                    {myStandings.rank} of {myStandings.total_users}
                  {:else}
                    Not ranked yet
                  {/if}
                </span>
              </div>
            {/if}
          </div>

          {#if season.is_active}
            <div class="card actions-card">
              <h3 style="margin-top: 0; margin-bottom: 12px;">Quick Links</h3>
              <div class="button-group">
                <a href="/seasons/{seasonId}/standings" class="button primary">
                  View Current Standings
                </a>
                <a href="/seasons/{seasonId}/points" class="button">
                  Get Point Breakdown
                </a>
                <a href="/seasons/{seasonId}/winners" class="button">
                  Weekly Winners
                </a>
              </div>
            </div>
          {/if}

          <div class="weeks-section">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
              <h2 style="margin: 0;">Weeks</h2>
              <span class="text-sm opacity-half">Click a week to view details</span>
            </div>
            
            {#if !season.weeks || season.weeks.length === 0}
              <div class="card">
                <p class="text-sm opacity-half">No weeks have been created for this season yet.</p>
              </div>
            {:else}
              <div class="grid-3">
                {#each season.weeks as week}
                  <a href="/seasons/{seasonId}/weeks/{week.id}" class="week-card card">
                    <div class="week-header">
                      <h3>Week {week.number}</h3>
                      <span class="status-badge {getStatusBadgeClass(week.status)}">
                        {getStatusLabel(week.status)}
                      </span>
                    </div>
                    <div class="week-meta">
                      <div class="text-sm opacity-half">
                        Created: {formatShortDate(week.created_at)}
                      </div>
                      <div class="text-sm opacity-half">
                        Updated: {formatShortDate(week.updated_at)}
                      </div>
                    </div>
                  </a>
                {/each}
              </div>
            {/if}
          </div>

        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .weeks-section {
    margin-bottom: 30px;
  }

  .weeks-section h2 {
    font-size: 1.3rem;
  }

  .week-card {
    text-decoration: none;
    transition: all 0.2s ease;
    cursor: pointer;
    padding: 15px;
  }

  .week-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px 0 rgba(0, 0, 0, 0.9);
  }

  .week-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .week-header h3 {
    margin: 0;
    font-size: 1.1rem;
  }

  .week-meta {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .actions-card {
    padding: 20px;
    margin-bottom: 20px;
  }

  @media only screen and (max-width: 45em) {
    .button-group {
      flex-direction: column;
    }
  }
</style>