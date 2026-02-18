<script lang="ts">
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'

  export let data: PageData

  let { session, supabase } = data
  $: ({ session, supabase } = data)

  let loading = false
  let year = new Date().getFullYear()
  let numberOfWeeks = 18
  let isPostseason = false
  let error = ''

  async function createSeason() {
    try {
      loading = true

      const result = await api.post<{ year: number }>('/api/commissioner/seasons', { year, number_of_weeks: numberOfWeeks, is_postseason: isPostseason }, session?.access_token)
      toasts.success(`Season ${result.year} created successfully!`)
      goto('/commissioner')

    } catch (err) {
      // Just show toast for action errors
      handleApiError(err, 'Failed to create season')
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Create New Season</title>
</svelte:head>

<div class="form-widget">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/commissioner" class="button">← Back</a>
    </div>
    <h1 class="header-title">Create New Season</h1>
    <div class="header-right"></div>
  </div>

  {#if error}
    <div class="card error-card">
      <p class="error">{error}</p>
    </div>
  {/if}

  <form on:submit|preventDefault={createSeason}>
    <div class="card">
      <h2>Season Details</h2>

      <div class="form-group">
        <label for="year">Season Year</label>
        <input
          id="year"
          type="number"
          bind:value={year}
          min="2000"
          max="2999"
          required
          placeholder="e.g., 2025"
        />
      </div>

      <div class="form-group">
        <label for="weeks">Number of Weeks</label>
        <input
          id="weeks"
          type="number"
          bind:value={numberOfWeeks}
          min="1"
          max="22"
          required
        />
        <p class="hint">NFL regular season is currently 18 weeks</p>
      </div>

      <div class="form-group">
        <label class="checkbox-label">
          <input
            type="checkbox"
            bind:checked={isPostseason}
          />
          Postseason
        </label>
        <p class="hint">Week numbers will map to postseason weeks (1-5)</p>
      </div>

      <button type="submit" class="button primary block" disabled={loading}>
        {loading ? 'Creating...' : 'Create Season'}
      </button>
    </div>
  </form>
</div>

<style>
  .form-widget {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .card h2 {
    font-size: 1.1rem;
    margin: 0 0 20px 0;
    padding-bottom: 10px;
    border-bottom: var(--custom-border);
  }

  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: 500;
  }

  .form-group input {
    width: 100%;
    padding: 12px;
    font-size: 1rem;
  }

  .hint {
    margin: 6px 0 0;
    font-size: 0.85rem;
    color: var(--custom-color-secondary, #666);
  }

  button[type="submit"] {
    padding: 14px;
    font-size: 1rem;
  }
</style>