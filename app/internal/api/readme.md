# api

HTTP layer. Route definitions, request handlers, and middleware.

- `router.go` - All route definitions with auth middleware groups
- `handlers/` - Request handlers grouped by domain
- `middleware/` - Auth (JWT validation, role-based access) and request metrics
