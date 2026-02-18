<script lang="ts">
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = true
  let saving = false
  let error = ''

  let settings = {
    id: '',
    pick_cutoff_minutes: 0,
    allow_pick_edits: false,
    points_per_correct_pick: 0,
    competition_timezone: '',
    allow_commissioner_overrides: false
  }

  onMount(fetchSettings)

  async function fetchSettings() {
    loading = true
    error = ''

    try {
      const json = await api.get<{ settings: typeof settings }>('/api/settings', session?.access_token)
      settings = json.settings
    } catch (err) {
      error = handleApiError(err, 'Failed to load settings')
    } finally {
      loading = false
    }
  }

  async function submit() {
    saving = true

    try {
      const json = await api.put<{ settings: typeof settings }>('/api/admin/settings', settings, session?.access_token)
      settings = json.settings
      toasts.success('Settings saved successfully!')
    } catch (err) {
      // Just show toast for action errors
      handleApiError(err, 'Failed to update settings')
    } finally {
      saving = false
    }
  }
</script>

<svelte:head>
  <title>Admin Settings</title>
</svelte:head>

<div class="settings-container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/admin" class="button">← Back</a>
    </div>
    <h1 class="header-title">Global Settings</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {:else}
    {#if error}
      <div class="card error-card">
        <p class="error">{error}</p>
      </div>
    {/if}

    <form on:submit|preventDefault={submit}>
      <div class="card">
        <h2>Competition Settings</h2>

        <div class="form-group">
          <label for="pick_cutoff_minutes">Pick Cutoff Minutes</label>
          <input
            id="pick_cutoff_minutes"
            type="number"
            bind:value={settings.pick_cutoff_minutes}
            placeholder="e.g., 60"
          />
          <p class="help-text">Minutes before game start when picks are locked</p>
        </div>

        <div class="form-group">
          <label for="points_per_correct_pick">Points Per Correct Pick</label>
          <input
            id="points_per_correct_pick"
            type="number"
            bind:value={settings.points_per_correct_pick}
            placeholder="e.g., 1"
          />
        </div>

        <div class="form-group">
          <label for="competition_timezone">Competition Timezone</label>
          <input
            id="competition_timezone"
            type="text"
            bind:value={settings.competition_timezone}
            placeholder="e.g., America/New_York"
          />
        </div>

        <div class="form-group checkbox">
          <label>
            <input
              type="checkbox"
              bind:checked={settings.allow_pick_edits}
            />
            Allow pick edits
          </label>
        </div>

        <div class="form-group checkbox">
          <label>
            <input
              type="checkbox"
              bind:checked={settings.allow_commissioner_overrides}
            />
            Allow commissioner overrides
          </label>
        </div>

        <button type="submit" class="button primary" disabled={saving}>
          {saving ? 'Saving…' : 'Save Settings'}
        </button>
      </div>
    </form>
  {/if}
</div>

<style>
  .settings-container {
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

  .form-group input[type="text"],
  .form-group input[type="number"] {
    width: 100%;
    padding: 12px;
    font-size: 1rem;
  }

  .form-group.checkbox label {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
  }

  .form-group.checkbox input[type="checkbox"] {
    width: auto;
  }

  .help-text {
    margin: 8px 0 0 0;
    font-size: 0.85rem;
    opacity: 0.7;
  }

  button[type="submit"] {
    width: 100%;
    padding: 14px;
    font-size: 1rem;
  }
</style>
