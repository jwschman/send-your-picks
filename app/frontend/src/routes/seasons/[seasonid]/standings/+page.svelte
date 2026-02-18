<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import { Chart, registerables } from 'chart.js'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { logData } from '$lib/utils/logger'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import type { Standing, Badge, StandingsHistoryEntry, ChartDataset } from '$lib/types'

  Chart.register(...registerables)

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let standings: Standing[] = []
  let badgesByUser: Record<string, Badge[]> = {}
  let seasonId: string = ''
  let seasonYear: number | null = null
  let isPostseason: boolean = false

  let chartCanvas: HTMLCanvasElement
  let chart: Chart | null = null
  let chartData: { labels: string[]; datasets: ChartDataset[] } | null = null

  const colors = [
    '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF',
    '#FF9F40', '#FF9F40', '#C9CBCF', '#4BC0C0', '#FF6384'
  ]

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''

    if (!seasonId) {
      error = 'Missing season ID'
      return
    }

    fetchStandings()
    loadStandingsHistory()
  })

  $: if (chartData && chartCanvas) {
    createChart()
  }

  async function fetchStandings() {
    try {
      loading = true
      error = ''

      // Fetch season info (non-critical)
      try {
        const seasonData = await api.get<{ season: { year: number; is_postseason: boolean } }>(`/api/seasons/${seasonId}`, session?.access_token)
        seasonYear = seasonData.season?.year
        isPostseason = seasonData.season?.is_postseason ?? false
      } catch {
        // Season info is supplementary
      }

      // Fetch standings and badges
      const [standingsData, badgesRes] = await Promise.all([
        api.get<{ standings: Standing[] }>(`/api/seasons/${seasonId}/standings`, session?.access_token),
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

  async function loadStandingsHistory() {
    try {
      const result = await api.get<{ history: StandingsHistoryEntry[] }>(`/api/seasons/${seasonId}/standings/history`, session?.access_token)

      if (!result.history || result.history.length === 0) {
        logData('STANDINGS_UNAVAILABLE', { seasonId, reason: 'no_history_data' })
        return
      }

      // Transform data for Chart.js
      const userMap = new Map<string, { username: string; data: number[] }>()
      const weeks = new Set<string>()

      result.history.forEach((entry: StandingsHistoryEntry) => {
        if (!userMap.has(entry.user_id)) {
          userMap.set(entry.user_id, {
            username: entry.username,
            data: []
          })
        }
        weeks.add(entry.week_id)
      })

      const sortedWeeks = Array.from(weeks)
      const weekMap = new Map<string, string>()
      result.history.forEach((entry: StandingsHistoryEntry) => {
        if (!weekMap.has(entry.week_id)) {
          weekMap.set(entry.week_id, entry.computed_at)
        }
      })
      sortedWeeks.sort((a, b) => {
        return new Date(weekMap.get(a)!).getTime() - new Date(weekMap.get(b)!).getTime()
      })

      const datasets: ChartDataset[] = []
      let colorIndex = 0

      userMap.forEach((userData, userId) => {
        const weeklyPoints = new Map<string, number>()

        result.history.forEach((entry: StandingsHistoryEntry) => {
          if (entry.user_id === userId) {
            weeklyPoints.set(entry.week_id, entry.points)
          }
        })

        const dataPoints = sortedWeeks.map(weekId => {
          return weeklyPoints.get(weekId) || 0
        })

        datasets.push({
          label: userData.username,
          data: dataPoints,
          borderColor: colors[colorIndex % colors.length],
          backgroundColor: colors[colorIndex % colors.length] + '33',
          tension: 0.1,
          fill: false,
          borderWidth: 2,
          pointRadius: 4,
          pointHoverRadius: 6
        })

        colorIndex++
      })

      const labels = sortedWeeks.map((_, index) => `Week ${index + 1}`)
      chartData = { labels, datasets }

    } catch (err) {
      logData('CHART_RENDER_FAILURE', {
        seasonId,
        error: err instanceof Error ? err.message : 'Unknown error'
      })
    }
  }

  function createChart() {
    if (chart) {
      chart.destroy()
    }

    const ctx = chartCanvas.getContext('2d')
    if (!ctx || !chartData) return
    chart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: chartData.labels,
        datasets: chartData.datasets
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: `Cumulative Points Over ${isPostseason ? 'Postseason' : 'Season'}`,
            font: {
              size: 20,
              weight: 'bold'
            },
            color: '#fff'
          },
          legend: {
            display: true,
            position: 'bottom',
            labels: {
              padding: 15,
              font: {
                size: 12
              },
              color: '#fff'
            }
          },
          tooltip: {
            mode: 'index',
            intersect: false,
            callbacks: {
              label: function(context) {
                return `${context.dataset.label}: ${context.parsed.y} points`
              }
            }
          }
        },
        scales: {
          y: {
            beginAtZero: true,
            title: {
              display: true,
              text: 'Total Points',
              font: {
                size: 14,
                weight: 'bold'
              },
              color: '#fff'
            },
            ticks: {
              font: {
                size: 12
              },
              color: '#fff'
            },
            grid: {
              color: 'rgba(255, 255, 255, 0.1)'
            }
          },
          x: {
            title: {
              display: true,
              text: 'Week',
              font: {
                size: 14,
                weight: 'bold'
              },
              color: '#fff'
            },
            ticks: {
              font: {
                size: 12
              },
              color: '#fff'
            },
            grid: {
              color: 'rgba(255, 255, 255, 0.1)'
            }
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    })
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
  <title>{seasonYear ? `${seasonYear} ${isPostseason ? 'Postseason' : 'Season'} Standings` : 'Season Standings'}</title>
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
        <div class="standings-container">
          <!-- Header -->
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">{isPostseason ? 'Postseason' : 'Season'} Standings</h1>
              <p class="text-sm opacity-half">{seasonYear ? `${seasonYear} ${isPostseason ? 'Postseason' : 'Season'}` : 'Current Season'}</p>
            </div>
            <div class="header-right"></div>
          </div>

          <!-- Leaderboard -->
          {#if standings.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No standings available yet. Standings will appear after the first week is completed.</p>
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
                      <a href="/users/{standing.user_id}" class="player-name">
                        {standing.username}
                        {#if isCurrentUser(standing.user_id)}
                          <span class="you-badge">You</span>
                        {/if}
                        {#if badgesByUser[standing.user_id]}
                          <BadgeList badges={badgesByUser[standing.user_id]} compact />
                        {/if}
                      </a>
                    </div>
                    <div class="points-col">
                      <span class="points-value">{standing.points}</span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>

            <!-- Chart -->
            {#if chartData}
              <div class="chart-container card">
                <canvas bind:this={chartCanvas}></canvas>
              </div>
            {/if}
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

  .leaderboard-card {
    margin-bottom: 30px;
  }

  .chart-container {
    position: relative;
    height: 500px;
    width: 100%;
    padding: 1rem;
  }

  @media only screen and (max-width: 45em) {
    .chart-container {
      height: 400px;
    }
  }
</style>