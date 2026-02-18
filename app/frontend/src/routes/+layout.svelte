<script lang="ts">
  import { page } from '$app/stores'
  import { invalidate, goto } from '$app/navigation'
  import { onMount } from 'svelte'
  import { env } from '$env/dynamic/public'
  import ToastContainer from '$lib/components/ToastContainer.svelte'
  import { logAuth } from '$lib/utils/authLogger'
  import { logUser } from '$lib/utils/logger'
  import { DEFAULT_AVATAR } from '$lib/utils/formatters'
  import '../styles.css'

  export let data
  let { session, supabase } = data
  $: ({ session, supabase } = data)

  let userRole = ''
  let username = ''
  let email = ''
  let avatarUrl: string | null = null
  let userMenuOpen = false
  let userInfoFetched = false
  let mounted = false

  // Reactively fetch user info when session becomes available
  $: if (mounted && session && !userInfoFetched) {
    fetchUserInfo()
  }

  // Reset fetch flag when session changes (e.g., different user)
  $: if (!session) {
    userInfoFetched = false
    userRole = ''
    username = ''
    email = ''
    avatarUrl = null
  }

  onMount(() => {
    mounted = true

    const { data: authData } = supabase.auth.onAuthStateChange((event, _session) => {
      // Log auth state changes
      if (event === 'SIGNED_IN') {
        logAuth('LOGIN_SUCCESS', { email: _session?.user?.email })
        // Reset so we fetch the new user's info
        userInfoFetched = false
      } else if (event === 'SIGNED_OUT') {
        logAuth('LOGOUT', { email: session?.user?.email })
      } else if (event === 'TOKEN_REFRESHED') {
        logAuth('TOKEN_REFRESH', {
          email: _session?.user?.email,
          details: {
            expires_at: _session?.expires_at ? new Date(_session.expires_at * 1000).toISOString() : null
          }
        })
      }

      if (_session?.expires_at !== session?.expires_at) {
        invalidate('supabase:auth')
      }
    })

    return () => authData.subscription.unsubscribe()
  })

  async function fetchUserInfo() {
    if (typeof window === 'undefined') return

    // Mark as fetched immediately to prevent duplicate calls
    userInfoFetched = true

    try {
      const token = session?.access_token
      const response = await fetch(`${env.PUBLIC_API_BASE_URL}/api/whoami`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      })

      if (response.ok) {
        const profile = await response.json()
        userRole = profile.role || ''
        username = profile.username || profile.email || 'User'
        email = profile.email || ''
        avatarUrl = profile.avatar_url || null
        logUser('INFO_FETCH_SUCCESS', { hasSession: true })
      } else {
        logUser('INFO_FETCH_FAILURE', {
          error: `Status ${response.status}`,
          hasSession: !!session
        })
      }
    } catch (err) {
      logUser('INFO_FETCH_FAILURE', {
        error: err instanceof Error ? err.message : 'Unknown error',
        hasSession: !!session
      })
    }
  }

  function toggleUserMenu() {
    userMenuOpen = !userMenuOpen
  }

  function closeUserMenu() {
    userMenuOpen = false
  }

  async function handleSignOut() {
    await supabase.auth.signOut()
    goto('/login')
  }

  $: showNav = session && $page.url.pathname !== '/login'

  // Close dropdown when clicking outside
  function handleClickOutside(event: MouseEvent) {
    if (userMenuOpen) {
      const target = event.target as HTMLElement
      if (!target.closest('.user-menu-container')) {
        closeUserMenu()
      }
    }
  }
</script>

<svelte:window on:click={handleClickOutside} />

<svelte:head>
  <title>{env.PUBLIC_APP_NAME || 'Send Your Picks'}</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="">
  <link
    href="https://fonts.googleapis.com/css2?family=Barrio&display=swap"
    rel="stylesheet"
  >
</svelte:head>

{#if showNav}
  <nav class="top-nav">
    <div class="nav-container">
      <div class="nav-left">
        <a href="/dashboard" class="app-title">{env.PUBLIC_APP_NAME || 'Send Your Picks'}</a>
      </div>

      <div class="nav-right">
        <a href="/dashboard" class="nav-link">Dashboard</a>

        <div class="user-menu-container">
          <button class="avatar-button" on:click={toggleUserMenu}>
            <img
              src={avatarUrl || DEFAULT_AVATAR}
              alt="User avatar"
              class="avatar"
            />
          </button>

          {#if userMenuOpen}
            <div class="user-dropdown">
              <div class="user-info">
                <div class="dropdown-username">{username}</div>
                <div class="dropdown-email">{email}</div>
              </div>
              <div class="dropdown-divider"></div>
              <a href="/account" class="dropdown-item" on:click={closeUserMenu}>
                My Profile
              </a>
              {#if userRole === 'commissioner' || userRole === 'admin'}
                <a href="/commissioner" class="dropdown-item" on:click={closeUserMenu}>
                  Commissioner Dashboard
                </a>
              {/if}
              {#if userRole === 'admin'}
                <a href="/admin" class="dropdown-item" on:click={closeUserMenu}>
                  Admin Dashboard
                </a>
              {/if}
              <div class="dropdown-divider"></div>
              <button class="dropdown-item sign-out" on:click={handleSignOut}>
                Sign Out
              </button>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </nav>
{/if}

<ToastContainer />

<main class:with-nav={showNav}>
  <slot />
</main>

<style>
  .top-nav {
    background-color: var(--custom-color-brand);
    border-bottom: 2px solid rgba(255, 255, 255, 0.1);
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .nav-container {
    padding: 0 15px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 60px;
  }

  .nav-left {
    display: flex;
    align-items: center;
  }

  .app-title {
    font-size: 1.5rem;
    font-weight: bold;
    color: white;
    text-decoration: none;
    font-family: 'Barrio', cursive;
    letter-spacing: 1px;
    transition: transform 0.2s ease;
  }

  .app-title:hover {
    transform: scale(1.05);
  }

  .nav-right {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .nav-link {
    color: white;
    text-decoration: none;
    font-weight: 500;
    padding: 8px 12px;
    border-radius: var(--custom-border-radius);
    transition: background-color 0.2s ease;
  }

  .nav-link:hover {
    background-color: rgba(255, 255, 255, 0.15);
  }

  .user-menu-container {
    position: relative;
    margin-left: 20px;
    padding-left: 20px;
    border-left: 1px solid rgba(255, 255, 255, 0.3);
  }

  .avatar-button {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    display: flex;
    align-items: center;
    transition: transform 0.2s ease;
  }

  .avatar-button:hover {
    transform: scale(1.05);
  }

  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid rgba(255, 255, 255, 0.3);
    transition: border-color 0.2s ease;
  }

  .avatar-button:hover .avatar {
    border-color: rgba(255, 255, 255, 0.6);
  }

  .user-dropdown {
    position: absolute;
    top: calc(100% + 10px);
    right: 0;
    background: var(--custom-panel-color);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: var(--custom-border-radius);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    min-width: 220px;
    z-index: 1001;
    overflow: hidden;
  }

  .user-info {
    padding: 12px 16px;
  }

  .dropdown-username {
    font-weight: 600;
    font-size: 0.95rem;
    color: var(--custom-color);
    margin-bottom: 4px;
  }

  .dropdown-email {
    font-size: 0.8rem;
    color: rgba(255, 255, 255, 0.6);
  }

  .dropdown-divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
  }

  .dropdown-item {
    display: block;
    width: 100%;
    padding: 10px 16px;
    text-align: left;
    background: none;
    border: none;
    color: var(--custom-color);
    font-size: 0.9rem;
    cursor: pointer;
    transition: background-color 0.2s ease;
    text-decoration: none;
  }

  .dropdown-item:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .dropdown-item.sign-out {
    color: #ff6b6b;
  }

  .dropdown-item.sign-out:hover {
    background-color: rgba(255, 107, 107, 0.1);
  }

  main {
    min-height: 100vh;
    padding: 20px 0;
  }

  main.with-nav {
    padding-top: 10px;
  }

  @media only screen and (max-width: 768px) {
    .nav-container {
      padding: 0 10px;
    }

    .nav-right {
      gap: 10px;
    }

    .nav-link {
      padding: 6px 8px;
      font-size: 0.9rem;
    }

    .user-menu-container {
      margin-left: 10px;
      padding-left: 10px;
    }

    .avatar {
      width: 36px;
      height: 36px;
    }

    .app-title {
      font-size: 1.2rem;
    }
  }

  @media only screen and (max-width: 400px) {
    .app-title {
      font-size: 1rem;
    }

    .nav-link {
      padding: 6px;
      font-size: 0.85rem;
    }

    .user-menu-container {
      margin-left: 8px;
      padding-left: 8px;
    }

    .avatar {
      width: 32px;
      height: 32px;
    }
  }
</style>
