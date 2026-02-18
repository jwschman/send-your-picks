# commissioner

Commissioner-only routes for managing seasons and weeks.

## Routes

- `/commissioner` - Dashboard
- `/commissioner/seasons` - Season list
- `/commissioner/seasons/new` - Create a new season
- `/commissioner/seasons/{seasonID}` - Season detail with week list and actions
- `/commissioner/seasons/{seasonID}/weeks/new` - Create next week
- `/commissioner/seasons/{seasonID}/weeks/{weekID}` - Week management (import games, set spreads, activate, advance state)
- `/commissioner/seasons/{seasonID}/weeks/{weekID}/import` - Preview and import games
- `/commissioner/seasons/{seasonID}/weeks/{weekID}/edit` - Set spreads
