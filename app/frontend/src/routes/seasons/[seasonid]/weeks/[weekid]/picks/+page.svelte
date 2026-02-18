<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import { logPicks } from '$lib/utils/logger'
  import { formatWeekdayDate, formatTime, getFullTeamName } from '$lib/utils/formatters'
  import type { Week, Settings, Pick } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let saving = false
  let error = ''
  let successMessage = ''

  let week: Week | null = null
  let seasonId: string = ''
  let weekId: string = ''
  let picks: { [gameId: string]: string | null } = {}
  let userPicksLocked = false
  let settings: Settings | null = null
  let cutoffMinutes = 120

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

      // Fetch settings (non-critical)
      try {
        const settingsData = await api.get<{ settings: Settings }>('/api/settings', session?.access_token)
        settings = settingsData.settings
        cutoffMinutes = settings.pick_cutoff_minutes
      } catch {
        // Settings are optional
      }

      // Fetch week data
      const weekData = await api.get<{ week: Week }>(`/api/weeks/${weekId}`, session?.access_token)
      week = weekData.week

      // Initialize picks object with empty values
      if (week) {
        week.games.forEach(game => {
          picks[game.id] = null
        })
      }

      // Fetch existing picks (non-critical)
      let existingPickCount = 0
      try {
        const picksData = await api.get<{ picks: Pick[] }>(`/api/weeks/${weekId}/picks`, session?.access_token)
        if (picksData.picks && Array.isArray(picksData.picks)) {
          picksData.picks.forEach((pick: Pick) => {
            picks[pick.game_id] = pick.selected_team_id
          })
          picks = { ...picks }
          existingPickCount = picksData.picks.length
          // Check if any picks are user-locked
          userPicksLocked = picksData.picks.some((pick: Pick) => pick.user_locked_at !== null)
        }
      } catch {
        // Existing picks are optional
      }

      logPicks('PAGE_LOAD', {
        weekId,
        seasonId,
        pickCount: existingPickCount
      })
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch week')
    } finally {
      loading = false
    }
  }

  function selectTeam(gameId: string, teamId: string) {
    if (userPicksLocked) {
      toasts.error('Your picks are locked and cannot be changed.')
      return
    }
    if (isGameLocked(gameId)) {
      logPicks('GAME_LOCKED', { weekId, gameId })
      return
    }
    const wasSelected = picks[gameId] === teamId
    picks[gameId] = wasSelected ? null : teamId
    picks = { ...picks }

    logPicks(wasSelected ? 'PICK_DESELECTED' : 'PICK_SELECTED', {
      weekId,
      gameId,
      team: teamId
    })
  }

  function isGameLocked(gameId: string): boolean {
    if (userPicksLocked) return true
    if (!week) return true
    const game = week.games.find(g => g.id === gameId)
    if (!game) return true
    if (cutoffMinutes === -1) return false

    const kickoffTime = new Date(game.kickoff_time)
    const cutoffTime = new Date(kickoffTime.getTime() - cutoffMinutes * 60 * 1000)
    const now = new Date()

    return now >= cutoffTime
  }

  function getPicksSubmittedCount(): number {
    return Object.values(picks).filter(pick => pick !== null).length
  }

  function getTotalGamesCount(): number {
    return week?.games.length || 0
  }

  function canSubmit(): boolean {
    if (!week) return false
    if (userPicksLocked) return false
    return true
  }

  async function savePicks() {
    if (!week || !canSubmit()) return

    const pickCount = getPicksSubmittedCount()
    const lockedCount = week.games.filter(g => isGameLocked(g.id)).length

    logPicks('SUBMIT_ATTEMPT', {
      weekId,
      seasonId,
      pickCount,
      lockedCount
    })

    try {
      saving = true
      error = ''
      successMessage = ''

      const picksArray = Object.entries(picks).map(([gameId, teamId]) => ({
        game_id: gameId,
        selected_team_id: teamId
      }))

      await api.put(`/api/weeks/${weekId}/picks`, { picks: picksArray }, session?.access_token)

      logPicks('SUBMIT_SUCCESS', {
        weekId,
        seasonId,
        pickCount,
        lockedCount
      })

      toasts.success('Picks saved successfully!')
      successMessage = 'Picks saved successfully! Redirecting...'
      setTimeout(() => {
        goto(`/seasons/${seasonId}/weeks/${weekId}`)
      }, 2000)
    } catch (err) {
      logPicks('SUBMIT_FAILURE', {
        weekId,
        seasonId,
        error: err instanceof Error ? err.message : 'Unknown error'
      })
      // Just show toast, don't set persistent error for save failures
      handleApiError(err, 'Failed to save picks')
    } finally {
      saving = false
    }
  }

  function getAwaySpread(homeSpread: number | null): string {
    if (homeSpread === null) return 'TBD'
    if (homeSpread === 0) return 'EVEN'
    const awaySpread = -homeSpread
    if (awaySpread > 0) return `+${awaySpread}`
    return awaySpread.toString()
  }

  function getHomeSpread(homeSpread: number | null): string {
    if (homeSpread === null) return 'TBD'
    if (homeSpread === 0) return 'EVEN'
    if (homeSpread > 0) return `+${homeSpread}`
    return homeSpread.toString()
  }

  function getTimeUntilLock(gameId: string): string {
    if (!week) return ''
    const game = week.games.find(g => g.id === gameId)
    if (!game) return ''

    // Testing mode
    if (cutoffMinutes === -1) return 'Never (testing mode)'

    const kickoffTime = new Date(game.kickoff_time)
    const cutoffTime = new Date(kickoffTime.getTime() - cutoffMinutes * 60 * 1000)
    const now = new Date()
    const diff = cutoffTime.getTime() - now.getTime()

    if (diff <= 0) return 'Locked'

    const hours = Math.floor(diff / (1000 * 60 * 60))
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

    if (hours > 24) {
      const days = Math.floor(hours / 24)
      return `${days}d ${hours % 24}h`
    }
    return `${hours}h ${minutes}m`
  }
</script>

<svelte:head>
  <title>{week ? `Make Picks - Week ${week.number}` : 'Make Picks'}</title>
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

      {#if successMessage}
        <div class="card success-card">
          <p class="success">{successMessage}</p>
        </div>
      {/if}

      {#if week && !loading}
        <div class="picks-container">
          <!-- Header -->
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">{userPicksLocked ? 'View Your Picks' : 'Make Your Picks'}</h1>
              <p class="text-sm opacity-half">Week {week.number} - {week.year} {week.is_postseason ? 'Postseason' : 'Season'}</p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Pick Status -->
          <div class="card status-card">
            {#if userPicksLocked}
              <div class="status-message locked">
                🔒 Your picks are locked ({getPicksSubmittedCount()}/{getTotalGamesCount()})
              </div>
            {:else if getPicksSubmittedCount() === getTotalGamesCount()}
              <div class="status-message complete">
                ✅ All picks submitted ({getPicksSubmittedCount()}/{getTotalGamesCount()})
              </div>
            {:else}
              <div class="status-message incomplete">
                📋 {getPicksSubmittedCount()}/{getTotalGamesCount()} picks submitted
              </div>
            {/if}
            <p class="text-sm opacity-half" style="margin-top: 8px;">
              {#if userPicksLocked}
                Your picks have been locked and cannot be changed.
              {:else if cutoffMinutes === -1}
                🧪 Testing mode: Picks never lock
              {:else}
                Picks lock {cutoffMinutes} minute{cutoffMinutes !== 1 ? 's' : ''} before each game's kickoff
              {/if}
            </p>
          </div>

          <!-- Games List -->
          <div class="games-section">
            <h2>{userPicksLocked ? 'Your Picks' : 'Select Your Picks'}</h2>
            {#if !userPicksLocked}
              <p class="text-sm opacity-half" style="margin-bottom: 20px;">
                Click on a team to select them. Click again to deselect. You must beat the spread to win.
              </p>
            {/if}

            <div class="games-list">
              {#each week.games as game}
                {@const locked = isGameLocked(game.id)}
                {@const awaySelected = picks[game.id] === game.away_team_id}
                {@const homeSelected = picks[game.id] === game.home_team_id}

                <div class="pick-card card" class:locked>
                  <div class="game-time-info text-sm opacity-half">
                    <div>{formatWeekdayDate(game.kickoff_time)} • {formatTime(game.kickoff_time)}</div>
                    <div class="lock-time" class:warning={!locked && getTimeUntilLock(game.id) !== 'Locked'}>
                      {locked ? '🔒 Locked' : `Locks in ${getTimeUntilLock(game.id)}`}
                    </div>
                  </div>

                  <div class="pick-options">
                    <!-- Away Team -->
                    <button
                      class="team-option"
                      class:selected={awaySelected}
                      class:disabled={locked}
                      on:click={() => selectTeam(game.id, game.away_team_id)}
                      disabled={locked}
                    >
                      <div class="team-info">
                        {#if game.away_team_logo_url}
                          <img src="/images/team-logos/{game.away_team_logo_url}" alt="{game.away_team_name}" class="team-logo" />
                        {/if}
                        <div class="team-name">{getFullTeamName(game.away_team_city, game.away_team_name)}</div>
                        <div class="team-label text-sm opacity-half">Away</div>
                        <div class="team-spread">{getAwaySpread(game.home_spread)}</div>
                      </div>
                      {#if awaySelected}
                        <div class="selected-indicator">✓</div>
                      {/if}
                    </button>

                    <!-- At Symbol -->
                    <div class="spread-divider">
                      <div class="at-symbol">@</div>
                    </div>

                    <!-- Home Team -->
                    <button
                      class="team-option"
                      class:selected={homeSelected}
                      class:disabled={locked}
                      on:click={() => selectTeam(game.id, game.home_team_id)}
                      disabled={locked}
                    >
                      <div class="team-info">
                        {#if game.home_team_logo_url}
                          <img src="/images/team-logos/{game.home_team_logo_url}" alt="{game.home_team_name}" class="team-logo" />
                        {/if}
                        <div class="team-name">{getFullTeamName(game.home_team_city, game.home_team_name)}</div>
                        <div class="team-label text-sm opacity-half">Home</div>
                        <div class="team-spread">{getHomeSpread(game.home_spread)}</div>
                      </div>
                      {#if homeSelected}
                        <div class="selected-indicator">✓</div>
                      {/if}
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- Submit Button -->
          {#if !userPicksLocked}
            <div class="submit-section card">
              <button
                class="button primary submit-button"
                on:click={savePicks}
                disabled={saving || !canSubmit()}
              >
                {saving ? 'Saving...' : 'Submit Picks'}
              </button>
              <p class="text-sm opacity-half">
                {#if getPicksSubmittedCount() === 0}
                  Select at least one pick to submit
                {:else if getPicksSubmittedCount() < getTotalGamesCount()}
                  You can submit partial picks and come back to complete them later
                {:else}
                  All picks submitted! You can still edit unlocked picks
                {/if}
              </p>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .picks-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .status-card {
    padding: 20px;
    margin-bottom: 30px;
  }

  .status-message {
    font-size: 1.1rem;
    font-weight: 500;
  }

  .status-message.complete {
    color: var(--custom-color-brand);
  }

  .status-message.incomplete {
    color: #f59e0b;
  }

  .status-message.locked {
    color: var(--custom-color-secondary);
  }

  .games-section {
    margin-bottom: 30px;
  }

  .games-section h2 {
    font-size: 1.3rem;
    margin-bottom: 10px;
  }

  .games-list {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .pick-card {
    padding: 15px;
    transition: all 0.2s ease;
  }

  .game-time-info {
    display: flex;
    justify-content: space-between;
    margin-bottom: 15px;
    padding-bottom: 10px;
    border-bottom: var(--custom-border);
  }

  .lock-time {
    font-weight: 500;
  }

  .lock-time.warning {
    color: #f59e0b;
  }

  .pick-options {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    gap: 15px;
    align-items: stretch;
  }

  .team-option {
    background-color: var(--custom-panel-color);
    border: 2px solid var(--custom-border-color, #333);
    border-radius: var(--custom-border-radius);
    padding: 20px;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: center;
    min-height: 80px;
  }

  .team-option:hover:not(.disabled) {
    border-color: var(--custom-color-brand);
    background-color: rgba(36, 180, 126, 0.1);
    transform: translateY(-2px);
  }

  .team-option.selected {
    border-color: var(--custom-color-brand);
    background-color: rgba(36, 180, 126, 0.2);
    border-width: 3px;
  }

  .team-option.disabled {
    cursor: not-allowed;
    opacity: 1;
    pointer-events: auto;
  }

  .team-info {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .team-logo {
    width: 50px;
    height: 50px;
    object-fit: contain;
  }

  .team-name {
    font-size: 1rem;
    font-weight: 500;
    margin-bottom: 5px;
  }

  .team-label {
    text-transform: uppercase;
    font-size: 0.7rem;
    margin-bottom: 8px;
  }

  .team-spread {
    font-size: 1.3rem;
    font-weight: bold;
    color: var(--custom-color-brand);
    margin-top: 8px;
  }

  .selected-indicator {
    position: absolute;
    top: 8px;
    right: 8px;
    font-size: 1.5rem;
    color: var(--custom-color-brand);
  }

  .spread-divider {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .at-symbol {
    font-size: 1.5rem;
    color: var(--custom-color-secondary);
  }

  .submit-section {
    padding: 25px;
    text-align: center;
    position: sticky;
    bottom: 20px;
    background: var(--custom-panel-color);
    border: 2px solid var(--custom-color-brand);
  }

  .submit-button {
    width: 100%;
    max-width: 400px;
    padding: 15px;
    font-size: 1.1rem;
    margin-bottom: 10px;
  }

  @media only screen and (max-width: 45em) {
    .pick-options {
      grid-template-columns: 1fr;
      gap: 10px;
    }

    .spread-divider {
      flex-direction: row;
      order: -1;
    }
  }
</style>