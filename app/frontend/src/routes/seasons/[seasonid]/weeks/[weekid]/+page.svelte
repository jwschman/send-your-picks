<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import { toasts } from '$lib/stores/toast'
  import ConfirmModal from '$lib/components/ConfirmModal.svelte'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { formatWeekdayDateTime, formatDateTime, getFullTeamName } from '$lib/utils/formatters'
  import type { Week, Pick } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''
  let showLockConfirm = false

  let week: Week | null = null
  let picks: Pick[] = []
  let seasonId: string = ''
  let weekId: string = ''

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
      return
    }

    fetchWeekAndPicks()
  })

  async function fetchWeekAndPicks() {
    try {
      loading = true
      error = ''

      const weekData = await api.get<{ week: Week }>(`/api/weeks/${weekId}`, session?.access_token)
      week = weekData.week

      try {
        const picksData = await api.get<{ picks: Pick[] }>(`/api/weeks/${weekId}/picks`, session?.access_token)
        picks = picksData.picks || []
      } catch {
        // Picks fetch is not critical
        picks = []
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        error = 'Week not found'
      } else {
        error = handleApiError(err, 'Failed to fetch week')
      }
    } finally {
      loading = false
    }
  }

  function showLockConfirmation() {
    showLockConfirm = true
  }

  async function confirmLockPicks() {
    // Validate that user has picks to lock
    if (getPicksCount() === 0) {
      toasts.error('You must make at least one pick before locking.')
      return
    }

    try {
      loading = true

      await api.post(`/api/weeks/${weekId}/picks/lock`, {}, session?.access_token)
      toasts.success('Picks locked successfully!')
      await fetchWeekAndPicks()
    } catch (err) {
      // Just show toast for action errors, don't set persistent error
      handleApiError(err, 'Failed to lock picks')
    } finally {
      loading = false
    }
  }

  function getAwaySpread(homeSpread: number | null): number | null {
    if (homeSpread === null) return null
    return -homeSpread
  }

  function getStatusBadgeClass(status: string): string {
    switch (status?.toLowerCase()) {
      case 'draft':
      case 'scheduled':
        return 'status-draft'
      case 'active':
      case 'in_progress':
        return 'status-active'
      case 'played':
      case 'final':
        return 'status-played'
      default:
        return 'status-draft'
    }
  }

  function getPicksCount(): number {
    return picks.filter(pick => pick.selected_team_id !== null).length
  }

  function getTotalGamesCount(): number {
    return week?.games.length || 0
  }

  function getPickForGame(gameId: string): Pick | undefined {
    return picks.find(pick => pick.game_id === gameId)
  }

  function isTeamPicked(gameId: string, teamId: string): boolean {
    const pick = getPickForGame(gameId)
    return pick?.selected_team_id === teamId
  }

  function canMakePicks(): boolean {
    return week?.status === 'active'
  }

  function canViewResults(): boolean {
    return week?.status === 'final'
  }

  function canViewStandings(): boolean {
    return week?.status === 'final'
  }

  function canViewAllPicks(): boolean {
    // Can view all picks once week is active (picks are being locked)
    const activeStatuses = ['active', 'played', 'picks_results_calculated', 'scored', 'final']
    return week?.status ? activeStatuses.includes(week.status) : false
  }

  function areAllPicksLocked(): boolean {
    const completed = picks.filter(p => p.selected_team_id !== null)
    if (completed.length === 0) return false
    return completed.every(p => p.user_locked_at !== null)
  }
</script>

<svelte:head>
  <title>{week ? `Week ${week.number} - ${week.year}${week.is_postseason ? ' Postseason' : ''}` : 'Week'}</title>
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
        </div>
      {/if}

      {#if week && !loading}
        <div class="week-container">
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">Week {week.number}</h1>
              <p class="text-sm opacity-half">{week.year} {week.is_postseason ? 'Postseason' : 'Season'}</p>
            </div>
            <div class="header-right">
              <span class="status-badge {getStatusBadgeClass(week.status)}">
                {week.status}
              </span>
            </div>
          </div>

          <div class="card picks-status-card">
            {#if getPicksCount() === getTotalGamesCount()}
              <div class="status-message complete">
                All picks submitted ({getPicksCount()}/{getTotalGamesCount()})
              </div>
              {#if areAllPicksLocked()}
                <div class="text-sm opacity-half">Picks are locked</div>
              {/if}
            {:else if getPicksCount() > 0}
              <div class="status-message partial">
                {getPicksCount()}/{getTotalGamesCount()} picks submitted
              </div>
            {:else}
              <div class="status-message none">
                No picks submitted yet ({getPicksCount()}/{getTotalGamesCount()})
              </div>
            {/if}
          </div>

          <div class="card actions-panel">
            <h3>Actions</h3>
            <div class="button-grid">
              {#if canMakePicks() && !areAllPicksLocked()}
                <a
                  href="/seasons/{seasonId}/weeks/{weekId}/picks"
                  class="button primary"
                >
                  Make or Edit Picks
                </a>
              {:else}
                <a
                  href="/seasons/{seasonId}/weeks/{weekId}/picks"
                  class="button"
                >
                  View My Picks
                </a>
              {/if}
              <button
                class="button {( !canMakePicks() || areAllPicksLocked() ) ? 'disabled' : ''}"
                class:disabled={!canMakePicks() || areAllPicksLocked()}
                on:click={showLockConfirmation}
              >
                Lock Picks
              </button>
              <a 
                href="/seasons/{seasonId}/weeks/{weekId}/results"
                class="button {!canViewResults() ? 'disabled' : ''}"
                class:disabled={!canViewResults()}
              >
                View Week Results
              </a>
              <a
                href="/seasons/{seasonId}/weeks/{weekId}/standings"
                class="button {!canViewStandings() ? 'disabled' : ''}"
                class:disabled={!canViewStandings()}
              >
                View Standings
              </a>
              <a
                href="/seasons/{seasonId}/weeks/{weekId}/allpicks"
                class="button {!canViewAllPicks() ? 'disabled' : ''}"
                class:disabled={!canViewAllPicks()}
              >
                View All Picks
              </a>
            </div>
          </div>

          <div class="games-section">
            <h2>Games ({week.games.length})</h2>
            
            {#if week.games.length === 0}
              <div class="card">
                <p class="text-sm opacity-half">No games scheduled for this week yet.</p>
              </div>
            {:else}
              <div class="games-list">
                {#each week.games as game}
                  {@const pick = getPickForGame(game.id)}
                  {@const awayPicked = isTeamPicked(game.id, game.away_team_id)}
                  {@const homePicked = isTeamPicked(game.id, game.home_team_id)}
                  
                  <div class="game-card card">
                    <div class="game-time-header text-sm opacity-half">
                      {formatWeekdayDateTime(game.kickoff_time)}
                    </div>
                    <div class="game-info">
                      <div class="matchup">
                        <div class="team away-team" class:picked={awayPicked}>
                          <span class="team-label text-sm opacity-half">Away</span>
                          <div class="team-name-with-logo">
                            {#if game.away_team_logo_url}
                              <img src="/images/team-logos/{game.away_team_logo_url}" alt="{game.away_team_name}" class="team-logo" />
                            {/if}
                            <span class="team-name">
                              {getFullTeamName(game.away_team_city, game.away_team_name)}
                              {#if awayPicked}
                                <span class="pick-indicator">✓</span>
                              {/if}
                            </span>
                          </div>
                          {#if game.away_score !== null}
                            <span class="score">{game.away_score}</span>
                          {/if}
                          {#if game.home_spread !== null}
                            {@const awaySpread = getAwaySpread(game.home_spread)}
                            <span class="spread text-sm opacity-half">
                              Spread: {awaySpread !== null && awaySpread > 0 ? '+' : ''}{awaySpread}
                            </span>
                          {/if}
                        </div>
                        <div class="at-symbol">@</div>
                        <div class="team home-team" class:picked={homePicked}>
                          <span class="team-label text-sm opacity-half">Home</span>
                          <div class="team-name-with-logo">
                            {#if game.home_team_logo_url}
                              <img src="/images/team-logos/{game.home_team_logo_url}" alt="{game.home_team_name}" class="team-logo" />
                            {/if}
                            <span class="team-name">
                              {getFullTeamName(game.home_team_city, game.home_team_name)}
                              {#if homePicked}
                                <span class="pick-indicator">✓</span>
                              {/if}
                            </span>
                          </div>
                          {#if game.home_score !== null}
                            <span class="score">{game.home_score}</span>
                          {/if}
                          {#if game.home_spread !== null}
                            <span class="spread text-sm opacity-half">
                              Spread: {game.home_spread > 0 ? '+' : ''}{game.home_spread}
                            </span>
                          {/if}
                        </div>
                      </div>
                      <div class="game-details">
                        <span class="game-status text-sm">
                          Status: <span class="status-badge-small {getStatusBadgeClass(game.status)}">{game.status}</span>
                        </span>
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<ConfirmModal
  bind:open={showLockConfirm}
  title="Lock Picks"
  message="Once locked, you will not be able to change your picks. Are you sure you want to lock them?"
  confirmText="Lock Picks"
  confirmVariant="primary"
  on:confirm={confirmLockPicks}
/>

<style>
  .week-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .picks-status-card {
    padding: 20px;
    margin-bottom: 20px;
  }

  .status-message {
    font-size: 1.1rem;
    font-weight: 500;
  }

  .status-message.complete {
    color: var(--custom-color-brand);
  }

  .status-message.partial {
    color: #f59e0b;
  }

  .status-message.none {
    color: #ef4444;
  }

  .actions-panel {
    padding: 25px;
    margin-bottom: 30px;
  }

  .actions-panel h3 {
    margin: 0 0 20px 0;
    font-size: 1.1rem;
  }

  .button-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 12px;
  }

  .button-grid .button {
    margin: 0;
    text-align: center;
  }

  .games-section {
    margin-bottom: 30px;
  }

  .games-section h2 {
    font-size: 1.3rem;
    margin-bottom: 20px;
  }

  .games-list {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .game-card {
    padding: 20px;
  }

  .game-time-header {
    text-align: center;
    margin-bottom: 15px;
    padding-bottom: 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    font-weight: 500;
  }

  .game-info {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .matchup {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .team {
    display: flex;
    flex-direction: column;
    gap: 5px;
    flex: 1;
    padding: 10px;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .team.picked {
    background-color: rgba(36, 180, 126, 0.15);
    border: 2px solid var(--custom-color-brand);
  }

  .team-label {
    text-transform: uppercase;
    font-size: 0.7rem;
  }

  .team-name-with-logo {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .team-logo {
    width: 40px;
    height: 40px;
    object-fit: contain;
  }

  .team-name {
    font-size: 1rem;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pick-indicator {
    color: var(--custom-color-brand);
    font-size: 1.2rem;
    font-weight: bold;
  }

  .spread {
    font-weight: 500;
  }

  .score {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--custom-color-brand);
  }

  .at-symbol {
    color: var(--custom-color-secondary);
    font-size: 1.2rem;
  }

  .game-details {
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
    align-items: center;
  }

  .game-status {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-badge-small {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 8px;
    font-size: 0.65rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  @media only screen and (max-width: 45em) {
    .button-grid {
      grid-template-columns: 1fr;
    }

    .matchup {
      flex-direction: column;
      gap: 10px;
    }

    .game-details {
      flex-direction: column;
      gap: 8px;
    }
  }
</style>