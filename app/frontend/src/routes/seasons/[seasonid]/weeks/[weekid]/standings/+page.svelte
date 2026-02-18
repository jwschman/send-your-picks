<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import type { Standing, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let standings: Standing[] = []
  let badgesByUser: Record<string, Badge[]> = {}
  let seasonId: string = ''
  let weekId: string = ''
  let weekNumber: number | null = null
  let seasonYear: number | null = null
  let isPostseason: boolean = false

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
      return
    }

    fetchStandings()
  })

  async function fetchStandings() {
    try {
      loading = true
      error = ''

      // Fetch week info (non-critical)
      try {
        const weekData = await api.get<{ week: { number: number; year: number; is_postseason: boolean } }>(`/api/weeks/${weekId}`, session?.access_token)
        weekNumber = weekData.week?.number
        seasonYear = weekData.week?.year
        isPostseason = weekData.week?.is_postseason ?? false
      } catch {
        // Week info is supplementary
      }

      // Fetch standings and badges
      const [standingsData, badgesRes] = await Promise.all([
        api.get<{ standings: Standing[] }>(`/api/weeks/${weekId}/standings`, session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token).catch((): { badges: Record<string, Badge[]> } => ({ badges: {} }))
      ])
      standings = standingsData.standings || []
      badgesByUser = badgesRes.badges
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch standings')
    } finally {
      loading = false
    }
  }

  function getRankDisplay(rank: number): string {
    if (rank === 1) return '🥇'
    if (rank === 2) return '🥈'
    if (rank === 3) return '🥉'
    return `${rank}`
  }

  function isCurrentUser(userId: string): boolean {
    return userId === session?.user?.id
  }
</script>

<svelte:head>
  <title>{weekNumber ? `Standings Through Week ${weekNumber}` : 'Standings'}</title>
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
            <a href="/seasons/{seasonId}/weeks/{weekId}" class="button">← Back to Week</a>
          </div>
        </div>
      {/if}

      {#if !loading && !error}
        <div class="standings-container">
          <!-- Header -->
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">{isPostseason ? 'Postseason' : 'Season'} Standings</h1>
              <p class="text-sm opacity-half">
                {#if weekNumber && seasonYear}
                  {seasonYear} {isPostseason ? 'Postseason' : 'Season'} • Through Week {weekNumber}
                {:else}
                  Current Standings
                {/if}
              </p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Leaderboard -->
          {#if standings.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No standings available yet.</p>
            </div>
          {:else}
            <div class="leaderboard-card card">
              <div class="leaderboard-header">
                <div class="rank-col">Rank</div>
                <div class="player-col">Player</div>
                <div class="points-col">Points</div>
              </div>
              
              <div class="leaderboard-body">
                {#each standings as standing}
                  <div 
                    class="leaderboard-row" 
                    class:current-user={isCurrentUser(standing.user_id)}
                    class:top-three={standing.rank <= 3}
                  >
                    <div class="rank-col">
                      <span class="rank-badge" class:rank-1={standing.rank === 1} class:rank-2={standing.rank === 2} class:rank-3={standing.rank === 3}>
                        {getRankDisplay(standing.rank)}
                      </span>
                    </div>
                    <div class="player-col">
                      <span class="player-name">
                        {standing.username}
                        {#if isCurrentUser(standing.user_id)}
                          <span class="you-badge">You</span>
                        {/if}
                        {#if badgesByUser[standing.user_id]}
                          <BadgeList badges={badgesByUser[standing.user_id]} compact />
                        {/if}
                      </span>
                    </div>
                    <div class="points-col">
                      <span class="points-value">{standing.points}</span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .standings-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }
</style>