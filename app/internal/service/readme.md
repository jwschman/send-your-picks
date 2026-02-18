# service

Business logic that doesn't belong in HTTP handlers.

## What goes here

**Read-only loaders:**
- `GetWeekGames`, `GetWeekPicks`, `GetWeekStatus`
- `WeekExists`, `SeasonExists`, `WeekFinal`

**Mutating domain logic:**
- `CalculatePickResults`, `CalculateWeekPoints`, `CalculateSeasonSnapshot`
- `ImportGamesForWeek`, `ImportScoresForWeek`
- `CreateNextWeekForSeason`

**State machine / workflow orchestration:**
- `AdvanceWeekState` - manages week lifecycle through automated and manual states

## Guidelines

These functions:
- Accept a `*sqlx.DB` or `*sqlx.Tx`
- Do not reference Gin, HTTP, or request context
- Return `(result, error)` only
- Handle cross-entity logic that spans multiple tables

---

## Advance Week State Machine (advanceweek.go)

The `AdvanceWeekState` function manages the week lifecycle through automated and manual states.

### Week Status Flow

```
draft ──(auto)──► games_imported ──(manual)──► spreads_set ──(manual)──► active
                  Import games        Set spreads            Activate week

active ──(auto)──► played ──(auto)──► picks_results_calculated ──(auto)──► scored ──(auto)──► final
   Import scores     Calc picks           Calc week points          Calc season standings
```

### Loop Behavior

The state machine loops through automated states until it hits a stopping condition:

- `draft` → imports games → loops to `games_imported`
- `games_imported` → **stops** (manual: commissioner sets spreads)
- `spreads_set` → **stops** (manual: commissioner activates week)
- `active` → imports scores → if all games done, loops to `played`; otherwise **stops** (waiting)
- `played` → calculates pick results → loops to `picks_results_calculated`
- `picks_results_calculated` → calculates week points → loops to `scored`
- `scored` → calculates season standings → loops to `final`
- `final` → **stops** (done)

### Exit Conditions

1. `ErrManualActionRequired` - needs commissioner input (spreads or activation)
2. `ErrWaitingForGames` - week is active but games still in progress
3. `nil` - reached final state successfully
4. Any other `error` - DB error, external API timeout, etc.

### Idempotency

Each state transition updates the database before moving on. If an error occurs mid-loop:
- The current state is preserved in the DB
- Next invocation picks up from where it left off
- Safe to call repeatedly (e.g., from a cron job)

### Usage

Designed to be called from:
- Commissioner dashboard button (manual trigger)
- Kubernetes CronJob running every N minutes during game windows (automated)

## Badges (badges.go)

Week/season winner and previous season loser badges.

## Services (service.go)

Shared business logic queries and operations.
