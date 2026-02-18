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

  let seasonId = ''
  let weekId = ''

  let totalGames = 0

  type UserPickSummary = {
    user_id: string
    username: string
    picks_submitted: number
  }

  let users: UserPickSummary[] = []

  onMount(() => {
    seasonId = $page.params.seasonid ?? ''
    weekId = $page.params.weekid ?? ''

    if (!seasonId || !weekId) {
      error = 'Missing season or week ID'
      return
    }

    fetchPickSummary()
  })

  async function fetchPickSummary() {
    try {
      loading = true
      error = ''

      const json = await api.get<{ users: UserPickSummary[]; total_games: number }>(
        `/api/commissioner/weeks/${weekId}/picks`,
        session?.access_token
      )

      users = json.users || []
      totalGames = json.total_games || 0
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch pick summary')
    } finally {
      loading = false
    }
  }

  function isComplete(user: UserPickSummary): boolean {
    return totalGames > 0 && user.picks_submitted === totalGames
  }
</script>

<svelte:head>
  <title>Week Pick Summary</title>
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
        <div class="summary-container">
          <div class="page-header-three-col">
            <div class="header-left">
              <a href="/commissioner/seasons/{seasonId}/weeks/{weekId}" class="button">← Back</a>
            </div>
            <div class="header-title-group">
              <h1 class="header-title">User Pick Summary</h1>
              <p class="text-sm opacity-half">Total games this week: {totalGames}</p>
            </div>
            <div class="header-right"></div>
          </div>

          {#if users.length === 0}
            <div class="card">
              <p class="text-sm opacity-half">No users returned.</p>
            </div>
          {:else}
            <div class="card">
              <table class="summary-table">
                <thead>
                  <tr>
                    <th>User</th>
                    <th>Picks Submitted</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {#each users as user}
                    <tr>
                      <td>{user.username}</td>
                      <td>{user.picks_submitted} / {totalGames}</td>
                      <td>
                        {#if isComplete(user)}
                          Complete
                        {:else}
                          Incomplete
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      {/if}

    </div>
  </div>
</div>

<style>
  .summary-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .summary-table {
    width: 100%;
    border-collapse: collapse;
  }

  .summary-table th,
  .summary-table td {
    text-align: left;
    padding: 12px 10px;
    border-bottom: 1px solid var(--custom-border-color, #333);
  }

  .summary-table th {
    text-transform: uppercase;
    font-size: 0.75rem;
    letter-spacing: 0.5px;
    opacity: 0.8;
  }

  .summary-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }
</style>
