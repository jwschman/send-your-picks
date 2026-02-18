# Send Your Picks Frontend

A modern SvelteKit 2 application for managing NFL pool picks and standings. Built with Svelte 5, TypeScript, and Supabase authentication.

## Table of Contents

- [Quick Start](#quick-start)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Routes](#routes)
- [Components](#components)
- [State Management](#state-management)
- [API Client](#api-client)
- [Authentication](#authentication)
- [Styling](#styling)
- [Logging](#logging)
- [Types](#types)
- [Configuration](#configuration)
- [Development](#development)
- [Deployment](#deployment)

---

## Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Type check
npm run check

# Production build
npm run build
```

The dev server runs at `http://localhost:5173` with `--host` flag for network access.

---

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| SvelteKit | 2.49+ | Full-stack framework |
| Svelte | 5.45+ | UI framework |
| TypeScript | 5.9+ | Type safety |
| Supabase | 2.89+ | Authentication |
| Chart.js | 4.5+ | Data visualization |
| Vite | 7.2+ | Build tool |

---

## Project Structure

```
src/
├── lib/
│   ├── api/
│   │   └── client.ts              # HTTP client with auth & logging
│   ├── components/
│   │   ├── BadgeList.svelte       # User achievement badges
│   │   ├── ConfirmModal.svelte    # Confirmation dialogs
│   │   ├── EmptyState.svelte      # Empty state placeholders
│   │   ├── LoadingCard.svelte     # Skeleton loading cards
│   │   ├── Skeleton.svelte        # Shimmer loading effect
│   │   ├── Spinner.svelte         # Loading spinner
│   │   └── ToastContainer.svelte  # Toast notifications
│   ├── stores/
│   │   └── toast.ts               # Toast notification store
│   ├── types/
│   │   └── index.ts               # TypeScript type definitions
│   ├── utils/
│   │   ├── auth.ts                # JWT parsing & role checks
│   │   ├── authLogger.ts          # Auth event logging
│   │   ├── errors.ts              # Error handling utilities
│   │   ├── formatters.ts          # Shared date/time, team & avatar utilities
│   │   └── logger.ts              # Structured logging system
│   └── index.ts                   # $lib exports
├── routes/                        # SvelteKit pages (see Routes section)
├── app.d.ts                       # Global TypeScript declarations
├── app.html                       # HTML template
├── hooks.server.ts                # Server hooks (auth setup)
└── styles.css                     # Global styles & CSS variables
```

---

## Routes

### Public Routes

| Route | Purpose |
|-------|---------|
| `/` | Redirects to dashboard (logged in) or login |
| `/login` | Login and signup with email verification |
| `/auth/callback` | Email verification callback (PKCE code exchange) |
| `/auth/error` | Authentication error display |

### User Routes (requires authentication)

| Route | Purpose |
|-------|---------|
| `/dashboard` | Main dashboard with active week & standings |
| `/account` | Profile management (username, tagline, avatar) |
| `/seasons` | List all seasons |
| `/seasons/[seasonid]` | Season detail with weeks list |
| `/seasons/[seasonid]/standings` | Season leaderboard with chart |
| `/seasons/[seasonid]/points` | Points breakdown visualization |
| `/seasons/[seasonid]/winners` | Weekly winners history |
| `/seasons/[seasonid]/weeks/[weekid]` | Week detail with games |
| `/seasons/[seasonid]/weeks/[weekid]/picks` | Make/view picks |
| `/seasons/[seasonid]/weeks/[weekid]/allpicks` | View all users' picks |
| `/seasons/[seasonid]/weeks/[weekid]/results` | Game results |
| `/seasons/[seasonid]/weeks/[weekid]/standings` | Weekly leaderboard |
| `/users` | List all pool users |
| `/users/[userid]` | User public profile |
| `/settings` | User settings |
| `/whoami` | Current user info (debug) |

### Commissioner Routes (requires commissioner or admin role)

| Route | Purpose |
|-------|---------|
| `/commissioner` | Commissioner dashboard |
| `/commissioner/seasons` | Manage seasons |
| `/commissioner/seasons/new` | Create new season |
| `/commissioner/seasons/[seasonid]` | Season detail (view-only info, participants, weeks) |
| `/commissioner/seasons/[seasonid]/manage` | Season management (activate/deactivate, settings) |
| `/commissioner/seasons/[seasonid]/participants` | Add/remove season participants |
| `/commissioner/seasons/[seasonid]/weeks/[weekid]` | Week management & activation |
| `/commissioner/seasons/[seasonid]/weeks/[weekid]/spreads` | Set game spreads |
| `/commissioner/seasons/[seasonid]/weeks/[weekid]/picks` | View all picks |

### Admin Routes (requires admin role)

| Route | Purpose |
|-------|---------|
| `/admin` | Admin dashboard |
| `/admin/users` | User management & role assignment |
| `/admin/settings` | Global pool settings |

### Route-Specific Notes

#### Login Page (`/login`)

The login page supports both sign-in and sign-up modes:
- **Mode Tabs**: Toggle between "Sign In" and "Create Account" with clear visual indicator
- **Confirm Password**: Sign-up mode requires password confirmation with validation
- **Email Verification**: After signup, displays instructions to check email for verification link
- **Layout**: Form is positioned from top (not centered) so mode switching doesn't shift content

#### Users Page (`/users`)

Displays all pool users in a responsive grid:
- **Text Overflow**: Long usernames and taglines truncate with ellipsis to prevent overflow into adjacent cards
- **Avatar Fallback**: Shows first letter of username when no avatar is set
- **Badge Display**: Shows user achievement badges (week winner, season winner/loser)

#### Week Page (`/seasons/[seasonid]/weeks/[weekid]`)

Displays week details and actions:
- **Make or Edit Picks**: Primary button when week is active and picks aren't locked
- **View My Picks**: Regular button when picks are locked or week isn't active (differentiates from "View All Picks")
- **Lock Picks**: Validates user has made at least one pick before allowing lock, shows toast error if no picks

#### Picks Page (`/seasons/[seasonid]/weeks/[weekid]/picks`)

Make or view picks for a week:
- **Locked State Detection**: Checks if user has locked their picks via `user_locked_at` field
- **View Mode**: When picks are locked, page shows "View Your Picks" header, hides submit button, and displays lock indicator
- **Team Selection**: Locked picks display at full opacity with 🔒 emoji indicator (not grayed out)
- **Lock Feedback**: Clicking a locked pick shows toast "Your picks are locked and cannot be changed"

#### Commissioner Season Detail (`/commissioner/seasons/[seasonid]`)

View-only season information page:
- **Season Info**: Year, status, total weeks
- **Participants List**: Shows avatars and usernames of all season participants (view-only)
- **Weeks Grid**: Lists all weeks with status badges and links to week detail/spread editing
- **Manage Season**: Button links to the management page for edits
- **Advance Season State**: Temporary button for advancing the season state machine

#### Season Management Page (`/commissioner/seasons/[seasonid]/manage`)

Intentionally separate from the detail page to prevent accidental changes:
- **Season Status**: Shows active/inactive badge with toggle button (Activate/Deactivate)
  - Activate is primary (blue), Deactivate is danger (red)
  - Activating when another season is active shows backend error via toast
  - Deactivating an already inactive season shows error via toast
- **Season Settings**: Inline-edit pattern for number of weeks
  - Displays current value with "Edit" button
  - Clicking Edit swaps to number input with Save/Cancel
  - Backend prevents setting below the number of weeks already created
- **Management Links**: Links to participant management

#### Participant Management (`/commissioner/seasons/[seasonid]/participants`)

Add or remove users from a season:
- **Current Participants**: List with avatars, usernames, and remove button per user
- **Add Participants**: Picker showing non-participant users with Select All/None
- **Remove Confirmation**: Uses ConfirmModal component (not browser `confirm()`)
- **Idempotent Adds**: Adding existing participants is a no-op (ON CONFLICT DO NOTHING)

#### Commissioner Week Page (`/commissioner/seasons/[seasonid]/weeks/[weekid]`)

Week management for commissioners:
- **Activate Week**: Button is always clickable; shows toast error "You must set spreads for all games before activating the week" if spreads aren't set (rather than being silently disabled)

---

## Components

### Spinner

Loading indicator with customizable size and color.

```svelte
<Spinner size="md" color="var(--custom-color-brand)" />
```

**Props:**
- `size`: `'sm'` | `'md'` | `'lg'` (default: `'md'`)
- `color`: CSS color string (default: brand color)

### Skeleton

Shimmer loading placeholder.

```svelte
<Skeleton width="100%" height="20px" count={3} gap="8px" />
```

**Props:**
- `width`: CSS width (default: `'100%'`)
- `height`: CSS height (default: `'1em'`)
- `borderRadius`: CSS border radius (default: `'4px'`)
- `count`: Number of skeleton lines (default: `1`)
- `gap`: Gap between lines (default: `'0.5em'`)

### LoadingCard

Card-formatted loading state with skeleton lines.

```svelte
<LoadingCard lines={4} />
```

**Props:**
- `lines`: Number of skeleton lines to show

### EmptyState

Placeholder for empty data states.

```svelte
<EmptyState
  title="No Seasons"
  message="There are no seasons yet."
  icon="📅"
  actionText="Create Season"
  actionHref="/commissioner/seasons/new"
/>
```

**Props:**
- `title`: Heading text
- `message`: Description text
- `icon`: Emoji icon (optional)
- `actionText`: Button text (optional)
- `actionHref`: Button link (optional)
- Also accepts a slot for custom content

### BadgeList

Displays user achievement badges.

```svelte
<BadgeList badges={userBadges} compact={true} />
```

**Props:**
- `badges`: Array of `Badge` objects
- `compact`: Boolean for compact display mode

**Badge Types:**
- `previous_week_winner` - Won the previous week
- `previous_season_winner` - Won the previous season
- `previous_season_loser` - Last place previous season

### ToastContainer

Displays toast notifications. Include once in root layout.

```svelte
<ToastContainer />
```

Uses the `toasts` store for notifications. Supports auto-dismiss and manual close.

### ConfirmModal

Confirmation dialog for destructive actions.

```svelte
<ConfirmModal
  open={showModal}
  title="Delete Season"
  message="Are you sure? This cannot be undone."
  confirmText="Delete"
  confirmVariant="danger"
  on:confirm={handleDelete}
  on:cancel={() => showModal = false}
/>
```

**Props:**
- `open`: Boolean to show/hide modal
- `title`: Modal title
- `message`: Confirmation message
- `confirmText`: Confirm button text (default: `'Confirm'`)
- `cancelText`: Cancel button text (default: `'Cancel'`)
- `confirmVariant`: `'primary'` | `'danger'` (default: `'primary'`)

**Events:**
- `on:confirm`: Fired when confirmed
- `on:cancel`: Fired when cancelled

---

## Layout & Navigation

### Root Layout (`+layout.svelte`)

The root layout provides the app shell including the top navigation bar, toast container, and auth state management.

#### Top Navigation Bar

The sticky top nav includes:
- **Logo/Title**: App name (configurable via `PUBLIC_APP_NAME`) links to dashboard
- **Dashboard Link**: Always visible for authenticated users
- **User Avatar Menu**: Dropdown with profile, role-based links, and sign out

The user dropdown menu contains:
- My Profile (link to `/account`)
- Commissioner Dashboard (visible for commissioner/admin roles)
- Admin Dashboard (visible for admin role only)
- Sign Out

#### Responsive Behavior

The navigation stays on a single line on mobile with slightly reduced sizing. Elements scale down at breakpoints:
- **768px**: Smaller padding, fonts, and avatar
- **400px**: Further size reduction for narrow screens

#### User Info Fetching

The layout fetches user profile data (role, username, avatar) reactively:

```typescript
let userInfoFetched = false
let mounted = false

// Reactively fetch when session becomes available
$: if (mounted && session && !userInfoFetched) {
  fetchUserInfo()
}

// Reset when session changes
$: if (!session) {
  userInfoFetched = false
  // Clear user state...
}
```

This pattern ensures the avatar and user info load correctly even when:
- The session arrives after initial SSR hydration
- The user signs in (triggers re-fetch for new user)
- The user signs out (clears state)

#### Auth State Change Handling

The layout listens for Supabase auth events:
- `SIGNED_IN`: Logs success, resets `userInfoFetched` to trigger profile fetch
- `SIGNED_OUT`: Logs logout
- `TOKEN_REFRESHED`: Logs refresh with expiry info, invalidates auth

---

## State Management

### Toast Store

Global toast notification system.

```typescript
import { toasts } from '$lib/stores/toast'

// Show notifications
toasts.success('Changes saved!')
toasts.error('Something went wrong')
toasts.info('Pick deadline approaching')
toasts.warning('Some games are locked')

// With custom duration (ms)
toasts.success('Saved!', 2000)

// Remove a specific toast
toasts.remove(toastId)
```

Default duration is 4000ms. Toasts auto-dismiss and can be manually closed.

### Reactive Session State

Session and Supabase client are passed through layout data:

```svelte
<script lang="ts">
  export let data
  let { session, supabase } = data
  $: ({ session, supabase } = data)
</script>
```

---

## API Client

The API client provides typed HTTP methods with automatic authentication and logging.

### Usage

```typescript
import { api, ApiError } from '$lib/api/client'

// GET request
const data = await api.get<ResponseType>('/api/endpoint', session?.access_token)

// POST request
const result = await api.post<ResponseType>('/api/endpoint', { body }, token)

// PUT request
await api.put<ResponseType>('/api/endpoint', { body }, token)

// PATCH request
await api.patch<ResponseType>('/api/endpoint', { body }, token)

// DELETE request
await api.delete<ResponseType>('/api/endpoint', token)
```

### Error Handling

The `handleApiError` function shows a toast AND returns the error message. Use different patterns for different error types:

**Page Load Errors** (page can't function without data):
```typescript
// Show both toast AND persistent error card
try {
  const data = await api.get('/api/seasons', token)
} catch (err) {
  error = handleApiError(err, 'Failed to load seasons')
}
```

**Action Errors** (page still works, user can retry):
```typescript
// Show toast only, don't set persistent error
try {
  await api.post('/api/endpoint', data, token)
  toasts.success('Saved!')
} catch (err) {
  handleApiError(err, 'Failed to save')  // Just toast, no assignment
}
```

**ApiError Properties:**
```typescript
if (err instanceof ApiError) {
  console.log(err.status)  // HTTP status code
  console.log(err.message) // Error message from API
}
```

### Features

- Automatic `Authorization: Bearer` header
- JSON serialization/deserialization
- Structured logging of all requests
- Duration tracking
- Network error handling

---

## Authentication

### Overview

Authentication uses Supabase with SSR support. The flow:

1. **hooks.server.ts**: Creates Supabase server client, validates JWT
2. **+layout.server.ts**: Gets session via `safeGetSession()`
3. **+layout.ts**: Creates browser/server client, passes session to pages
4. **+layout.svelte**: Listens for auth state changes, fetches user profile

### Server-Side Validation

```typescript
// hooks.server.ts
event.locals.safeGetSession = async () => {
  // Validates JWT by calling getUser() (not just getSession())
  const { data: { user }, error } = await supabase.auth.getUser()
  if (error) return { session: null, user: null }

  const { data: { session } } = await supabase.auth.getSession()
  return { session, user }
}
```

### Protected Routes

The root `+layout.server.ts` redirects unauthenticated users to `/login` for all non-public routes (public routes: `/login`, `/auth/*`, `/`). This eliminates the need for client-side auth checks in individual pages.

Role-based routes are additionally protected via their own layout server files:

```typescript
// commissioner/+layout.server.ts
import { getRoleFromToken, isCommissioner } from '$lib/utils/auth'

export const load = async ({ locals }) => {
  const { session } = await locals.safeGetSession()

  if (!session) {
    throw redirect(303, '/login')
  }

  const role = getRoleFromToken(session.access_token)
  if (!isCommissioner(role)) {
    throw error(403, 'Commissioner access required')
  }

  return { session }
}
```

### Role Utilities

```typescript
import { getRoleFromToken, isCommissioner, isAdmin } from '$lib/utils/auth'

const role = getRoleFromToken(session?.access_token)

if (isAdmin(role)) {
  // Admin-only functionality
}

if (isCommissioner(role)) {
  // Commissioner or admin
}
```

### Email Verification (PKCE Flow)

After signup, users receive an email with a verification link. The flow uses PKCE (Proof Key for Code Exchange):

1. **Signup** (`/login`): User registers, `emailRedirectTo` is set to `/auth/callback`
2. **Email Link**: Supabase sends verification email with link to `https://supabase.co/auth/v1/verify?token=pkce_...&redirect_to=.../auth/callback`
3. **Supabase Verification**: User clicks link, Supabase verifies the token on their server
4. **Code Exchange**: Supabase redirects to `/auth/callback?code=...`
5. **Session Creation**: The callback endpoint exchanges the code for a session
6. **Redirect**: User is redirected to `/dashboard`, already authenticated

```typescript
// auth/callback/+server.ts
export const GET: RequestHandler = async ({ url, locals: { supabase } }) => {
  const code = url.searchParams.get('code')

  if (code) {
    const { error } = await supabase.auth.exchangeCodeForSession(code)
    if (!error) {
      redirect(303, '/dashboard')
    }
  }

  redirect(303, '/auth/error')
}
```

The `/auth/error` page displays a user-friendly error message with a link back to login.

**Fallback Handling:** The root route (`/`) also checks for a `code` parameter and performs the exchange. This handles edge cases where old emails redirect to the root instead of `/auth/callback`.

---

## Styling

### CSS Variables

The app uses a dark theme with CSS custom properties defined in `styles.css`:

```css
:root {
  --custom-bg-color: #101010;
  --custom-panel-color: #222;
  --custom-color: #fff;
  --custom-color-brand: #24b47e;
  --custom-color-secondary: #9ca3af;
  --custom-border-radius: 8px;
  --custom-border: 1px solid rgba(255, 255, 255, 0.1);
}
```

### Common Classes

```css
/* Cards */
.card { /* Standard card container */ }
.card-link { /* Clickable card */ }

/* Buttons */
.button { /* Base button */ }
.button.primary { /* Primary action */ }
.button.secondary { /* Secondary action */ }
.button.danger { /* Destructive action */ }
.button.block { /* Full width */ }

/* Layout */
.container { /* Page container */ }
.page-header-three-col { /* Three-column header */ }
.page-sm/md/lg/xl/xxl { /* Page width variants */ }

/* Status Badges */
.week-status.draft { /* Orange */ }
.week-status.active { /* Green */ }
.week-status.final { /* Gray */ }

/* Text Utilities */
.text-sm { /* Small text */ }
.opacity-half { /* 50% opacity */ }
```

### Grid System

```css
.container { width: 90%; max-width: 1000px; margin: 0 auto; }
.row { position: relative; }
.col-6 { width: 46%; } /* 50% minus gutters */
.col-12 { width: 96%; }
```

### Responsive Breakpoints

- **540px**: Container adjustments
- **720px**: Column layouts, mobile nav
- **960px**: Full desktop layout

---

## Logging

The frontend includes comprehensive structured logging for debugging.

### Auth Logging

```typescript
import { logAuth, getJwtDebugInfo } from '$lib/utils/authLogger'

logAuth('LOGIN_ATTEMPT', { email: user.email })
logAuth('LOGIN_SUCCESS', {
  email,
  details: { jwt: getJwtDebugInfo(token) }
})
logAuth('JWT_ERROR', { error: 'Token expired' })
```

**Event Types:**
- `LOGIN_ATTEMPT`, `LOGIN_SUCCESS`, `LOGIN_FAILURE`
- `SIGNUP_ATTEMPT`, `SIGNUP_SUCCESS`, `SIGNUP_FAILURE`
- `SESSION_VALID`, `SESSION_INVALID`, `SESSION_EXPIRED`
- `JWT_ERROR`, `TOKEN_REFRESH`, `LOGOUT`

### API Logging

```typescript
import { logApi } from '$lib/utils/logger'

// Logged automatically by API client
logApi('RESPONSE_SUCCESS', {
  method: 'GET',
  endpoint: '/api/seasons',
  status: 200,
  durationMs: 45
})
```

### Domain Logging

```typescript
import { logPicks, logWeek, logUser, logData } from '$lib/utils/logger'

// Picks
logPicks('SUBMIT_SUCCESS', { weekId, pickCount: 14 })

// Week management
logWeek('ACTIVATION_SUCCESS', { weekId, newStatus: 'active' })

// User profile
logUser('INFO_FETCH_SUCCESS', { hasSession: true })

// Data loading
logData('STANDINGS_LOAD_SUCCESS', { seasonId })
```

### Log Format

All logs follow a structured format that appears in container logs:

```
[CATEGORY] [2026-02-07T15:30:00.000Z] [EVENT] key=value key2="string value"
```

Example:
```
[API] [2026-02-07T15:30:00.000Z] [RESPONSE_ERROR] method="PUT" endpoint="/api/picks" status=400 error="Game is locked" durationMs=45
[AUTH] [2026-02-07T15:30:00.000Z] [JWT_ERROR] email=user@example.com error="Token expired" details={"jwt":{"issued_in_future":true}}
```

---

## Types

### Core Types

```typescript
// Season
type Season = {
  id: string
  year: number
  is_active: boolean
  is_postseason: boolean
  number_of_weeks: number
  weeks?: Week[]
}

// Participant (season member with profile info)
type Participant = {
  user_id: string
  username: string | null
  avatar_url: string    // full URL (default avatar if none set)
  joined_at: string
}

// Week
type Week = {
  id: string
  season_id: string
  number: number
  status: 'draft' | 'games_imported' | 'spreads_set' | 'active'
        | 'played' | 'picks_results_calculated' | 'scored' | 'final'
  activated_at: string | null
  closed_at: string | null
  year?: number
  is_postseason?: boolean
  games: Game[]
}

// Game
type Game = {
  id: string
  home_team_id: string
  away_team_id: string
  home_score: number | null
  away_score: number | null
  home_spread: number | null
  kickoff_time: string
  status: string
  home_team_name: string
  home_team_city: string
  home_team_abbr?: string
  home_team_logo_url: string
  away_team_name: string
  away_team_city: string
  away_team_abbr?: string
  away_team_logo_url: string
}

// Pick
type Pick = {
  id: string
  user_id: string
  week_id: string
  game_id: string
  selected_team_id: string
  points_earned: number | null
  is_correct: boolean | null
}

// Standing
type Standing = {
  user_id: string
  username: string
  points: number
  rank: number
}

// Settings
type Settings = {
  pick_cutoff_minutes: number
  allow_pick_edits: boolean
  points_per_correct_pick: number
  competition_timezone: string
}

// Badge
type Badge = {
  type: 'previous_week_winner' | 'previous_season_winner' | 'previous_season_loser'
  label: string
}
```

Full type definitions are in `lib/types/index.ts`.

---

## Configuration

### Environment Variables

Create a `.env` file:

```env
PUBLIC_SUPABASE_URL=https://your-project.supabase.co
PUBLIC_SUPABASE_PUBLISHABLE_KEY=your_publishable_key
PUBLIC_API_BASE_URL=https://your-api-domain.com
```

All environment variables must be prefixed with `PUBLIC_` to be available client-side.

### SvelteKit Config

`svelte.config.js` uses the Node adapter for deployment:

```javascript
import adapter from '@sveltejs/adapter-node'

export default {
  kit: {
    adapter: adapter()
  }
}
```

### Vite Config

`vite.config.ts` includes allowed hosts for development:

```typescript
export default defineConfig({
  plugins: [sveltekit()],
  server: {
    allowedHosts: ['pp.pawked.com', 'pool.pawked.com']
  }
})
```

### TypeScript Config

Strict mode is enabled with ES2020+ target. SvelteKit path aliases are configured automatically.

---

## Development

### Commands

```bash
# Install dependencies
npm install

# Start dev server (with hot reload)
npm run dev

# Type check
npm run check

# Type check with watch mode
npm run check:watch

# Build for production
npm run build

# Preview production build
npm run preview
```

### Adding a New Page

1. Create route folder in `src/routes/`
2. Add `+page.svelte` for the UI
3. Add `+page.server.ts` for server-side data loading (optional)
4. Add to protected layout if authentication required

### Adding a New Component

1. Create `.svelte` file in `src/lib/components/`
2. Use TypeScript with `<script lang="ts">`
3. Use scoped styles with `<style>`
4. Export props with `export let propName`

### Code Patterns

**Data Loading:**
```svelte
<script lang="ts">
  import { onMount } from 'svelte'
  import { api } from '$lib/api/client'

  let loading = true
  let error = ''
  let data = null

  onMount(async () => {
    try {
      data = await api.get('/api/endpoint', session?.access_token)
    } catch (err) {
      error = handleApiError(err)
    } finally {
      loading = false
    }
  })
</script>

{#if loading}
  <Spinner />
{:else if error}
  <p class="error">{error}</p>
{:else}
  <!-- Render data -->
{/if}
```

**Form Submission:**
```svelte
<script lang="ts">
  let saving = false

  async function handleSubmit() {
    try {
      saving = true
      await api.post('/api/endpoint', formData, token)
      toasts.success('Saved!')
    } catch (err) {
      handleApiError(err)
    } finally {
      saving = false
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <!-- Form fields -->
  <button type="submit" disabled={saving}>
    {saving ? 'Saving...' : 'Save'}
  </button>
</form>
```

---

## Deployment

### Docker

The project includes a `Dockerfile` for containerized deployment:

```bash
# Build image
docker build -t sendyourpicks-frontend .

# Run container
docker run -p 3000:3000 \
  -e PUBLIC_SUPABASE_URL=... \
  -e PUBLIC_SUPABASE_PUBLISHABLE_KEY=... \
  -e PUBLIC_API_BASE_URL=... \
  -e PUBLIC_APP_NAME="Send Your Picks" \
  sendyourpicks-frontend
```

### Node.js

```bash
# Build
npm run build

# Run (production)
node build
```

The app listens on port 3000 by default. Set the `PORT` environment variable to change.

### Static Assets

Static files are served from the `/static` directory. Team logos should be placed in `/static/images/team-logos/`.

---

## Backend Integration

The frontend communicates with the Go backend API:

**Base URL:** Configured via `PUBLIC_API_BASE_URL`

**Authentication:** All requests include `Authorization: Bearer {jwt_token}`

**Key Endpoints Used:**

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/whoami` | GET | Current user profile |
| `/api/account` | GET/PUT | User account management |
| `/api/seasons` | GET | List seasons |
| `/api/seasons/active` | GET | Get active season |
| `/api/seasons/{id}` | GET | Season details |
| `/api/seasons/{id}/participants` | GET | List season participants |
| `/api/seasons/{id}/standings` | GET | Season standings |
| `/api/weeks/{id}` | GET | Week details with games |
| `/api/weeks/{id}/picks` | GET/PUT | User picks |
| `/api/commissioner/seasons/{id}/activate` | PATCH | Activate season |
| `/api/commissioner/seasons/{id}/deactivate` | PATCH | Deactivate season |
| `/api/commissioner/seasons/{id}/weeks-count` | PATCH | Update number of weeks |
| `/api/commissioner/seasons/{id}/participants` | POST | Add participants to season |
| `/api/commissioner/seasons/{id}/participants/{uid}` | DELETE | Remove participant |
| `/api/commissioner/weeks/{id}/activate` | POST | Activate week |
| `/api/commissioner/weeks/{id}/spreads` | PUT | Set spreads |
| `/api/admin/users` | GET | List all users |
| `/api/settings` | GET/PUT | Pool settings |

---

## License

Send Your Picks - NFL picks-against-the-spread tracker
