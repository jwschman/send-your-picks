<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import type { WeekWinnersData, UserWinCount, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let weekWinners: WeekWinnersData[] = []
  let userWinCounts: UserWinCount[] = []
  let badgesByUser: Record<string, Badge[]> = {}
  let seasonId: string = ''

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    if (!seasonId) {
      error = 'No season ID provided'
      return
    }

    fetchData()
  })

  async function fetchData() {
    try {
      loading = true
      error = ''

      const [weekData, userData, badgesRes] = await Promise.all([
        api.get<{ weeks: WeekWinnersData[] }>(`/api/seasons/${seasonId}/week-winners`, session?.access_token),
        api.get<{ users: UserWinCount[] }>(`/api/seasons/${seasonId}/win-counts`, session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token).catch((): { badges: Record<string, Badge[]> } => ({ badges: {} }))
      ])

      weekWinners = weekData.weeks || []
      userWinCounts = userData.users || []
      badgesByUser = badgesRes.badges
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch winners')
    } finally {
      loading = false
    }
  }

  function isCurrentUser(userId: string): boolean {
    return userId === session?.user?.id
  }
</script>

<svelte:head>
  <title>Weekly Winners</title>
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
            <a href="/seasons/{seasonId}" class="button">← Back to Season</a>
          </div>
        </div>
      {/if}

      {#if !loading && !error}
        <div class="winners-container">
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">Weekly Winners</h1>
              <p class="text-sm opacity-half">Season leaderboard by weekly wins</p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Win Counts Leaderboard -->
          <h2 class="section-title">Win Counts</h2>
          {#if userWinCounts.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No completed weeks yet.</p>
            </div>
          {:else}
            <div class="leaderboard-card card">
              <div class="leaderboard-header wins-header">
                <div class="rank-col">Rank</div>
                <div class="player-col">Player</div>
                <div class="stat-col">Wins</div>
                <div class="stat-col">Ties</div>
              </div>

              <div class="leaderboard-body">
                {#each userWinCounts as user, index}
                  <a
                    href="/users/{user.user_id}"
                    class="leaderboard-row wins-row"
                    class:current-user={isCurrentUser(user.user_id)}
                    class:top-three={index < 3}
                  >
                    <div class="rank-col">
                      <span class="rank-badge" class:rank-1={index === 0} class:rank-2={index === 1} class:rank-3={index === 2}>
                        {#if index === 0}
                          🥇
                        {:else if index === 1}
                          🥈
                        {:else if index === 2}
                          🥉
                        {:else}
                          {index + 1}
                        {/if}
                      </span>
                    </div>
                    <div class="player-col">
                      <img src={user.avatar_url} alt={user.username} class="avatar-small" />
                      <span class="player-name">
                        {user.username}
                        {#if isCurrentUser(user.user_id)}
                          <span class="you-badge">You</span>
                        {/if}
                        {#if badgesByUser[user.user_id]}
                          <BadgeList badges={badgesByUser[user.user_id]} compact />
                        {/if}
                      </span>
                    </div>
                    <div class="stat-col">
                      <span class="stat-value">{user.wins}</span>
                    </div>
                    <div class="stat-col">
                      <span class="stat-value tie-value">{user.ties}</span>
                    </div>
                  </a>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Week by Week Results -->
          <h2 class="section-title">Week by Week</h2>
          {#if weekWinners.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No completed weeks yet.</p>
            </div>
          {:else}
            <div class="weeks-grid">
              {#each weekWinners as week}
                <div class="week-winner-card card">
                  <div class="week-winner-header">
                    <span class="week-label">Week {week.week_number}</span>
                    <span class="week-points">{week.points} pts</span>
                  </div>
                  <div class="week-winner-body">
                    {#if week.is_tie}
                      <span class="tie-badge">TIE</span>
                    {/if}
                    <div class="winners-list">
                      {#each week.winners as winner}
                        <a
                          href="/users/{winner.user_id}"
                          class="winner-row"
                          class:is-you={isCurrentUser(winner.user_id)}
                        >
                          <img src={winner.avatar_url} alt={winner.username} class="avatar-small" />
                          <span class="winner-name">
                            {winner.username}
                            {#if isCurrentUser(winner.user_id)}
                              <span class="you-badge">You</span>
                            {/if}
                          </span>
                        </a>
                      {/each}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .winners-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .section-title {
    font-size: 1.2rem;
    margin: 30px 0 15px;
  }

  .section-title:first-of-type {
    margin-top: 0;
  }

  /* Win counts leaderboard adjustments */
  .wins-header, .wins-row {
    grid-template-columns: 80px 1fr 80px 80px;
  }

  .wins-row {
    text-decoration: none;
    color: inherit;
  }

  .stat-col {
    text-align: center;
  }

  .stat-value {
    font-weight: 600;
  }

  .tie-value {
    opacity: 0.7;
  }

  /* Avatar styles */
  .avatar-small {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
  }

  /* Week cards grid */
  .weeks-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 15px;
  }

  .week-winner-card {
    padding: 15px;
  }

  .week-winner-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--custom-border-color, #333);
  }

  .week-label {
    font-weight: 600;
    font-size: 1rem;
  }

  .week-points {
    font-size: 0.85rem;
    opacity: 0.7;
  }

  .week-winner-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .tie-badge {
    display: inline-block;
    padding: 2px 8px;
    background: rgba(255, 193, 7, 0.2);
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 700;
    color: #ffc107;
    width: fit-content;
  }

  .winners-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .winner-row {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    color: inherit;
    padding: 6px 8px;
    margin: -6px -8px;
    border-radius: 6px;
    transition: background 0.2s ease;
  }

  .winner-row:hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .winner-row.is-you {
    background: rgba(36, 180, 126, 0.1);
  }

  .winner-row.is-you:hover {
    background: rgba(36, 180, 126, 0.15);
  }

  .winner-name {
    font-size: 0.95rem;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  @media only screen and (max-width: 45em) {
    .wins-header, .wins-row {
      grid-template-columns: 60px 1fr 50px 50px;
      gap: 8px;
    }

    .weeks-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
