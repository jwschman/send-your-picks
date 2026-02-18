<script lang="ts">
  import { env } from '$env/dynamic/public'
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import EmptyState from '$lib/components/EmptyState.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import { formatTime, DEFAULT_AVATAR } from '$lib/utils/formatters'
  import type { Week, Game, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = true
  let error = ''

  // Specialized local types for this page
  type PickDetail = {
    game_id: string
    selected_team_id: string
    is_correct: boolean | null
  }

  type UserWithPicks = {
    user_id: string
    username: string
    avatar_url: string | null
    picks: PickDetail[]
  }

  let week: Week | null = null
  let users: UserWithPicks[] = []
  let badgesByUser: Record<string, Badge[]> = {}
  let seasonId = ''
  let weekId = ''

  onMount(async () => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
      loading = false
      return
    }

    await fetchData()
  })

  async function fetchData() {
    try {
      loading = true
      error = ''

      // Fetch week data, picks, and badges
      const [weekData, picksData, badgesRes] = await Promise.all([
        api.get<{ week: Week }>(`/api/weeks/${weekId}`, session?.access_token),
        api.get<{ users: UserWithPicks[] }>(`/api/weeks/${weekId}/picks/locked`, session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token).catch((): { badges: Record<string, Badge[]> } => ({ badges: {} }))
      ])
      week = weekData.week
      users = picksData.users || []
      badgesByUser = badgesRes.badges

    } catch (err) {
      error = handleApiError(err, 'Failed to fetch data')
    } finally {
      loading = false
    }
  }

  // Helper: Get user's pick for a specific game
  function getUserPickForGame(user: UserWithPicks, gameId: string): PickDetail | null {
    return user.picks.find(p => p.game_id === gameId) || null
  }

  // Helper: Check if user picked the away team
  function pickedAwayTeam(user: UserWithPicks, game: Game): PickDetail | null {
    const pick = getUserPickForGame(user, game.id)
    if (pick?.selected_team_id === game.away_team_id) {
      return pick
    }
    return null
  }

  // Helper: Check if user picked the home team
  function pickedHomeTeam(user: UserWithPicks, game: Game): PickDetail | null {
    const pick = getUserPickForGame(user, game.id)
    if (pick?.selected_team_id === game.home_team_id) {
      return pick
    }
    return null
  }

  // Helper: Get full avatar URL from storage path
  function getAvatarUrl(avatarPath: string | null): string {
    if (!avatarPath) return DEFAULT_AVATAR
    return `${env.PUBLIC_SUPABASE_URL}/storage/v1/object/public/avatars/${avatarPath}`
  }

  // Helper: Check if game result is known
  function isGameFinal(game: Game): boolean {
    return game.status === 'final' || game.status === 'Final'
  }

  // Helper: Get pick display symbol
  function getPickSymbol(pick: PickDetail | null, game: Game): string {
    if (!pick) return ''

    if (!isGameFinal(game)) {
      return '●' // Dot for picks on non-final games
    }

    // Game is final - check result
    if (pick.is_correct === null) {
      return '—' // Dash for push/draw
    }

    return pick.is_correct ? '✓' : '✗'
  }

  // Helper: Get pick CSS class
  function getPickClass(pick: PickDetail | null, game: Game): string {
    if (!pick) return ''

    if (!isGameFinal(game)) {
      return 'pending'
    }

    // Game is final - check result
    if (pick.is_correct === null) {
      return 'push' // Push/draw
    }

    return pick.is_correct ? 'correct' : 'incorrect'
  }
</script>

<svelte:head>
  <title>{week ? `All Picks - Week ${week.number}` : 'All Picks'}</title>
</svelte:head>

<div class="container">
  <div class="row">
    <div class="col-12">
      <div class="allpicks-container">
        <div class="page-header-three-col">
          <div class="header-left">
            <a href="/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
          </div>
          <div class="header-title-group">
            <h1 class="header-title">All Picks</h1>
            {#if week}
              <p class="text-sm opacity-half">Week {week.number} - {week.year} {week.is_postseason ? 'Postseason' : 'Season'}</p>
            {/if}
          </div>
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

        {#if !loading && !error && week}
          {#if users.length === 0}
            <div class="card">
              <EmptyState
                icon="👥"
                title="No Users Yet"
                message="Once users start making picks, they'll appear here."
              />
            </div>
          {:else}
            <div class="card picks-table-card">
              <div class="picks-table-wrapper">
                <table class="picks-table">
                  <thead>
                    <tr>
                      <!-- Away team user columns -->
                      {#each users as user}
                        <th class="user-col away" title={user.username}>
                          <img src={getAvatarUrl(user.avatar_url)} alt={user.username} class="avatar" />
                        </th>
                      {/each}

                      <!-- Game info column -->
                      <th class="game-col">Matchup</th>

                      <!-- Home team user columns -->
                      {#each users as user}
                        <th class="user-col home" title={user.username}>
                          <img src={getAvatarUrl(user.avatar_url)} alt={user.username} class="avatar" />
                        </th>
                      {/each}
                    </tr>
                  </thead>
                  <tbody>
                    {#each week.games as game}
                      {@const gameIsFinal = isGameFinal(game)}
                      <tr class:final={gameIsFinal}>
                        <!-- Away team picks -->
                        {#each users as user}
                          {@const pick = pickedAwayTeam(user, game)}
                          <td class="pick-cell away">
                            {#if pick}
                              <span class="pick-mark {getPickClass(pick, game)}">
                                {getPickSymbol(pick, game)}
                              </span>
                            {/if}
                          </td>
                        {/each}

                        <!-- Game info -->
                        <td class="game-cell">
                          <div class="matchup">
                            <span class="team away-team">
                              {#if game.away_team_logo_url}
                                <img src="/images/team-logos/{game.away_team_logo_url}" alt="{game.away_team_name}" class="team-logo-small" />
                              {/if}
                              {game.away_team_abbr}
                            </span>
                            <span class="at">@</span>
                            <span class="team home-team">
                              {#if game.home_team_logo_url}
                                <img src="/images/team-logos/{game.home_team_logo_url}" alt="{game.home_team_name}" class="team-logo-small" />
                              {/if}
                              {game.home_team_abbr}
                            </span>
                          </div>
                          <div class="game-time text-sm opacity-half">
                            {formatTime(game.kickoff_time)}
                          </div>
                          {#if gameIsFinal}
                            <div class="game-score">
                              {game.away_score} - {game.home_score}
                            </div>
                          {/if}
                        </td>

                        <!-- Home team picks -->
                        {#each users as user}
                          {@const pick = pickedHomeTeam(user, game)}
                          <td class="pick-cell home">
                            {#if pick}
                              <span class="pick-mark {getPickClass(pick, game)}">
                                {getPickSymbol(pick, game)}
                              </span>
                            {/if}
                          </td>
                        {/each}
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Legend -->
            <div class="card legend-card">
              <h3>Users</h3>
              <div class="user-legend">
                {#each users as user}
                  <div class="legend-item">
                    <img src={getAvatarUrl(user.avatar_url)} alt={user.username} class="legend-avatar" />
                    <span class="legend-name">{user.username}</span>
                    {#if badgesByUser[user.user_id]}
                      <BadgeList badges={badgesByUser[user.user_id]} compact />
                    {/if}
                    <span class="legend-picks text-sm opacity-half">
                      ({user.picks.length} pick{user.picks.length !== 1 ? 's' : ''})
                    </span>
                  </div>
                {/each}
              </div>
            </div>

            <!-- Key -->
            <div class="card key-card">
              <h3>Key</h3>
              <div class="key-items">
                <div class="key-item">
                  <span class="pick-mark pending">●</span>
                  <span>Locked pick (game not final)</span>
                </div>
                <div class="key-item">
                  <span class="pick-mark correct">✓</span>
                  <span>Correct pick</span>
                </div>
                <div class="key-item">
                  <span class="pick-mark incorrect">✗</span>
                  <span>Incorrect pick</span>
                </div>
                <div class="key-item">
                  <span class="pick-mark push">—</span>
                  <span>Push (no winner)</span>
                </div>
              </div>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .allpicks-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .picks-table-card {
    overflow: hidden;
  }

  .picks-table-wrapper {
    overflow-x: auto;
  }

  .picks-table {
    width: 100%;
    border-collapse: collapse;
    min-width: 600px;
  }

  .picks-table th,
  .picks-table td {
    padding: 8px 4px;
    text-align: center;
    border-bottom: 1px solid var(--custom-border-color, #333);
  }

  .picks-table th {
    font-weight: 600;
    font-size: 0.85rem;
    padding-top: 12px;
    padding-bottom: 12px;
  }

  /* Avatar styles */
  .avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: inline-block;
    object-fit: cover;
  }

  /* User columns */
  .user-col {
    min-width: 36px;
    max-width: 44px;
  }

  .user-col.away {
    background: rgba(255, 100, 100, 0.08);
  }

  .user-col.home {
    background: rgba(100, 150, 255, 0.08);
  }

  /* Game column */
  .game-col {
    min-width: 120px;
    background: var(--custom-panel-color);
  }

  .game-cell {
    background: var(--custom-panel-color);
    padding: 10px 8px;
  }

  .matchup {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 6px;
    font-weight: 600;
    font-size: 1.05rem;
  }

  .team {
    min-width: 35px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .team.away-team {
    text-align: right;
    justify-content: flex-end;
  }

  .team-logo-small {
    width: 20px;
    height: 20px;
    object-fit: contain;
  }

  .team.home-team {
    text-align: left;
  }

  .at {
    opacity: 0.5;
    font-size: 0.8rem;
  }

  .game-time {
    margin-top: 2px;
    font-size: 0.75rem;
  }

  .game-score {
    margin-top: 3px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  /* Pick cells */
  .pick-cell {
    min-width: 36px;
    max-width: 44px;
  }

  .pick-cell.away {
    background: rgba(255, 100, 100, 0.03);
  }

  .pick-cell.home {
    background: rgba(100, 150, 255, 0.03);
  }

  .pick-mark {
    font-weight: bold;
    font-size: 1rem;
  }

  .pick-mark.pending {
    color: var(--custom-color-secondary, #888);
  }

  .pick-mark.correct {
    color: #10b981;
  }

  .pick-mark.incorrect {
    color: #ef4444;
  }

  .pick-mark.push {
    color: #f59e0b;
  }

  /* Final game rows */
  tr.final .game-cell {
    background: rgba(255, 255, 255, 0.02);
  }

  /* Hover effect on rows */
  .picks-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .picks-table tbody tr:hover .game-cell {
    background: rgba(255, 255, 255, 0.05);
  }

  /* Legend */
  .legend-card,
  .key-card {
    margin-top: 20px;
  }

  .legend-card h3,
  .key-card h3 {
    margin: 0 0 15px 0;
    font-size: 1rem;
  }

  .user-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .legend-avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
  }

  .legend-name {
    font-size: 0.9rem;
  }

  .legend-picks {
    font-size: 0.8rem;
  }

  /* Key */
  .key-items {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
  }

  .key-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.9rem;
  }

  .key-item .pick-mark {
    width: 20px;
    text-align: center;
  }
</style>
