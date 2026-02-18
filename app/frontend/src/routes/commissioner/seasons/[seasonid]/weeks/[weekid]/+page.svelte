<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import { logWeek } from '$lib/utils/logger'
  import { formatWeekdayDateTime } from '$lib/utils/formatters'
  import type { Week, Game } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let activating = false
  let error = ''
  let successMessage = ''

  let week: Week | null = null
  let seasonId = ''
  let weekId = ''

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing route parameters'
      return
    }

    fetchWeek()
  })

  async function fetchWeek() {
    try {
      loading = true
      error = ''

      const responseData = await api.get<{ week: Week }>(`/api/weeks/${weekId}`, session?.access_token)
      week = responseData.week
    } catch (err) {
      error = handleApiError(err, 'Failed to load week')
    } finally {
      loading = false
    }
  }

  async function activateWeek() {
    if (!week) return

    if (week.status !== 'spreads_set') {
      toasts.error('You must set spreads for all games before activating the week.')
      return
    }

    logWeek('ACTIVATION_ATTEMPT', {
      weekId,
      seasonId,
      status: week.status,
      gameCount: week.games.length
    })

    try {
      activating = true
      error = ''
      successMessage = ''

      await api.post(`/api/commissioner/weeks/${weekId}/activate`, {}, session?.access_token)

      logWeek('ACTIVATION_SUCCESS', {
        weekId,
        seasonId,
        newStatus: 'active'
      })

      await fetchWeek()
      toasts.success('Week activated successfully!')
      successMessage = 'Week activated successfully!'
      setTimeout(() => { successMessage = '' }, 3000)
    } catch (err) {
      logWeek('ACTIVATION_FAILURE', {
        weekId,
        seasonId,
        error: err instanceof Error ? err.message : 'Unknown error'
      })
      // Just show toast for action errors
      handleApiError(err, 'Activation failed')
    } finally {
      activating = false
    }
  }

  function goToSpreads() {
    goto(`/commissioner/seasons/${seasonId}/weeks/${weekId}/spreads`)
  }

  function getSpreadWinner(game: Game): 'home' | 'away' | 'push' | null {
    if (game.home_score === null || game.away_score === null || game.home_spread === null) {
      return null
    }

    // Calculate the adjusted home score (home score + spread)
    const adjustedHomeScore = game.home_score + game.home_spread
    
    if (adjustedHomeScore > game.away_score) {
      return 'home'
    } else if (adjustedHomeScore < game.away_score) {
      return 'away'
    } else {
      return 'push'
    }
  }

  function formatSpread(spread: number | null): string {
    if (spread === null) return 'N/A'
    return spread > 0 ? `+${spread}` : `${spread}`
  }

</script>

<svelte:head>
  <title>Week {week?.number ?? ''} (Commissioner)</title>
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
    </div>
  {/if}

  {#if successMessage}
    <div class="card success-card">
      <p class="success">{successMessage}</p>
    </div>
  {/if}

  {#if week && !loading}
    <div class="week-container">
      <div class="page-header-three-col">
        <div class="header-left">
          <a href="/commissioner/seasons/{seasonId}" class="button">← Back</a>
        </div>
        <h1 class="header-title">Week {week.number}</h1>
        <div class="header-right">
          <span class="week-status {week.status}">
            {week.status.toUpperCase()}
          </span>
        </div>
      </div>

      <div class="card">
        <h3 style="margin-top: 0; margin-bottom: 12px;">Week Actions</h3>
        <div style="display: flex; gap: 12px; flex-wrap: wrap;">
          {#if week.status === 'games_imported' || week.status === 'spreads_set'}
            <button
              class="button"
              on:click={goToSpreads}
            >
              Add / Edit Spreads
            </button>

            <button
              class="button primary"
              disabled={activating}
              on:click={activateWeek}
            >
              {activating ? 'Activating…' : 'Activate Week'}
            </button>
          {/if}

          <a
            href="/commissioner/seasons/{seasonId}/weeks/{weekId}/picks"
            class="button"
          >
            View User Picks
          </a>
        </div>
      </div>

      <div class="games-section">
        <h2>Games</h2>

        {#if week.games.length === 0}
          <div class="card">
            <p>No games imported.</p>
          </div>
        {:else}
          <div class="games-grid">
            {#each week.games as game}
              {@const spreadWinner = getSpreadWinner(game)}
              <div class="card game-card">
                <div class="teams">
                  <div class="team-row">
                    {#if game.away_team_logo_url}
                      <img src="/images/team-logos/{game.away_team_logo_url}" alt="{game.away_team_name}" class="team-logo" />
                    {/if}
                    <span class="team-abbr">{game.away_team_abbr}</span>
                    {#if game.away_score !== null}
                      <span class="score">{game.away_score}</span>
                    {/if}
                    {#if game.home_spread !== null}
                      <span class="spread">{formatSpread(-game.home_spread)}</span>
                    {/if}
                    {#if spreadWinner === 'away'}
                      <span class="spread-winner">✓</span>
                    {:else if spreadWinner === 'push'}
                      <span class="spread-push">-</span>
                    {/if}
                  </div>
                  <span class="at-symbol">@</span>
                  <div class="team-row">
                    {#if game.home_team_logo_url}
                      <img src="/images/team-logos/{game.home_team_logo_url}" alt="{game.home_team_name}" class="team-logo" />
                    {/if}
                    <span class="team-abbr">{game.home_team_abbr}</span>
                    {#if game.home_score !== null}
                      <span class="score">{game.home_score}</span>
                    {/if}
                    {#if game.home_spread !== null}
                      <span class="spread">{formatSpread(game.home_spread)}</span>
                    {/if}
                    {#if spreadWinner === 'home'}
                      <span class="spread-winner">✓</span>
                    {:else if spreadWinner === 'push'}
                      <span class="spread-push">-</span>
                    {/if}
                  </div>
                </div>
                <div class="meta">
                  <span>{formatWeekdayDateTime(game.kickoff_time)}</span>
                  <span class="game-status">{game.status}</span>
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
  .week-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .games-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
  }

  .game-card {
    padding: 12px;
  }

  .teams {
    display: flex;
    flex-direction: column;
    gap: 8px;
    font-weight: bold;
    margin-bottom: 8px;
  }

  .team-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .team-logo {
    width: 28px;
    height: 28px;
    object-fit: contain;
  }

  .team-abbr {
    min-width: 40px;
  }

  .score {
    font-size: 1.2rem;
    color: var(--custom-color-brand);
    min-width: 30px;
  }

  .spread {
    font-size: 0.85rem;
    color: var(--custom-color-secondary);
    min-width: 40px;
  }

  .spread-winner {
    color: var(--custom-color-brand);
    font-size: 1.2rem;
    font-weight: bold;
  }

  .spread-push {
    color: var(--custom-color-secondary);
    font-size: 1.2rem;
    font-weight: bold;
  }

  .meta {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
    opacity: 0.8;
  }
</style>