<script lang="ts">
  import { env } from '$env/dynamic/public'
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { logUser } from '$lib/utils/logger'

  export let data: PageData
  let { session, supabase } = data
  $: ({ session, supabase } = data)

  type WhoAmIProfile = {
    id: string
    username: string | null
    email: string
    role: string
    avatar_url: string | null
  }

  let loading = true
  let profile: WhoAmIProfile | null = null
  let error = ''

  onMount(async () => {
    await fetchProfile()
  })

  async function fetchProfile() {
    try {
      loading = true
      error = ''

      const token = session?.access_token
      if (!token) {
        error = 'No session token'
        logUser('PROFILE_FETCH_FAILURE', { error: 'No session token', hasSession: false })
        return
      }

      const response = await fetch(`${env.PUBLIC_API_BASE_URL}/api/whoami`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || `API returned ${response.status}`)
      }

      profile = await response.json()
      logUser('PROFILE_FETCH_SUCCESS', { hasSession: true })
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error'
      error = errorMessage
      logUser('PROFILE_FETCH_FAILURE', { error: errorMessage, hasSession: !!session })
    } finally {
      loading = false
    }
  }

  function getRoleBadgeClass(role: string): string {
    if (role === 'admin') return 'admin'
    if (role === 'commissioner') return 'commissioner'
    return 'user'
  }
</script>

<svelte:head>
  <title>Who Am I?</title>
</svelte:head>

<div class="container">
  <div class="row">
    <div class="col-12">
      <div class="page-container">
        <div class="page-header-three-col">
          <div class="header-left">
            <a href="/dashboard" class="button">← Back</a>
          </div>
          <h1 class="header-title">Who Am I?</h1>
          <div class="header-right"></div>
        </div>

        {#if loading}
          <div class="card">
            <Spinner />
          </div>
        {/if}

        {#if error}
          <div class="card error-card">
            <p class="error">{error}</p>
            <div style="margin-top: 15px;">
              <button class="button" on:click={fetchProfile}>Retry</button>
            </div>
          </div>
        {/if}

        {#if profile && !loading}
          <div class="profile-container">
            <div class="card info-card">
              <h2>Your Information</h2>

              {#if profile.avatar_url}
                <div class="info-row">
                  <span class="info-label">Avatar:</span>
                  <span class="info-value">
                    <img
                      src={profile.avatar_url}
                      alt="Avatar"
                      style="width: 64px; height: 64px; border-radius: 50%; object-fit: cover;"
                    />
                  </span>
                </div>
              {/if}

              <div class="info-row">
                <span class="info-label">User ID:</span>
                <span class="info-value" style="font-family: monospace; font-size: 0.85rem;">
                  {profile.id}
                </span>
              </div>

              {#if profile.username}
                <div class="info-row">
                  <span class="info-label">Username:</span>
                  <span class="info-value">{profile.username}</span>
                </div>
              {/if}

              <div class="info-row">
                <span class="info-label">Email:</span>
                <span class="info-value">{profile.email}</span>
              </div>

              <div class="info-row">
                <span class="info-label">Role:</span>
                <span class="info-value">
                  <span class="role-badge {getRoleBadgeClass(profile.role)}">
                    {profile.role}
                  </span>
                </span>
              </div>

            </div>

            <div class="actions">
              <a href="/account" class="button primary block">Edit Account</a>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .page-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .profile-container {
    max-width: 1000px;
  }
</style>
