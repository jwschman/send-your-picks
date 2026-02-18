<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import type { Season } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = true
  let error = ''
  let seasonId = ''
  let season: Season | null = null

  // Action loading state for activate/deactivate toggle
  let toggling = false

  // Week count inline editing
  let editingWeeks = false
  let editedWeekCount = 0
  let savingWeeks = false

  onMount(async () => {
    seasonId = $page.params.seasonid ?? ''
    if (!seasonId) {
      error = 'No season ID provided'
      loading = false
      return
    }

    await fetchSeason()
  })

  async function fetchSeason() {
    try {
      loading = true
      error = ''

      const res = await api.get<{ season: Season }>(`/api/seasons/${seasonId}`, session?.access_token)
      season = res.season
      editedWeekCount = season.number_of_weeks
    } catch (err) {
      error = handleApiError(err, 'Failed to load season')
    } finally {
      loading = false
    }
  }

  // Toggles the season between active and inactive.
  // Activate will fail on the backend if another season is already active,
  // and the error message will be shown via the toast.
  async function toggleSeasonActive() {
    if (!season) return

    const action = season.is_active ? 'deactivate' : 'activate'

    try {
      toggling = true
      await api.patch(`/api/commissioner/seasons/${seasonId}/${action}`, {}, session?.access_token)
      toasts.success(`Season ${action}d successfully!`)
      await fetchSeason()
    } catch (err) {
      handleApiError(err, `Failed to ${action} season`)
    } finally {
      toggling = false
    }
  }

  function startEditingWeeks() {
    if (!season) return
    editedWeekCount = season.number_of_weeks
    editingWeeks = true
  }

  function cancelEditingWeeks() {
    editingWeeks = false
  }

  // Saves the updated number of weeks.
  // Backend validates it can't go below the number of weeks already created.
  async function saveWeekCount() {
    if (!season || editedWeekCount === season.number_of_weeks) {
      editingWeeks = false
      return
    }

    try {
      savingWeeks = true
      await api.patch(`/api/commissioner/seasons/${seasonId}/weeks-count`, {
        number_of_weeks: editedWeekCount
      }, session?.access_token)
      toasts.success('Number of weeks updated!')
      editingWeeks = false
      await fetchSeason()
    } catch (err) {
      handleApiError(err, 'Failed to update number of weeks')
    } finally {
      savingWeeks = false
    }
  }
</script>

<svelte:head>
  <title>Manage Season{season ? ` - ${season.year} ${season.is_postseason ? 'Postseason' : 'Season'}` : ''}</title>
</svelte:head>

<div class="container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/commissioner/seasons/{seasonId}" class="button">← Back</a>
    </div>
    <h1 class="header-title">Manage Season</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {:else if error}
    <div class="card error-card">
      <p class="error">{error}</p>
    </div>
  {:else if season}
    <!-- Season Status Section -->
    <div class="section">
      <h2>Season Status</h2>
      <div class="card">
        <div class="status-row">
          <div class="status-info">
            <span class="status-label">Current Status:</span>
            <span class="status-badge {season.is_active ? 'active' : 'inactive'}">
              {season.is_active ? 'Active' : 'Inactive'}
            </span>
          </div>

          <button
            class="button {season.is_active ? 'danger' : 'primary'}"
            on:click={toggleSeasonActive}
            disabled={toggling}
          >
            {#if toggling}
              {season.is_active ? 'Deactivating...' : 'Activating...'}
            {:else}
              {season.is_active ? 'Deactivate Season' : 'Activate Season'}
            {/if}
          </button>
        </div>
      </div>
    </div>

    <!-- Season Settings Section -->
    <div class="section">
      <h2>Season Settings</h2>
      <div class="card">
        <div class="setting-row">
          <span class="setting-label">Number of Weeks</span>
          <div class="setting-value">
            {#if editingWeeks}
              <!-- Edit mode: input with save/cancel -->
              <input
                type="number"
                min="1"
                max="22"
                bind:value={editedWeekCount}
                class="week-input"
              />
              <button
                class="button primary small"
                on:click={saveWeekCount}
                disabled={savingWeeks}
              >
                {savingWeeks ? 'Saving...' : 'Save'}
              </button>
              <button
                class="button small"
                on:click={cancelEditingWeeks}
                disabled={savingWeeks}
              >
                Cancel
              </button>
            {:else}
              <!-- Display mode: value with edit button -->
              <span class="week-count-display">{season.number_of_weeks}</span>
              <button class="button small" on:click={startEditingWeeks}>Edit</button>
            {/if}
          </div>
        </div>
      </div>
    </div>

    <!-- Management Links Section -->
    <div class="section">
      <h2>Management</h2>
      <div class="card">
        <div class="management-links">
          <a href="/commissioner/seasons/{seasonId}/participants" class="management-link">
            <div class="link-content">
              <span class="link-title">Manage Participants</span>
              <span class="link-description text-sm opacity-half">Add or remove users from this season</span>
            </div>
            <span class="link-arrow">→</span>
          </a>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .container {
    max-width: 800px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .section {
    margin-bottom: 25px;
  }

  .section h2 {
    margin: 0 0 15px 0;
  }

  .status-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    flex-wrap: wrap;
  }

  .status-info {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .status-label {
    font-weight: 500;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .status-badge.active {
    background: rgba(34, 197, 94, 0.2);
    color: #22c55e;
  }

  .status-badge.inactive {
    background: rgba(156, 163, 175, 0.2);
    color: #9ca3af;
  }

  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 15px;
    flex-wrap: wrap;
  }

  .setting-label {
    font-weight: 500;
  }

  .setting-value {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .week-count-display {
    font-size: 0.95rem;
    min-width: 30px;
    text-align: center;
  }

  .week-input {
    width: 70px;
    text-align: center;
  }

  .management-links {
    display: flex;
    flex-direction: column;
  }

  .management-link {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 15px;
    text-decoration: none;
    color: inherit;
    border-radius: 8px;
    transition: background-color 0.15s ease;
  }

  .management-link:hover {
    background: var(--custom-bg-secondary, rgba(0, 0, 0, 0.05));
  }

  .link-content {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .link-title {
    font-weight: 500;
  }

  .link-arrow {
    font-size: 1.2rem;
    color: var(--custom-color-secondary, #666);
  }

  .button.danger {
    background: var(--custom-color-error, #dc3545);
    color: white;
  }

  .button.danger:hover:not(:disabled) {
    background: #c82333;
  }

  .button.danger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
