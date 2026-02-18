# internal

Core application packages. The `internal` directory follows Go's convention
for non-importable packages.

- `api/` - HTTP routes, handlers, and middleware (Gin)
- `db/` - PostgreSQL connection setup (pgx/sqlx)
- `external/` - External API clients (BallDontLie, The Odds API)
- `id/` - ULID generation
- `logger/` - Structured logging (slog wrapper)
- `models/` - Data model definitions
- `service/` - Business logic and state machine
- `settings/` - Global application settings
