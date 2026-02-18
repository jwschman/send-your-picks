<script lang="ts">
  import { onMount } from 'svelte'
  import { goto } from '$app/navigation'
  import type { PageData } from './$types'

  export let data: PageData
  let { session } = data

  onMount(() => {
    if (session) {
      goto('/dashboard')
    } else {
      goto('/login')
    }
  })
</script>

<div class="loading">
  <p>Loading...</p>
</div>