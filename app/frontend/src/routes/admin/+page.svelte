<script lang="ts">
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'

  export let data: PageData
  let { session, supabase } = data
  $: ({ session, supabase } = data)

  async function handleSignOut() {
    await supabase.auth.signOut()
    goto('/login')
  }
</script>

<svelte:head>
  <title>Admin Dashboard</title>
</svelte:head>

<div class="dashboard">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/dashboard" class="button">← Back</a>
    </div>
    <h1 class="header-title">Admin Dashboard</h1>
    <div class="header-right"></div>
  </div>

  <div class="nav-links">
    <h2>Links</h2>

    <a href="/admin/settings" class="button block">Global Settings</a>
    <a href="/admin/users" class="button block">User Management</a>

    <button class="button block" on:click={handleSignOut}>
      Sign Out
    </button>
  </div>
</div>

<style>
  .dashboard {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }
</style>
