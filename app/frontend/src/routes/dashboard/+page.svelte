<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'
  import LoadingCard from '$lib/components/LoadingCard.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { getRoleFromToken } from '$lib/utils/auth'
  import { handleApiError } from '$lib/utils/errors'
  import { logData } from '$lib/utils/logger'

  export let data: PageData
  let { session, supabase } = data
  $: ({ session, supabase } = data)

  let loading = true
  let error = ''
  let userRole = ''

  let activeSeason: { id: string; year: number; is_postseason: boolean } | null = null
  let activeWeek: { id: string; number: number } | null = null

  let picksSubmitted = 0
  let totalGames = 0

  let myStandings: {
    points: number
    rank: number | null
    total_users: number
  } | null = null

  onMount(async () => {
    userRole = getRoleFromToken(session?.access_token) || ''

    await fetchActiveSeason()

    if (activeSeason) {
      await Promise.all([
        fetchActiveWeekAndSummary(),
        fetchMyStandings()
      ])
    }

    loading = false
  })

  async function fetchActiveSeason() {
    try {
      activeSeason = await api.get<{ id: string; year: number; is_postseason: boolean }>('/api/seasons/active', session?.access_token)
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        activeSeason = null
        return
      }
      throw err
    }
  }

  async function fetchActiveWeekAndSummary() {
    try {
      activeWeek = await api.get<{ id: string; number: number }>(
        `/api/seasons/${activeSeason!.id}/weeks/active`,
        session?.access_token
      )
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        activeWeek = null
        return
      }
      throw err
    }

    const summaryData = await api.get<{ summary: { picks_completed: number; total_games: number } }>(
      `/api/weeks/${activeWeek.id}/picks/summary`,
      session?.access_token
    )
    picksSubmitted = summaryData.summary.picks_completed
    totalGames = summaryData.summary.total_games
  }

  async function fetchMyStandings() {
    try {
      const data = await api.get<{ my_standings: { points: number; rank: number | null; total_users: number } }>(
        `/api/seasons/${activeSeason!.id}/standings/me`,
        session?.access_token
      )
      myStandings = data.my_standings
      logData('STANDINGS_LOAD_SUCCESS', { seasonId: activeSeason!.id })
    } catch (err) {
      // Standings are supplementary - don't fail the whole page
      logData('STANDINGS_UNAVAILABLE', {
        seasonId: activeSeason?.id,
        reason: 'not_yet_available'
      })
    }
  }

  async function handleSignOut() {
    await supabase.auth.signOut()
    goto('/login')
  }
</script>

<svelte:head>
  <title>Dashboard</title>
</svelte:head>

<div class="dashboard">
  <div class="page-header-three-col">
    <div class="header-left"></div>
    <h1 class="header-title">Dashboard</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <LoadingCard lines={4} />
    <LoadingCard lines={3} />
  {:else if error}
    <p class="error">{error}</p>
  {:else}

    {#if activeSeason && activeWeek}
      <a
        href={`/seasons/${activeSeason.id}/weeks/${activeWeek.id}`}
        class="card info-card card-link"
      >
        <h2>Week {activeWeek.number}</h2>

        <div class="info-row">
          <span class="info-label">Season</span>
          <span class="info-value">{activeSeason.year}{activeSeason.is_postseason ? ' Postseason' : ''}</span>
        </div>

        <div class="info-row">
          <span class="info-label">Your Picks</span>
          <span class="info-value">
            {picksSubmitted} / {totalGames}
          </span>
        </div>

        <div class="week-card-hint">
          Click to view the week overview and make picks
        </div>
      </a>
    {/if}

    {#if activeSeason && myStandings}
      <div class="card info-card standings-card">
        <h2>{activeSeason.is_postseason ? 'Postseason' : 'Season'} Standings</h2>

        <a href={`/seasons/${activeSeason.id}/standings`} class="info-row info-row-link">
          <span class="info-label">Current Rank</span>
          <span class="info-value">
            {myStandings.rank ? `${myStandings.rank} / ${myStandings.total_users}` : 'Unranked'}
          </span>
        </a>

        <a href={`/seasons/${activeSeason.id}/points`} class="info-row info-row-link">
          <span class="info-label">Total Points</span>
          <span class="info-value">{myStandings.points}</span>
        </a>

        <a href={`/seasons/${activeSeason.id}/winners`} class="info-row info-row-link">
          <span class="info-label">Weekly Winners</span>
          <span class="info-value">View →</span>
        </a>
      </div>
    {/if}

    <div class="nav-links">
      <h2>Quick Links</h2>

      <a href="/account" class="button block">Manage My Account</a>
      <a href="/users" class="button block">View All Users</a>

      <br>

      <a href="/seasons" class="button block">View All Seasons</a>

      {#if activeSeason}
        <a href={`/seasons/${activeSeason.id}`} class="button block">
          View Current Season ({activeSeason.year}{activeSeason.is_postseason ? ' Postseason' : ''})
        </a>
      {:else}
        <div class="button block disabled">No Active Season</div>
      {/if}

      {#if userRole === 'commissioner' || userRole === 'admin'}
        <br>
        <a href="/commissioner" class="button block">Commissioner Dashboard</a>
      {/if}

      {#if userRole === 'admin'}
        <br>
        <a href="/admin" class="button block">Admin Dashboard</a>
      {/if}

      <br>
      <button class="button block" on:click={handleSignOut}>Sign Out</button>
    </div>
  {/if}
</div>

<style>
  .dashboard {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .week-card-hint {
    margin-top: 12px;
    font-size: 0.85rem;
    color: var(--custom-color-secondary);
  }

  .standings-card {
    margin-top: 20px;
  }

  .info-row-link {
    text-decoration: none;
    cursor: pointer;
    border-radius: var(--custom-border-radius);
    margin: 0 -12px;
    padding: 10px 12px;
    border-left: 3px solid transparent;
    transition: all 0.2s ease;
  }

  .info-row-link:hover {
    background-color: rgba(255, 255, 255, 0.05);
    border-left-color: var(--custom-color-brand);
    transform: translateX(4px);
  }

  .info-row-link:hover .info-value {
    color: var(--custom-color-brand);
  }
</style>
