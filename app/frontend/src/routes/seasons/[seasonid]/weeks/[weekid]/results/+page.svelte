<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import type { WeekResult, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let weekResults: WeekResult[] = []
  let badgesByUser: Record<string, Badge[]> = {}
  let seasonId: string = ''
  let weekId: string = ''
  let weekNumber: number | null = null

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
      return
    }

    fetchWeekResults()
  })

  async function fetchWeekResults() {
    try {
      loading = true
      error = ''

      // Fetch week info (non-critical)
      try {
        const weekData = await api.get<{ week: { number: number } }>(`/api/weeks/${weekId}`, session?.access_token)
        weekNumber = weekData.week?.number
      } catch {
        // Week info is supplementary
      }

      // Fetch results and badges
      const [resultsData, badgesRes] = await Promise.all([
        api.get<{ week_results: WeekResult[] }>(`/api/weeks/${weekId}/results`, session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token).catch((): { badges: Record<string, Badge[]> } => ({ badges: {} }))
      ])
      weekResults = resultsData.week_results || []
      badgesByUser = badgesRes.badges
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch results')
    } finally {
      loading = false
    }
  }

  function getUserDisplay(result: WeekResult): string {
    return result.username || 'Unknown User'
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
  <title>{weekNumber ? `Week ${weekNumber} Results` : 'Week Results'}</title>
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
        <div class="results-container">
          <!-- Header -->
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">Week {weekNumber || ''} Results</h1>
              <p class="text-sm opacity-half">Final Standings</p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Leaderboard -->
          {#if weekResults.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No results available yet.</p>
            </div>
          {:else}
            <div class="leaderboard-card card">
              <div class="leaderboard-header">
                <div class="rank-col">Rank</div>
                <div class="player-col">Player</div>
                <div class="points-col">Points</div>
              </div>
              
              <div class="leaderboard-body">
                {#each weekResults as result, index}
                  <div 
                    class="leaderboard-row" 
                    class:current-user={isCurrentUser(result.user_id)}
                    class:top-three={result.rank <= 3}
                  >
                    <div class="rank-col">
                      <span class="rank-badge" class:rank-1={result.rank === 1} class:rank-2={result.rank === 2} class:rank-3={result.rank === 3}>
                        {getRankDisplay(result.rank)}
                      </span>
                    </div>
                    <div class="player-col">
                      <span class="player-name">
                        {getUserDisplay(result)}
                        {#if isCurrentUser(result.user_id)}
                          <span class="you-badge">You</span>
                        {/if}
                        {#if badgesByUser[result.user_id]}
                          <BadgeList badges={badgesByUser[result.user_id]} compact />
                        {/if}
                      </span>
                    </div>
                    <div class="points-col">
                      <span class="points-value">{result.points}</span>
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
  .results-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }
</style>