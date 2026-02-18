<script lang="ts">
  import { env } from '$env/dynamic/public'
  import { onMount } from 'svelte'
  import type { PageData } from './$types'
  import Spinner from '$lib/components/Spinner.svelte'

  export let data: PageData
  let { session } = data
  $: ({ session } = data)

  let loading = true
  let error = ''
  let payload: unknown = null

  onMount(async () => {
    await fetchSettings()
  })

  async function fetchSettings() {
    try {
      const token = session?.access_token

      const response = await fetch(
        `${env.PUBLIC_API_BASE_URL}/api/settings`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      )

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`)
      }

      payload = await response.json()
    } catch (err: unknown) {
      error = err instanceof Error ? err.message : 'An unknown error occurred'
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>JSON Payload</title>
</svelte:head>

<div class="container">
  <h1>JSON PAYLOAD</h1>

  {#if loading}
    <Spinner />
  {:else if error}
    <p class="error">{error}</p>
  {:else}
    <pre>{JSON.stringify(payload, null, 2)}</pre>
  {/if}
</div>

<style>
  .container {
    max-width: 1000px;
    margin: 0 auto;
    padding: 0 20px 20px;
  }

  pre {
    background: #111;
    color: #eee;
    padding: 16px;
    overflow-x: auto;
    font-size: 14px;
    line-height: 1.4;
  }

  .error {
    color: red;
  }
</style>
