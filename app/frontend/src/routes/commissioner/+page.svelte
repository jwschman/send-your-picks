<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'

  export let data: PageData
  let { session, supabase } = data
  $: ({ session, supabase } = data)

  let loadingSeason = true
  let error = ''
  let currentSeasonId: string | null = null
  let currentSeasonYear: number | null = null
  let currentSeasonIsPostseason: boolean = false

  onMount(async () => {
    await fetchActiveSeason()
  })

  async function fetchActiveSeason() {
    try {
      loadingSeason = true
      const data = await api.get<{ id: string; year: number; is_postseason: boolean }>('/api/seasons/active', session?.access_token)
      currentSeasonId = data?.id || null
      currentSeasonYear = data?.year || null
      currentSeasonIsPostseason = data?.is_postseason ?? false
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        currentSeasonId = null
        currentSeasonYear = null
        return
      }
      error = handleApiError(err, 'Failed to fetch active season')
      currentSeasonId = null
      currentSeasonYear = null
    } finally {
      loadingSeason = false
    }
  }

  async function handleSignOut() {
    await supabase.auth.signOut()
    goto('/login')
  }
</script>

<svelte:head>
  <title>Commissioner Dashboard</title>
</svelte:head>

<div class="dashboard">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/dashboard" class="button">← Back</a>
    </div>
    <h1 class="header-title">Commissioner Dashboard</h1>
    <div class="header-right"></div>
  </div>

  {#if loadingSeason}
    <div class="card">
      <Spinner />
    </div>
  {:else}
    {#if error}
      <div class="card error-card">
        <p class="error">Error: {error}</p>
      </div>
    {/if}

    {#if currentSeasonId && currentSeasonYear}
      <div class="card current-season">
        <h2>Current Active Season: {currentSeasonYear}{currentSeasonIsPostseason ? ' Postseason' : ''}</h2>
        <div class="season-links">
          <a href={`/commissioner/seasons/${currentSeasonId}`} class="button block primary">
            View Current Season
          </a>
        </div>
      </div>
    {:else}
      <div class="card">
        <p class="text-sm opacity-half">No active season found.</p>
      </div>
    {/if}

    <div class="nav-links">
      <h2>Other Links</h2>
      <a href="/commissioner/seasons/new" class="button block">Create New Season</a>
      <a href="/commissioner/seasons" class="button block">View All Seasons</a>
    </div>
  {/if}
</div>

<style>
  .dashboard {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .current-season h2 {
    margin-top: 0;
    margin-bottom: 15px;
  }

  .season-links {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
</style>