<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api, ApiError } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import { getAvatarInitial } from '$lib/utils/formatters'
  import type { PublicProfile, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let user: PublicProfile | null = null
  let userBadges: Badge[] = []
  let userId: string = ''

  onMount(() => {
    userId = $page.params.userid ?? ''
    if (!userId) {
      error = 'No user ID provided'
      return
    }

    fetchUser()
  })

  async function fetchUser() {
    try {
      loading = true
      error = ''

      const [responseData, badgesRes] = await Promise.all([
        api.get<{ user: PublicProfile }>(`/api/users/${userId}`, session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token).catch((): { badges: Record<string, Badge[]> } => ({ badges: {} }))
      ])
      user = responseData.user
      userBadges = badgesRes.badges[userId] || []
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        error = 'User not found'
      } else {
        error = handleApiError(err, 'Failed to fetch user')
      }
    } finally {
      loading = false
    }
  }


</script>

<svelte:head>
  <title>{user?.username || 'User'} - Profile</title>
</svelte:head>

<div class="profile-container">
  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {/if}

  {#if error}
    <div class="card error-card">
      <p class="error">{error}</p>
      <div class="actions">
        <a href="/users" class="button">← Back to Users</a>
      </div>
    </div>
  {/if}

  {#if user && !loading}
    <div class="page-header-three-col">
      <div class="header-left">
        <a href="/users" class="button">← Back</a>
      </div>
      <h1 class="header-title">{user.username || 'User Profile'}</h1>
      <div class="header-right">
        <span class="role-badge {user.role}">{user.role}</span>
      </div>
    </div>

    <div class="profile-header card">
      <div class="avatar-section">
        {#if user.avatar_url}
          <img src={user.avatar_url} alt={user.username || 'User'} class="avatar image large" />
        {:else}
          <div class="avatarPlaceholder large">
            {getAvatarInitial(user.username)}
          </div>
        {/if}
      </div>

      <div class="user-details">
        {#if user.tagline}
          <p class="tagline">"{user.tagline}"</p>
        {/if}
        {#if userBadges.length > 0}
          <div class="profile-badges">
            <BadgeList badges={userBadges} />
          </div>
        {/if}
      </div>
    </div>

    <div class="profile-info card">
      <h2>Profile Information</h2>
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">Username</span>
          <div class="info-value">{user.username || 'Not set'}</div>
        </div>

        <div class="info-item">
          <span class="info-label">Role</span>
          <div class="info-value">
            <span class="role-badge {user.role}">{user.role}</span>
          </div>
        </div>

        {#if user.tagline}
          <div class="info-item">
            <span class="info-label">Tagline</span>
            <div class="info-value">{user.tagline}</div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .profile-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .profile-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    padding: 30px;
    margin-bottom: 20px;
  }

  .avatar-section {
    flex-shrink: 0;
  }

  .avatar.image.large {
    width: 120px;
    height: 120px;
    object-fit: cover;
    border-radius: 50%;
  }

  .avatarPlaceholder.large {
    width: 120px;
    height: 120px;
    font-size: 3rem;
  }

  .user-details {
    flex: 1;
    text-align: center;
  }

  .tagline {
    color: rgba(255, 255, 255, 0.7);
    font-style: italic;
    font-size: 1rem;
  }

  .profile-badges {
    margin-top: 12px;
  }

  .profile-info {
    padding: 25px;
  }

  .profile-info h2 {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .info-grid {
    display: grid;
    gap: 20px;
  }

  .info-item .info-label {
    display: block;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: rgba(255, 255, 255, 0.5);
    margin-bottom: 6px;
  }

  .info-value {
    font-size: 1rem;
    color: var(--custom-color);
  }

  @media only screen and (min-width: 45em) {
    .profile-header {
      flex-direction: row;
      align-items: flex-start;
    }

    .user-details {
      text-align: left;
    }
  }
</style>