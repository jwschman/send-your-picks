# Route Structure

## Svelte Routes

### Public / Auth

- `/` - Home / landing page
- `/login` - Login page
- `/auth/error` - Auth error page
- `/dashboard` - Main user dashboard with links to admin or commissioner dashboards based on role
- `/settings` - Global settings display
- `/whoami` - Displays whoami info (user id, role, username)

### User Pages

- `/account` - Info about the logged-in user, lets you edit it
- `/users` - List of all users
- `/users/[userid]` - Public info about a specific user
- `/seasons` - Lists all seasons with links
- `/seasons/[seasonid]` - Info about a season, displays weeks and related info
- `/seasons/[seasonid]/standings` - Most recent standings for the season
- `/seasons/[seasonid]/points` - Per-week points breakdown for the season
- `/seasons/[seasonid]/winners` - Week winners for the season
- `/seasons/[seasonid]/weeks/[weekid]` - Info about a specific week, shows games
- `/seasons/[seasonid]/weeks/[weekid]/picks` - View and edit picks
- `/seasons/[seasonid]/weeks/[weekid]/allpicks` - View all users' locked picks for a week
- `/seasons/[seasonid]/weeks/[weekid]/standings` - Season standings after a specific week
- `/seasons/[seasonid]/weeks/[weekid]/results` - Points and ranking of all users for a specific week

### Commissioner

Commissioner pages create, manage, and monitor shared resources.

- `/commissioner` - Commissioner dashboard
- `/commissioner/seasons` - All seasons and their status
- `/commissioner/seasons/new` - Create a new season
- `/commissioner/seasons/[seasonid]` - Info about a specific season, shows all weeks and their status
- `/commissioner/seasons/[seasonid]/manage` - Manage season (activate, deactivate, advance state)
- `/commissioner/seasons/[seasonid]/participants` - Manage season participants
- `/commissioner/seasons/[seasonid]/weeks/[weekid]` - Info about a specific week, advance week state
- `/commissioner/seasons/[seasonid]/weeks/[weekid]/picks` - List of all users and their pick status
- `/commissioner/seasons/[seasonid]/weeks/[weekid]/spreads` - Set and edit spreads for games in a week

### Admin

Admin pages manage users and system-level concerns.

- `/admin` - Admin dashboard
- `/admin/users` - User management
- `/admin/settings` - Edit global settings

## API Routes

### User Routes (any authenticated user)

#### Account & Users
- `GET /api/whoami` - Whoami data (user id, role, username)
- `GET /api/account` - Get my account info
- `PUT /api/account` - Update my account info
- `GET /api/users` - List of all users
- `GET /api/users/:user_id` - Public info about a single user
- `GET /api/badges` - Badges for all users (based on active season)

#### Teams
- `GET /api/teams` - Lists all active teams

#### Seasons
- `GET /api/seasons` - List all seasons
- `GET /api/seasons/active` - Get the current active season
- `GET /api/seasons/:season_id` - Get metadata and weeks for a season
- `GET /api/seasons/:season_id/weeks/active` - Get the active week in a season
- `GET /api/seasons/:season_id/participants` - List users participating in a season

#### Weeks
- `GET /api/weeks/:week_id` - Get metadata and games for a week

#### Picks
- `PUT /api/weeks/:week_id/picks` - Submit/update my picks for a week
- `GET /api/weeks/:week_id/picks` - Get my picks for a week
- `GET /api/weeks/:week_id/picks/summary` - Summary of my picks (e.g. 10 of 13 made, complete or not)
- `POST /api/weeks/:week_id/picks/lock` - Lock all my picks for a week
- `GET /api/weeks/:week_id/picks/locked` - Get all locked picks for all users for a week

#### Points, Standings & Results
- `GET /api/weeks/:week_id/results` - Points and rankings for a single week
- `GET /api/weeks/:week_id/standings` - Season standings snapshot after a given week
- `GET /api/seasons/:season_id/points` - My per-week points and standings for a season
- `GET /api/seasons/:season_id/standings` - Latest standings for the season
- `GET /api/seasons/:season_id/standings/me` - My current standings for the season
- `GET /api/seasons/:season_id/standings/history` - Point and ranking history for charting
- `GET /api/seasons/:season_id/week-winners` - Who won each week (with ties)
- `GET /api/seasons/:season_id/win-counts` - User win/tie counts for the season

#### Misc
- `GET /api/settings` - Get global settings
- `GET /health` - Health check

### Commissioner Routes (commissioner or admin role)

#### Season Management
- `POST /api/commissioner/seasons` - Create a new season
- `POST /api/commissioner/seasons/:season_id/advance` - Advance season state (state machine handles all transitions)
- `PATCH /api/commissioner/seasons/:season_id/activate` - Set a season as the active season
- `PATCH /api/commissioner/seasons/:season_id/deactivate` - Deactivate the active season
- `PATCH /api/commissioner/seasons/:season_id/weeks-count` - Correct the number of weeks in a season

#### Participant Management
- `POST /api/commissioner/seasons/:season_id/participants` - Add user(s) to a season
- `DELETE /api/commissioner/seasons/:season_id/participants/:user_id` - Remove a user from a season

#### Week Management
- `PUT /api/commissioner/weeks/:week_id/spreads` - Set/update spreads for games in a week
- `POST /api/commissioner/weeks/:week_id/spreads/auto-import` - Auto-import spreads from Odds API
- `POST /api/commissioner/weeks/:week_id/activate` - Activate a week for picks

#### Pick Management
- `GET /api/commissioner/weeks/:week_id/picks` - Pick summaries per user for a week (none/partial/complete + timestamp)

### Admin Routes (admin role only)

- `GET /api/admin/users` - List all user accounts
- `PUT /api/admin/settings` - Update global settings
