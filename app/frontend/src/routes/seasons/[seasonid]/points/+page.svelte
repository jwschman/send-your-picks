<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let seasonId: string = ''

  type WeekPoints = {
    week_number: number
    week_points: number
    week_rank: number | null
    total_points: number | null
    league_rank: number | null
  }

  let weeks: WeekPoints[] = []

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''

    if (!seasonId) {
      error = 'Missing season ID'
      return
    }

    fetchPoints()
  })

  async function fetchPoints() {
    try {
      loading = true
      error = ''

      const json = await api.get<{ weeks_with_points: WeekPoints[] }>(`/api/seasons/${seasonId}/points`, session?.access_token)
      weeks = json.weeks_with_points ?? []
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch points')
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Season Points</title>
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

      {#if !loading && !error}
        <div class="points-container">
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/seasons/{seasonId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">Points Summary</h1>
              <p class="text-sm opacity-half">Your weekly performance breakdown</p>
            </div>
            <div class="header-right"></div>
          </div>

          {#if weeks.length > 0}
            <div class="card points-card">
              <table class="points-table">
                <thead>
                  <tr>
                    <th class="week-col">Week</th>
                    <th class="points-col">Week Points</th>
                    <th class="rank-col">Week Rank</th>
                    <th class="points-col">Total Points</th>
                    <th class="rank-col">League Rank</th>
                  </tr>
                </thead>
                <tbody>
                  {#each weeks as w}
                    <tr>
                      <td class="week-col">
                        <span class="week-badge">Week {w.week_number}</span>
                      </td>
                      <td class="points-col">
                        <span class="points-value">{w.week_points}</span>
                      </td>
                      <td class="rank-col">
                        {#if w.week_rank}
                          <span class="rank-badge rank-{w.week_rank}">#{w.week_rank}</span>
                        {:else}
                          <span class="text-sm opacity-half">-</span>
                        {/if}
                      </td>
                      <td class="points-col">
                        {#if w.total_points !== null}
                          <span class="total-points">{w.total_points}</span>
                        {:else}
                          <span class="text-sm opacity-half">-</span>
                        {/if}
                      </td>
                      <td class="rank-col">
                        {#if w.league_rank}
                          <span class="rank-badge rank-{w.league_rank}">#{w.league_rank}</span>
                        {:else}
                          <span class="text-sm opacity-half">-</span>
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {:else}
            <div class="card">
              <p class="text-sm opacity-half">No points data available yet.</p>
            </div>
          {/if}
        </div>
      {/if}

    </div>
  </div>
</div>

<style>
  .points-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .points-card {
    padding: 0;
    overflow-x: auto;
  }

  .points-table {
    width: 100%;
    border-collapse: collapse;
  }

  .points-table th {
    padding: 16px 12px;
    text-align: center;
    border-bottom: 2px solid rgba(255, 255, 255, 0.1);
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    background: rgba(255, 255, 255, 0.02);
  }

  .points-table td {
    padding: 16px 12px;
    text-align: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .points-table tbody tr {
    transition: background-color 0.2s ease;
  }

  .points-table tbody tr:hover {
    background-color: rgba(255, 255, 255, 0.03);
  }

  .week-badge {
    display: inline-block;
    padding: 4px 12px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .points-value {
    font-size: 1.2rem;
    font-weight: 700;
    color: var(--custom-color-brand);
  }

  .total-points {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--custom-color);
  }

  .rank-badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
  }

  /* Top 3 rankings get special colors */
  .rank-badge.rank-1 {
    background: linear-gradient(135deg, #FFD700, #FFA500);
    color: #0a0a0a;
  }

  .rank-badge.rank-2 {
    background: linear-gradient(135deg, #C0C0C0, #A8A8A8);
    color: #0a0a0a;
  }

  .rank-badge.rank-3 {
    background: linear-gradient(135deg, #CD7F32, #8B4513);
    color: #fff;
  }

  /* All other rankings */
  .rank-badge:not(.rank-1):not(.rank-2):not(.rank-3) {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.8);
  }

  @media only screen and (max-width: 45em) {
    .points-table th,
    .points-table td {
      padding: 12px 8px;
      font-size: 0.85rem;
    }

    .week-badge {
      font-size: 0.75rem;
      padding: 3px 8px;
    }

    .points-value {
      font-size: 1rem;
    }

    .total-points {
      font-size: 0.95rem;
    }

    .rank-badge {
      font-size: 0.7rem;
      padding: 3px 8px;
    }
  }
</style>
