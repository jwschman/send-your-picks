<script lang="ts">
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { formatLastSignIn } from '$lib/utils/formatters'
  import type { User } from '$lib/types'

  export let data: PageData

  let { session } = data
  $: ({ session } = data)

  let loading = false
  let error = ''

  let users: User[] = []

  onMount(() => {
    fetchUsers()
  })

  async function fetchUsers() {
    try {
      loading = true
      error = ''

      const res = await api.get<{ users: User[] }>('/api/admin/users', session?.access_token)
      users = res.users
    } catch (err) {
      error = handleApiError(err, 'Failed to fetch users')
    } finally {
      loading = false
    }
  }


</script>

<svelte:head>
  <title>Admin – Users</title>
</svelte:head>

<div class="users-container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/admin" class="button">← Back</a>
    </div>
    <h1 class="header-title">All Users</h1>
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
    <div class="card users-card">
      <table class="users-table">
        <thead>
          <tr>
            <th>Username</th>
            <th>Email</th>
            <th>Role</th>
            <th>Tagline</th>
            <th>Last Sign In</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {#each users as user}
            <tr>
              <td class="username-cell">{user.username ?? '-'}</td>
              <td>{user.email}</td>
              <td><span class="role-badge {user.role}">{user.role}</span></td>
              <td class="tagline-cell">{user.tagline ?? '-'}</td>
              <td class="date-cell">{formatLastSignIn(user.last_sign_in_at)}</td>
              <td class="date-cell">{new Date(user.created_at).toLocaleString()}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .users-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .users-card {
    padding: 0;
    overflow-x: auto;
  }

  .users-table {
    width: 100%;
    border-collapse: collapse;
  }

  .users-table th {
    padding: 16px 12px;
    text-align: left;
    border-bottom: 2px solid rgba(255, 255, 255, 0.1);
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    background: rgba(255, 255, 255, 0.02);
  }

  .users-table td {
    padding: 14px 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .users-table tbody tr {
    transition: background-color 0.2s ease;
  }

  .users-table tbody tr:hover {
    background-color: rgba(255, 255, 255, 0.03);
  }

  .username-cell {
    font-weight: 500;
  }

  .tagline-cell {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .date-cell {
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.7);
  }

  @media only screen and (max-width: 768px) {
    .users-table th,
    .users-table td {
      padding: 12px 8px;
      font-size: 0.85rem;
    }

    .tagline-cell {
      max-width: 120px;
    }
  }
</style>