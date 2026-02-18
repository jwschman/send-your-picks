<script lang="ts">
  import { goto } from '$app/navigation'
  import { env } from '$env/dynamic/public'
  import { logAuth, getJwtDebugInfo } from '$lib/utils/authLogger'

  export let data

  let { supabase, session } = data
  $: ({ supabase, session } = data)

  let loading = false
  let email = ''
  let password = ''
  let confirmPassword = ''
  let mode: 'login' | 'signup' = 'login'
  let errorMsg = ''
  let signupSuccess = false

  async function handleAuth() {
    try {
      loading = true
      errorMsg = ''

      if (mode === 'signup') {
        if (password !== confirmPassword) {
          errorMsg = 'Passwords do not match'
          logAuth('SIGNUP_FAILURE', { email, error: 'Passwords do not match' })
          return
        }

        logAuth('SIGNUP_ATTEMPT', { email })

        const { error } = await supabase.auth.signUp({
          email,
          password,
          options: {
            emailRedirectTo: `${window.location.origin}/auth/callback`
          }
        })

        if (error) throw error

        logAuth('SIGNUP_SUCCESS', { email })
        signupSuccess = true
        password = ''
        confirmPassword = ''
      } else {
        logAuth('LOGIN_ATTEMPT', { email })

        const { data: authData, error } = await supabase.auth.signInWithPassword({
          email,
          password,
        })

        if (error) throw error

        // Log successful login with JWT debug info
        const jwtInfo = getJwtDebugInfo(authData.session?.access_token)
        logAuth('LOGIN_SUCCESS', {
          email,
          details: {
            jwt: jwtInfo,
            clientTime: new Date().toISOString()
          }
        })

        goto('/dashboard')
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error'
      errorMsg = errorMessage

      logAuth(mode === 'signup' ? 'SIGNUP_FAILURE' : 'LOGIN_FAILURE', {
        email,
        error: errorMessage,
        details: {
          clientTime: new Date().toISOString()
        }
      })
    } finally {
      loading = false
    }
  }

  function toggleMode() {
    mode = mode === 'login' ? 'signup' : 'login'
    errorMsg = ''
    signupSuccess = false
    password = ''
    confirmPassword = ''
  }
</script>

<svelte:head>
  <title>{mode === 'login' ? 'Sign In' : 'Sign Up'} - {env.PUBLIC_APP_NAME || 'Send Your Picks'}</title>
</svelte:head>

<div class="login-container">
  <div class="login-card card">
    <div class="login-header">
      <h1 class="app-title">{env.PUBLIC_APP_NAME || 'Send Your Picks'}</h1>
      <p class="login-subtitle">
        {#if session}
          Welcome back!
        {:else}
          {mode === 'login' ? 'Sign in to your account' : 'Create your account'}
        {/if}
      </p>
    </div>

    {#if session}
      <div class="already-logged-in">
        <p class="logged-in-email">{session.user.email}</p>
        <button class="button primary block" on:click={() => goto('/dashboard')}>
          Go to Dashboard
        </button>
      </div>
    {:else if signupSuccess}
      <div class="signup-success">
        <div class="success-icon">&#10003;</div>
        <h2>Check Your Email</h2>
        <p>We've sent a confirmation link to <strong>{email}</strong>.</p>
        <p>Please click the link in your email to verify your account before signing in.</p>
        <button class="button primary block" on:click={() => { signupSuccess = false; mode = 'login' }}>
          Back to Sign In
        </button>
      </div>
    {:else}
      <div class="mode-tabs">
        <button
          class="mode-tab"
          class:active={mode === 'login'}
          on:click={() => { if (mode !== 'login') toggleMode() }}
          disabled={loading}
        >
          Sign In
        </button>
        <button
          class="mode-tab"
          class:active={mode === 'signup'}
          on:click={() => { if (mode !== 'signup') toggleMode() }}
          disabled={loading}
        >
          Create Account
        </button>
      </div>

      <form class="login-form" on:submit|preventDefault={handleAuth}>
        <div class="form-group">
          <label for="email">Email</label>
          <input
            id="email"
            type="email"
            placeholder="you@example.com"
            bind:value={email}
            required
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            type="password"
            placeholder="Your password"
            bind:value={password}
            required
            minlength="6"
          />
        </div>

        {#if mode === 'signup'}
          <div class="form-group">
            <label for="confirmPassword">Confirm Password</label>
            <input
              id="confirmPassword"
              type="password"
              placeholder="Re-enter your password"
              bind:value={confirmPassword}
              required
              minlength="6"
            />
          </div>
        {/if}

        {#if errorMsg}
          <div class="error-message">
            {errorMsg}
          </div>
        {/if}

        <button type="submit" class="button primary block" disabled={loading}>
          {loading ? 'Loading...' : mode === 'login' ? 'Sign In' : 'Create Account'}
        </button>
      </form>

      <div class="login-footer">
        <button class="toggle-mode-btn" on:click={toggleMode} disabled={loading}>
          {mode === 'login' ? "Don't have an account? Sign up" : 'Already have an account? Sign in'}
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .login-container {
    min-height: 100vh;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 80px 20px 20px;
  }

  .login-card {
    width: 100%;
    max-width: 400px;
    padding: 40px;
  }

  .login-header {
    text-align: center;
    margin-bottom: 30px;
  }

  .app-title {
    font-family: 'Barrio', cursive;
    font-size: 2.5rem;
    color: var(--custom-color-brand);
    margin: 0 0 10px 0;
    letter-spacing: 1px;
  }

  .login-subtitle {
    color: rgba(255, 255, 255, 0.6);
    font-size: 0.95rem;
    margin: 0;
  }

  .mode-tabs {
    display: flex;
    margin-bottom: 24px;
    border-radius: var(--custom-border-radius);
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  .mode-tab {
    flex: 1;
    padding: 12px 16px;
    border: none;
    background: transparent;
    color: rgba(255, 255, 255, 0.6);
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .mode-tab:first-child {
    border-right: 1px solid rgba(255, 255, 255, 0.2);
  }

  .mode-tab:hover:not(.active):not(:disabled) {
    background: rgba(255, 255, 255, 0.05);
  }

  .mode-tab.active {
    background: var(--custom-color-brand);
    color: white;
  }

  .mode-tab:disabled {
    cursor: not-allowed;
  }

  .signup-success {
    text-align: center;
    padding: 20px 0;
  }

  .success-icon {
    width: 60px;
    height: 60px;
    margin: 0 auto 20px;
    background: rgba(34, 197, 94, 0.2);
    border: 2px solid #22c55e;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    color: #22c55e;
  }

  .signup-success h2 {
    margin: 0 0 16px 0;
    font-size: 1.4rem;
  }

  .signup-success p {
    color: rgba(255, 255, 255, 0.7);
    margin: 0 0 12px 0;
    font-size: 0.95rem;
    line-height: 1.5;
  }

  .signup-success p strong {
    color: rgba(255, 255, 255, 0.9);
  }

  .signup-success .button {
    margin-top: 24px;
  }

  .login-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .form-group label {
    font-weight: 500;
    font-size: 0.9rem;
  }

  .form-group input {
    padding: 12px;
    font-size: 1rem;
    border-radius: var(--custom-border-radius);
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.05);
    color: var(--custom-color);
    transition: border-color 0.2s ease, background-color 0.2s ease;
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--custom-color-brand);
    background: rgba(255, 255, 255, 0.08);
  }

  .form-group input::placeholder {
    color: rgba(255, 255, 255, 0.4);
  }

  .error-message {
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #ef4444;
    padding: 12px;
    border-radius: var(--custom-border-radius);
    font-size: 0.9rem;
  }

  .button.block {
    width: 100%;
    padding: 14px;
    font-size: 1rem;
  }

  .login-footer {
    margin-top: 24px;
    text-align: center;
  }

  .toggle-mode-btn {
    background: none;
    border: none;
    color: var(--custom-color-brand);
    cursor: pointer;
    font-size: 0.9rem;
    padding: 8px;
    transition: opacity 0.2s ease;
  }

  .toggle-mode-btn:hover {
    opacity: 0.8;
  }

  .toggle-mode-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .already-logged-in {
    text-align: center;
  }

  .logged-in-email {
    color: rgba(255, 255, 255, 0.8);
    margin-bottom: 20px;
    font-size: 0.95rem;
  }

  @media only screen and (max-width: 480px) {
    .login-container {
      padding-top: 40px;
    }

    .login-card {
      padding: 30px 20px;
    }

    .app-title {
      font-size: 2rem;
    }
  }
</style>
