<script lang="ts">
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'
  import ConfirmModal from '$lib/components/ConfirmModal.svelte'
  import { api } from '$lib/api/client'
  import { handleApiError } from '$lib/utils/errors'
  import { toasts } from '$lib/stores/toast'
  import { getAvatarInitial } from '$lib/utils/formatters'
  import type { Season, Participant, PublicProfile } from '$lib/types'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = true
  let error = ''
  let seasonId = ''
  let season: Season | null = null

  // Participant data
  let participants: Participant[] = []
  let allUsers: PublicProfile[] = []

  // UI state
  let selectedUserIds: Set<string> = new Set()
  let addingParticipants = false
  let removingUserId: string | null = null  // track which user is being removed

  // Confirm modal state
  let showRemoveModal = false
  let userToRemove: { id: string; username: string | null } | null = null

  onMount(async () => {
    seasonId = $page.params.seasonid ?? ''
    if (!seasonId) {
      error = 'No season ID provided'
      loading = false
      return
    }

    await fetchData()
  })

  // Fetch season, participants, and all users
  async function fetchData() {
    try {
      loading = true
      error = ''

      const [seasonRes, participantsRes, usersRes] = await Promise.all([
        api.get<{ season: Season }>(`/api/seasons/${seasonId}`, session?.access_token),
        api.get<{ participants: Participant[] }>(`/api/seasons/${seasonId}/participants`, session?.access_token),
        api.get<{ users: PublicProfile[] }>('/api/users', session?.access_token)
      ])

      season = seasonRes.season
      participants = participantsRes.participants
      allUsers = usersRes.users
    } catch (err) {
      error = handleApiError(err, 'Failed to load data')
    } finally {
      loading = false
    }
  }

  // Get users who are not yet participants
  function getNonParticipantUsers(): PublicProfile[] {
    const participantIds = new Set(participants.map(p => p.user_id))
    return allUsers.filter(u => !participantIds.has(u.id))
  }

  // Toggle user selection for adding
  function toggleUserSelection(userId: string) {
    if (selectedUserIds.has(userId)) {
      selectedUserIds.delete(userId)
    } else {
      selectedUserIds.add(userId)
    }
    selectedUserIds = selectedUserIds  // trigger reactivity
  }

  // Select all available users
  function selectAll() {
    const available = getNonParticipantUsers()
    selectedUserIds = new Set(available.map(u => u.id))
  }

  // Clear selection
  function selectNone() {
    selectedUserIds = new Set()
  }

  // Add selected users as participants
  async function addSelectedParticipants() {
    if (selectedUserIds.size === 0) return

    try {
      addingParticipants = true
      const userIds = Array.from(selectedUserIds)

      const res = await api.post<{ added: number; already_existed: number }>(
        `/api/commissioner/seasons/${seasonId}/participants`,
        { user_ids: userIds },
        session?.access_token
      )

      if (res.added > 0) {
        toasts.success(`Added ${res.added} participant${res.added > 1 ? 's' : ''}`)
      }

      selectedUserIds = new Set()
      await fetchData()
    } catch (err) {
      handleApiError(err, 'Failed to add participants')
    } finally {
      addingParticipants = false
    }
  }

  // Open the remove confirmation modal
  function confirmRemoveParticipant(userId: string, username: string | null) {
    userToRemove = { id: userId, username }
    showRemoveModal = true
  }

  // Actually remove the participant (called when modal is confirmed)
  async function removeParticipant() {
    if (!userToRemove) return

    const { id: userId, username } = userToRemove

    try {
      removingUserId = userId
      showRemoveModal = false

      await api.delete(
        `/api/commissioner/seasons/${seasonId}/participants/${userId}`,
        session?.access_token
      )

      toasts.success(`${username || 'User'} removed from season`)
      await fetchData()
    } catch (err) {
      handleApiError(err, 'Failed to remove participant')
    } finally {
      removingUserId = null
      userToRemove = null
    }
  }

  // Cancel removal
  function cancelRemove() {
    showRemoveModal = false
    userToRemove = null
  }


</script>

<svelte:head>
  <title>Manage Participants{season ? ` - ${season.year} ${season.is_postseason ? 'Postseason' : 'Season'}` : ''}</title>
</svelte:head>

<div class="container">
  <div class="page-header-three-col">
    <div class="header-left">
      <a href="/commissioner/seasons/{seasonId}" class="button">← Back</a>
    </div>
    <h1 class="header-title">Manage Participants</h1>
    <div class="header-right"></div>
  </div>

  {#if loading}
    <div class="card">
      <Spinner />
    </div>
  {:else if error}
    <div class="card error-card">
      <p class="error">{error}</p>
    </div>
  {:else}
    <!-- Add Participants Section -->
    <div class="card">
      <h2>Add Participants</h2>

      {#if getNonParticipantUsers().length === 0}
        <p class="text-sm opacity-half">All users are already participants in this season.</p>
      {:else}
        <div class="selection-controls">
          <button class="button small" on:click={selectAll}>Select All</button>
          <button class="button small" on:click={selectNone}>Select None</button>
          <span class="selection-count">{selectedUserIds.size} selected</span>
        </div>

        <div class="user-picker">
          {#each getNonParticipantUsers() as user}
            <label class="user-option" class:selected={selectedUserIds.has(user.id)}>
              <input
                type="checkbox"
                checked={selectedUserIds.has(user.id)}
                on:change={() => toggleUserSelection(user.id)}
              />
              <div class="user-avatar">
                {#if user.avatar_url}
                  <img src={user.avatar_url} alt={user.username || 'User'} />
                {:else}
                  <div class="avatar-placeholder">{getAvatarInitial(user.username)}</div>
                {/if}
              </div>
              <span class="user-name">{user.username || 'Anonymous'}</span>
            </label>
          {/each}
        </div>

        <div class="add-actions">
          <button
            class="button primary"
            on:click={addSelectedParticipants}
            disabled={selectedUserIds.size === 0 || addingParticipants}
          >
            {addingParticipants ? 'Adding...' : `Add ${selectedUserIds.size} Participant${selectedUserIds.size !== 1 ? 's' : ''}`}
          </button>
        </div>
      {/if}
    </div>

    <!-- Current Participants Section -->
    <div class="card">
      <h2>Current Participants ({participants.length})</h2>

      {#if participants.length === 0}
        <p class="text-sm opacity-half">No participants yet.</p>
      {:else}
        <div class="participants-list">
          {#each participants as participant}
            <div class="participant-row">
              <div class="participant-info">
                <div class="participant-avatar">
                  <img src={participant.avatar_url} alt={participant.username || 'User'} />
                </div>
                <span class="participant-name">{participant.username || 'Anonymous'}</span>
              </div>
              <button
                class="button small danger"
                on:click={() => confirmRemoveParticipant(participant.user_id, participant.username)}
                disabled={removingUserId === participant.user_id}
              >
                {removingUserId === participant.user_id ? 'Removing...' : 'Remove'}
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Remove participant confirmation modal -->
<ConfirmModal
  bind:open={showRemoveModal}
  title="Remove Participant"
  message="Remove {userToRemove?.username || 'this user'} from the season? They will no longer appear in standings."
  confirmText="Remove"
  cancelText="Cancel"
  confirmVariant="danger"
  on:confirm={removeParticipant}
  on:cancel={cancelRemove}
/>

<style>
  .container {
    max-width: 800px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  .card h2 {
    font-size: 1.1rem;
    margin: 0 0 15px 0;
    padding-bottom: 10px;
    border-bottom: var(--custom-border);
  }

  .card + .card {
    margin-top: 20px;
  }

  /* Selection controls */
  .selection-controls {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 15px;
  }

  .selection-count {
    margin-left: auto;
    font-size: 0.9rem;
    color: var(--custom-color-secondary, #666);
  }

  /* User picker grid */
  .user-picker {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 10px;
    margin-bottom: 15px;
    max-height: 400px;
    overflow-y: auto;
  }

  .user-option {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px;
    border: 1px solid var(--custom-border-color, #ddd);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .user-option:hover {
    border-color: var(--custom-color-brand);
  }

  .user-option.selected {
    border-color: var(--custom-color-brand);
    background: var(--custom-color-brand-light, rgba(0, 112, 243, 0.1));
  }

  .user-option input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  .user-avatar img {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
  }

  .user-avatar .avatar-placeholder {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: var(--custom-color-brand);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.9rem;
  }

  .user-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .add-actions {
    display: flex;
    justify-content: flex-end;
  }

  /* Participants list */
  .participants-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .participant-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px;
    border: 1px solid var(--custom-border-color, #ddd);
    border-radius: 8px;
  }

  .participant-info {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .participant-avatar img {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
  }

  .participant-name {
    font-weight: 500;
  }

  .button.danger {
    background: var(--custom-color-error, #dc3545);
    color: white;
  }

  .button.danger:hover:not(:disabled) {
    background: #c82333;
  }

  .button.danger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
