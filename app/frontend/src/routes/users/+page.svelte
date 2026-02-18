<script lang="ts">
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import BadgeList from '$lib/components/BadgeList.svelte'
  import { getAvatarInitial } from '$lib/utils/formatters'
  import type { PublicProfile, Badge } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let users: PublicProfile[] = []
  let badgesByUser: Record<string, Badge[]> = {}

  onMount(() => {
    fetchData()
  })

  async function fetchData() {
    try {
      loading = true
      error = ''

      const [usersRes, badgesRes] = await Promise.all([
        api.get<{ users: PublicProfile[] }>('/api/users', session?.access_token),
        api.get<{ badges: Record<string, Badge[]> }>('/api/badges', session?.access_token)
      ])

      users = usersRes.users
      badgesByUser = badgesRes.badges
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch users')
    } finally {
      loading = false
    }
  }


</script>

<svelte:head>
  <title>Users</title>
</svelte:head>

<div class="users-container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/dashboard" class="button">← Back</a>
    </div>
    <h1 class="header-title">Users</h1>
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
    </div>
  {/if}

  {#if !loading && users.length === 0}
    <div class="card">
      <p class="text-sm opacity-half">No users found.</p>
    </div>
  {/if}

  {#if users.length > 0}
    <div class="user-grid">
      {#each users as user}
        <a href="/users/{user.id}" class="user-card card">
          <div class="avatar-container">
            {#if user.avatar_url}
              <img src={user.avatar_url} alt={user.username || 'User'} class="avatar image" />
            {:else}
              <div class="avatarPlaceholder">
                {getAvatarInitial(user.username)}
              </div>
            {/if}
          </div>
          <div class="user-info">
            <div class="username">{user.username || 'Anonymous'}</div>
            {#if user.tagline}
              <div class="tagline text-sm opacity-half">{user.tagline}</div>
            {/if}
            {#if badgesByUser[user.id]}
              <div class="badge-list-wrapper">
                <BadgeList badges={badgesByUser[user.id]} />
              </div>
            {/if}
            <span class="role-badge {user.role}">{user.role}</span>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>

<style>
  .users-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .user-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 20px;
  }

  .user-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px;
    padding: 25px 20px;
    text-decoration: none;
    transition: all 0.2s ease;
  }

  .user-card:hover {
    transform: translateY(-2px);
    border-color: var(--custom-color-brand);
  }

  .avatar-container {
    width: 80px;
    height: 80px;
  }

  .avatar.image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
  }

  .avatarPlaceholder {
    width: 80px;
    height: 80px;
    font-size: 2rem;
  }

  .user-info {
    text-align: center;
    width: 100%;
  }

  .username {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 5px;
    color: var(--custom-color);
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tagline {
    margin-bottom: 10px;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .badge-list-wrapper {
    display: flex;
    justify-content: center;
    margin-bottom: 10px;
  }

  @media only screen and (max-width: 45em) {
    .user-grid {
      grid-template-columns: 1fr;
    }
  }
</style>