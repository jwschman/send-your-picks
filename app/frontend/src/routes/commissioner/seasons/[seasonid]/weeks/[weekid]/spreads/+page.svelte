<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import { logWeek } from '$lib/utils/logger'
  import { formatWeekdayDate, formatTime, getFullTeamName } from '$lib/utils/formatters'
  import type { Week } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let saving = false
  let importing = false
  let error = ''
  let successMessage = ''

  let week: Week | null = null
  let seasonId: string = ''
  let weekId: string = ''
  let spreads: { [gameId: string]: string } = {}
  let spreadErrors: { [gameId: string]: string } = {}

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
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

      // Initialize spreads object with current values (as strings)
      if (week) {
        week.games.forEach(game => {
          spreads[game.id] = game.home_spread !== null ? String(game.home_spread) : ''
        })
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

  async function autoImportSpreads() {
    if (!week) return

    if (isWeekLocked()) {
      error = 'Spreads can only be edited before the week is activated'
      return
    }

    logWeek('SPREADS_IMPORT_ATTEMPT', { weekId, seasonId })

    try {
      importing = true
      error = ''
      successMessage = ''

      const responseData = await api.post<{
        games_updated: number
        games_total: number
        bookmaker: string
        unmatched_games?: string[]
        warning?: string
      }>(
        `/api/commissioner/weeks/${weekId}/spreads/auto-import`,
        {},
        session?.access_token
      )

      // Refresh week data to get updated spreads
      await fetchWeek()

      let message = `Successfully imported spreads for ${responseData.games_updated} of ${responseData.games_total} games from ${responseData.bookmaker}`

      if (responseData.unmatched_games && responseData.unmatched_games.length > 0) {
        message += `\n\nNote: ${responseData.unmatched_games.length} game(s) could not be matched - you may need to set these manually.`
      }

      logWeek('SPREADS_IMPORT_SUCCESS', {
        weekId,
        seasonId,
        gameCount: responseData.games_total,
        spreadCount: responseData.games_updated,
        source: responseData.bookmaker
      })

      toasts.success(message)
      successMessage = message
      setTimeout(() => { successMessage = '' }, 5000)
    } catch (err) {
      logWeek('SPREADS_IMPORT_FAILURE', {
        weekId,
        seasonId,
        error: err instanceof Error ? err.message : 'Unknown error'
      })
      // Just show toast for action errors
      handleApiError(err, 'Failed to auto-import spreads')
    } finally {
      importing = false
    }
  }

  async function saveWeek() {
    if (!week) return

    if (isWeekLocked()) {
      error = 'Spreads can only be edited before the week is activated'
      return
    }

    // Validate all spreads
    spreadErrors = {}
    let hasErrors = false
    week.games.forEach(game => {
      const errorMsg = validateSpread(game.id, spreads[game.id])
      if (errorMsg) {
        spreadErrors[game.id] = errorMsg
        hasErrors = true
      }
    })

    if (hasErrors) {
      logWeek('SPREADS_VALIDATION_ERROR', {
        weekId,
        seasonId,
        error: 'Invalid spread values'
      })
      error = 'Invalid spread values detected. Please fix the errors shown below.'
      return
    }

    logWeek('SPREADS_SAVE_ATTEMPT', {
      weekId,
      seasonId,
      gameCount: week.games.length
    })

    try {
      saving = true
      error = ''
      successMessage = ''

      const games = week.games.map(game => {
        const spread = spreads[game.id]?.trim()
        const numSpread = !spread ? null : parseFloat(spread)
        return {
          game_id: game.id,
          home_spread: numSpread
        }
      })

      const responseData = await api.put<{ week: Week }>(
        `/api/commissioner/weeks/${weekId}/spreads`,
        { games },
        session?.access_token
      )
      week = responseData.week

      // Update spreads from response
      if (week) {
        week.games.forEach(game => {
          spreads[game.id] = game.home_spread !== null ? String(game.home_spread) : ''
        })
      }

      logWeek('SPREADS_SAVE_SUCCESS', {
        weekId,
        seasonId,
        spreadCount: games.filter(g => g.home_spread !== null).length
      })

      toasts.success('Spreads saved successfully!')
      goto(`/commissioner/seasons/${seasonId}/weeks/${weekId}`)
      return
    } catch (err) {
      logWeek('SPREADS_SAVE_FAILURE', {
        weekId,
        seasonId,
        error: err instanceof Error ? err.message : 'Unknown error'
      })
      // Just show toast for action errors
      handleApiError(err, 'Failed to update week')
    } finally {
      saving = false
    }
  }

  function getMissingSpreadsCount(): number {
    if (!week) return 0
    // Count games with no spread saved in the database
    return week.games.filter(game => game.home_spread === null).length
  }

  function hasValidationErrors(): boolean {
    return Object.keys(spreadErrors).length > 0
  }

  function isWeekLocked(): boolean {
    // Spreads can be edited when week is in games_imported or spreads_set status
    return week?.status !== 'games_imported' && week?.status !== 'spreads_set'
  }

  function validateSpread(gameId: string, spread: string | undefined | null): string {
    // Handle undefined/null
    if (spread === undefined || spread === null) {
      return '' // Empty is okay
    }

    // Trim whitespace
    const trimmed = spread.trim()

    if (trimmed === '') {
      return '' // Empty is okay
    }

    // Check if it matches valid number pattern (including partial inputs while typing):
    // - just minus (typing negative)
    // - digits with optional decimal and more digits
    // - decimal with optional digits (typing .5)
    const validNumberPattern = /^-?$|^-?\d+\.?\d*$|^-?\.\d*$/
    if (!validNumberPattern.test(trimmed)) {
      return 'Must be a valid number (e.g., -3.5, 7, .5)'
    }

    // Don't validate 0.5 increments for partial inputs (-, ., -., 1.)
    if (trimmed === '-' || trimmed === '.' || trimmed === '-.' || trimmed.endsWith('.')) {
      return ''
    }

    const numSpread = parseFloat(trimmed)
    if (isNaN(numSpread)) {
      return 'Must be a valid number'
    }

    // Check if it's a multiple of 0.5
    const remainder = Math.abs(numSpread % 0.5)
    if (remainder > 0.001) {
      return 'Must be a multiple of 0.5 (e.g., -3.5, 1.5, 7.0)'
    }

    return ''
  }

  function onSpreadChange(gameId: string) {
    const errorMsg = validateSpread(gameId, spreads[gameId])
    if (errorMsg) {
      spreadErrors[gameId] = errorMsg
    } else {
      delete spreadErrors[gameId]
    }
    spreadErrors = spreadErrors // Trigger reactivity
  }
</script>

<svelte:head>
  <title>{week ? `Edit Week ${week.number} - ${week.year}${week.is_postseason ? ' Postseason' : ''}` : 'Edit Week'}</title>
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
        <div class="page-container">
          <!-- Header -->
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/commissioner/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">Edit Week {week.number}</h1>
              <p class="text-sm opacity-half">{week.year} {week.is_postseason ? 'Postseason' : 'Season'} • Status: {week.status || 'draft'}</p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Validation Status -->
          <div class="card validation-status">
            {#if isWeekLocked()}
              <div class="status-message info">
                ℹ️ Week is in "{week.status}" status - spreads are locked after activation
              </div>
            {:else if getMissingSpreadsCount() === 0}
              <div class="status-message success">
                ✅ All spreads set
              </div>
            {:else}
              <div class="status-message warning">
                ⚠️ {getMissingSpreadsCount()} game(s) missing spreads
              </div>
            {/if}
          </div>

          <!-- Auto-Import Button -->
          {#if !isWeekLocked()}
            <div class="card auto-import-section">
              <div class="auto-import-header">
                <div>
                  <h3>Auto-Import Spreads</h3>
                  <p class="text-sm opacity-half">Automatically fetch spreads from the Odds API for all games</p>
                </div>
                <button
                  class="button secondary"
                  on:click={autoImportSpreads}
                  disabled={importing || saving}
                >
                  {importing ? '⏳ Importing...' : '📥 Auto-Import Spreads'}
                </button>
              </div>
              <div class="divider-text">
                <span>OR</span>
              </div>
              <p class="text-sm opacity-half" style="text-align: center;">
                Manually enter spreads below
              </p>
            </div>
          {/if}

          <!-- Games List -->
          <div class="games-section">
            <h2>Set Spreads</h2>
            <p class="text-sm opacity-half" style="margin-bottom: 20px;">
              Enter the home team spread for each game. Spreads must be in 0.5 increments. Negative means home team is favored.
            </p>

            <div class="games-list">
              {#each week.games as game}
                <div class="game-card card">
                  <div class="game-info">
                    <div class="matchup">
                      <div class="team away-team">
                        <span class="team-label text-sm opacity-half">Away</span>
                        <div class="team-name-with-logo">
                          {#if game.away_team_logo_url}
                            <img src="/images/team-logos/{game.away_team_logo_url}" alt="{game.away_team_name}" class="team-logo" />
                          {/if}
                          <span class="team-name">{getFullTeamName(game.away_team_city, game.away_team_name)}</span>
                        </div>
                      </div>
                      <div class="at-symbol">@</div>
                      <div class="team home-team">
                        <span class="team-label text-sm opacity-half">Home</span>
                        <div class="team-name-with-logo">
                          {#if game.home_team_logo_url}
                            <img src="/images/team-logos/{game.home_team_logo_url}" alt="{game.home_team_name}" class="team-logo" />
                          {/if}
                          <span class="team-name">{getFullTeamName(game.home_team_city, game.home_team_name)}</span>
                        </div>
                      </div>
                    </div>
                    <div class="game-time text-sm opacity-half">
                      {formatWeekdayDate(game.kickoff_time)} • {formatTime(game.kickoff_time)}
                    </div>
                  </div>

                  <div class="spread-input-group">
                    <label for="spread-{game.id}">Home Spread</label>
                    <input
                      id="spread-{game.id}"
                      type="text"
                      inputmode="decimal"
                      placeholder="e.g., -3.5 or 7"
                      bind:value={spreads[game.id]}
                      on:input={() => onSpreadChange(game.id)}
                      disabled={saving || isWeekLocked()}
                      class:invalid={spreadErrors[game.id]}
                    />
                    {#if spreadErrors[game.id]}
                      <span class="spread-error">{spreadErrors[game.id]}</span>
                    {:else if spreads[game.id]?.trim()}
                      <span class="spread-preview">
                        {parseFloat(spreads[game.id]) > 0 ? '+' : ''}{spreads[game.id]}
                      </span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="actions-footer card">
            <button
              class="button primary"
              on:click={() => saveWeek()}
              disabled={saving || isWeekLocked()}
            >
              {saving ? 'Saving...' : 'Save Spreads'}
            </button>
            <p class="text-sm opacity-half">
              Spreads are editable until the week is activated. Once you click the Activate button, spreads will be locked and cannot be changed.
            </p>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .page-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .validation-status {
    padding: 15px 20px;
    margin-bottom: 20px;
  }

  .status-message {
    font-weight: 500;
    font-size: 0.95rem;
  }

  .auto-import-section {
    padding: 20px;
    margin-bottom: 20px;
  }

  .auto-import-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    flex-wrap: wrap;
  }

  .auto-import-header h3 {
    font-size: 1.1rem;
    margin: 0 0 5px 0;
  }

  .divider-text {
    position: relative;
    text-align: center;
    margin: 20px 0;
  }

  .divider-text::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 1px;
    background: var(--custom-border-color);
  }

  .divider-text span {
    position: relative;
    background: var(--custom-panel-color);
    padding: 0 15px;
    font-size: 0.85rem;
    font-weight: 600;
    opacity: 0.5;
  }

  .status-message.success {
    color: #10b981;
  }

  .status-message.warning {
    color: #f59e0b;
  }

  .status-message.info {
    color: #3b82f6;
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

  .game-card {
    padding: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 30px;
    flex-wrap: wrap;
  }

  .game-info {
    flex: 1;
    min-width: 300px;
  }

  .matchup {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 10px;
  }

  .team {
    display: flex;
    flex-direction: column;
    gap: 5px;
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
    width: 35px;
    height: 35px;
    object-fit: contain;
  }

  .team-name {
    font-size: 1rem;
    font-weight: 500;
  }

  .spread-input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 200px;
  }

  .spread-input-group label {
    margin: 0;
  }

  .spread-input-group input {
    font-size: 1rem;
    padding: 10px;
  }

  .spread-preview {
    font-size: 0.85rem;
    color: var(--custom-color-brand);
    font-weight: 500;
  }

  .spread-error {
    font-size: 0.85rem;
    color: #ef4444;
    font-weight: 500;
  }

  input.invalid {
    border-color: #ef4444;
    background-color: rgba(239, 68, 68, 0.1);
  }

  .actions-footer {
    padding: 25px;
    position: sticky;
    bottom: 20px;
    background: var(--custom-panel-color);
    border: 2px solid var(--custom-color-brand);
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .actions-footer button {
    width: 100%;
  }

  @media only screen and (max-width: 45em) {
    .game-card {
      flex-direction: column;
      align-items: stretch;
    }

    .matchup {
      flex-direction: column;
      gap: 10px;
    }

    .spread-input-group {
      width: 100%;
    }

    .auto-import-header {
      flex-direction: column;
      align-items: stretch;
    }

    .auto-import-header button {
      width: 100%;
    }
  }
</style>