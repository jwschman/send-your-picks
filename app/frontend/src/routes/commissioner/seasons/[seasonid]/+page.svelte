<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import { formatShortDate } from '$lib/utils/formatters'
  import type { Season, WeekStatus, Participant } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let season: Season | null = null
  let seasonId = ''

  // Participants (display only - management is on separate page)
  let participants: Participant[] = []

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    if (!seasonId) {
      error = 'No season ID provided'
      return
    }

    // Fetch season and participants in parallel
    fetchSeason()
    fetchParticipants()
  })

  async function advanceSeasonState() {
    if (!season) return

    try {
      loading = true

      await api.post(`/api/commissioner/seasons/${seasonId}/advance`, {}, session?.access_token)
      toasts.success('Season state advanced successfully!')
      await fetchSeason()
    } catch (err) {
      // Just show toast for action errors
      handleApiError(err, 'Failed to advance season state')
    } finally {
      loading = false
    }
  }

  async function fetchSeason() {
    try {
      loading = true
      error = ''

      const responseData = await api.get<{ season: Season }>(`/api/seasons/${seasonId}`, session?.access_token)
      season = responseData.season
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch season')
    } finally {
      loading = false
    }
  }

  // Fetch participants for this season
  async function fetchParticipants() {
    try {
      const res = await api.get<{ participants: Participant[] }>(
        `/api/seasons/${seasonId}/participants`,
        session?.access_token
      )
      participants = res.participants
    } catch (err) {
      // Don't set main error - just log it, participants section will show empty
      console.error('Failed to fetch participants:', err)
    }
  }

  function getWeekStatusLabel(status: WeekStatus): string {
    return status.toUpperCase()
  }
</script>

<svelte:head>
  <title>{season ? `${season.year} ${season.is_postseason ? 'Postseason' : 'Season'} (Commissioner)` : 'Season'}</title>
</svelte:head>

<div class="container">
  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {/if}

  {#if error}
    <div class="card error-card">
      <p class="error">{error}</p>
      <div class="actions">
        <a href="/commissioner/seasons" class="button">← Back to Seasons</a>
      </div>
    </div>
  {/if}

  {#if season && !loading}
    <div class="season-container">
      <div class="page-header-three-col">
        <div class="header-left">
          <a href="/commissioner/seasons" class="button">← Back</a>
        </div>
        <h1 class="header-title">{season.year} {season.is_postseason ? 'Postseason' : 'Season'}</h1>
        <div class="header-right">
          <span class="status-badge {season.is_active ? 'active' : 'inactive'}">
            {season.is_active ? 'Active' : 'Inactive'}
          </span>
        </div>
      </div>

      <div class="action-buttons">
        <button
          class="button"
          on:click={advanceSeasonState}
          disabled={loading}
        >
          Advance Season State
        </button>
        <a href="/commissioner/seasons/{seasonId}/manage" class="button primary">
          Manage Season
        </a>
      </div>

      <div class="season-info-section">
        <h2>Season Information</h2>
        <div class="card season-info">
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
        </div>
      </div>

      <!-- Participants Section (view-only, management is via Manage Season page) -->
      <div class="participants-section">
        <div class="section-header">
          <h2>Participants ({participants.length})</h2>
        </div>

        {#if participants.length === 0}
          <div class="card">
            <p class="text-sm opacity-half">
              No participants yet. Add users to allow them to submit picks for this season.
            </p>
          </div>
        {:else}
          <div class="participants-list card">
            {#each participants as participant}
              <div class="participant-item">
                <div class="participant-avatar">
                  <img src={participant.avatar_url} alt={participant.username || 'User'} />
                </div>
                <span class="participant-name">{participant.username || 'Anonymous'}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="weeks-section">
        <div class="weeks-header">
          <h2>Weeks</h2>
        </div>

        {#if !season.weeks || season.weeks.length === 0}
          <div class="card">
            <p class="text-sm opacity-half">
              No weeks have been created for this season yet.
            </p>
          </div>
        {:else}
          <div class="weeks-grid">
            {#each season.weeks as week}
              <div class="week-card card">
                <div class="week-header">
                  <h3>Week {week.number}</h3>
                  <span class="week-status {week.status}">
                    {getWeekStatusLabel(week.status)}
                  </span>
                </div>

                <div class="week-meta">
                  <div class="meta-item text-sm opacity-half">
                    Created: {formatShortDate(week.created_at)}
                  </div>
                  <div class="meta-item text-sm opacity-half">
                    Updated: {formatShortDate(week.updated_at)}
                  </div>
                </div>

                <div class="week-actions">
                  <a
                    href={`/commissioner/seasons/${seasonId}/weeks/${week.id}`}
                    class="button small"
                  >
                    View
                  </a>
                  {#if week.status === 'games_imported' || week.status === 'spreads_set'}
                    <a
                      href={`/commissioner/seasons/${seasonId}/weeks/${week.id}/spreads`}
                      class="button small primary"
                    >
                      Edit Spreads
                    </a>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .season-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 10px;
    margin-bottom: 20px;
  }

  .weeks-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
    flex-wrap: wrap;
    gap: 10px;
  }

  .weeks-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 15px;
  }

  /* Season Info Section */
  .season-info-section h2 {
    margin: 0 0 15px 0;
  }

  /* Participants Section */
  .participants-section {
    margin-top: 20px;
    margin-bottom: 30px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 15px;
  }

  .section-header h2 {
    margin: 0;
  }

  .participants-list {
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
    padding: 15px;
  }

  .participant-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .participant-avatar img {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
  }

  .participant-name {
    font-size: 0.9rem;
  }
</style>
