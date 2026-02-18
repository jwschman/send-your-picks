<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import { toasts } from '$lib/stores/toast'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { DEFAULT_AVATAR } from '$lib/utils/formatters'
  import type { User } from '$lib/types'

  export let data: PageData

  let { session, supabase } = data
  $: ({ session, supabase } = data)

  let loading = true
  let saving = false
  let uploadingAvatar = false
  let profile: User | null = null
  let error = ''

  let username = ''
  let tagline = ''
  let avatarUrl: string | null = null
  let avatarPreviewUrl: string | null = null

  onMount(async () => {
    await fetchProfile()
  })

  async function fetchProfile() {
    try {
      loading = true
      error = ''

      profile = await api.get<User>('/api/account', session?.access_token)

      username = profile.username || ''
      tagline = profile.tagline || ''
      avatarUrl = profile.avatar_url || null

      if (avatarUrl) {
        avatarPreviewUrl = supabase.storage
          .from('avatars')
          .getPublicUrl(avatarUrl).data.publicUrl
      }

    } catch (err) {
      error = handleApiError(err, 'Failed to fetch profile')
    } finally {
      loading = false
    }
  }

  async function uploadAvatar(file: File) {
    if (!session?.user?.id) return

    try {
      uploadingAvatar = true

      const ext = file.name.split('.').pop()
      const path = `${session.user.id}/avatar.${ext}`

      const { error: uploadError } = await supabase.storage
        .from('avatars')
        .upload(path, file, { upsert: true })

      if (uploadError) {
        throw uploadError
      }

      avatarUrl = path
      avatarPreviewUrl = supabase.storage
        .from('avatars')
        .getPublicUrl(path).data.publicUrl

    } catch (err) {
      // Just show toast for action errors
      handleApiError(err, 'Failed to upload avatar')
    } finally {
      uploadingAvatar = false
    }
  }

  async function updateProfile() {
    try {
      saving = true

      profile = await api.put<User>('/api/account', {
        username: username || null,
        tagline: tagline || null,
        avatar_url: avatarUrl
      }, session?.access_token)

      toasts.success('Profile updated successfully!')

    } catch (err) {
      // Just show toast for action errors
      handleApiError(err, 'Failed to update profile')
    } finally {
      saving = false
    }
  }

  async function handleSignOut() {
    await supabase.auth.signOut()
    goto('/login')
  }
</script>

<svelte:head>
  <title>Account</title>
</svelte:head>

<div class="account-container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/dashboard" class="button">← Back</a>
    </div>
    <h1 class="header-title">My Account</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>

  {:else if error}
    <div class="card error-card">
      <p class="error">{error}</p>
      <button class="button" on:click={fetchProfile}>Retry</button>
    </div>

  {:else if profile}
    <form on:submit|preventDefault={updateProfile}>
      <!-- Avatar Section -->
      <div class="card avatar-section">
        <h2>Profile Picture</h2>
        <div class="avatar-content">
          <div class="avatar-preview">
            <img src={avatarPreviewUrl || DEFAULT_AVATAR} alt="Your avatar" class="avatar-image" />
          </div>
          <label class="button upload-btn">
            {uploadingAvatar ? 'Uploading...' : 'Upload New Profile Picture'}
            <input
              type="file"
              accept="image/*"
              disabled={uploadingAvatar}
              on:change={(e) => {
                const target = e.target as HTMLInputElement
                const file = target.files?.[0]
                if (file) uploadAvatar(file)
              }}
            />
          </label>
        </div>
      </div>

      <!-- Account Info (Read-only) -->
      <div class="card info-card">
        <h2>Account Information</h2>
        <div class="info-row">
          <span class="info-label">Email</span>
          <span class="info-value">{profile.email}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Role</span>
          <span class="info-value">
            <span class="role-badge {profile.role}">{profile.role}</span>
          </span>
        </div>
      </div>

      <!-- Editable Profile -->
      <div class="card">
        <h2>Edit Profile</h2>
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            type="text"
            bind:value={username}
            placeholder="Choose a display name"
          />
        </div>
        <div class="form-group">
          <label for="tagline">Tagline</label>
          <input
            id="tagline"
            type="text"
            bind:value={tagline}
            placeholder="A short bio or motto"
          />
        </div>
        <button type="submit" class="button primary save-btn" disabled={saving}>
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </form>
  {/if}
</div>

<style>
  .account-container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .card {
    margin-bottom: 20px;
  }

  .card h2 {
    font-size: 1.1rem;
    margin: 0 0 20px 0;
    padding-bottom: 10px;
    border-bottom: var(--custom-border);
  }

  /* Avatar Section */
  .avatar-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 20px;
  }

  .avatar-preview {
    flex-shrink: 0;
  }

  .avatar-image {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid var(--custom-color-brand);
  }

  .upload-btn {
    position: relative;
    cursor: pointer;
  }

  .upload-btn input[type="file"] {
    position: absolute;
    opacity: 0;
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;
    cursor: pointer;
  }

  /* Form */
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

  .save-btn {
    width: 100%;
    padding: 14px;
    font-size: 1rem;
  }

</style>
