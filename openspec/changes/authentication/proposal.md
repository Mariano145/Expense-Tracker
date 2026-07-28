# Proposal: User Authentication (JWT + bcrypt)

## Intent

Add secure user authentication so each expense tracker account is private. Users register with email and password, receive short-lived JWT access tokens and long-lived refresh tokens, and can access protected resources. This unblocks multi-user support and personal data isolation.

## Scope

### In Scope
- DB migrations: `users` and `refresh_tokens` tables
- Backend: 3-layer architecture (handler → service → repository)
- Backend: 5 auth endpoints (`register`, `login`, `refresh`, `logout`, `me`)
- Backend: JWT validation middleware + CORS config
- Frontend: Zustand auth store, Axios client with interceptors
- Frontend: Login/Register pages, protected routes via `middleware.ts`, auto-refresh logic
- Tests: unit tests for JWT/bcrypt/handlers/services + integration tests for auth flow
- Docs: README + ARQUITECTURA.md updates

### Out of Scope
- OAuth, password reset, email verification
- Rate limiting, CSRF tokens (SameSite=Strict sufficient)
- Multi-factor authentication

## Capabilities

### New Capabilities
- `user-auth`: Email/password registration, login, logout, token refresh, and protected session access
- `token-management`: JWT access (15 min) + refresh token (7 days) with family-based rotation and theft detection

### Modified Capabilities
- None

## Approach

Implement a 3-layer Go backend with `golang-jwt/jwt/v5` (HS256) and `bcrypt` (cost 12). Access tokens live in-memory/Zustand; refresh tokens are httpOnly Secure cookies stored hashed in PostgreSQL with family-based rotation (reuse revokes the entire family). Frontend uses Axios interceptors to auto-refresh expired tokens. All endpoints under `/api/auth/*`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/handler/` | New | Auth HTTP handlers |
| `backend/internal/service/` | New | Auth business logic |
| `backend/internal/repository/` | New | DB queries for users and tokens |
| `backend/internal/middleware/` | New | JWT extraction/validation, CORS |
| `backend/internal/model/` | New | User and RefreshToken domain types |
| `backend/migrations/` | New | Users and refresh_tokens tables |
| `frontend/src/app/(auth)/` | New | Login and register pages |
| `frontend/src/store/` | New | Zustand auth store |
| `frontend/src/lib/` | New | Axios API client with interceptors |
| `frontend/src/middleware.ts` | New | Route protection |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Token theft via XSS | Low | Access token in-memory only; refresh httpOnly Secure SameSite=Strict |
| Refresh token replay | Low | Family-based rotation revokes family on reuse |
| Bcrypt cost slows login | Low | Cost 12 is standard; monitor p99 latency |

## Rollback Plan

1. Revert migrations (`migrate down` for users + refresh_tokens).
2. Remove auth middleware from protected routes (restore public access temporarily if needed).
3. Delete auth-related backend packages and frontend pages/store.
4. Re-deploy previous frontend build without `middleware.ts`.

## Dependencies

- `golang-jwt/jwt/v5`, `bcrypt` (Go stdlib extensions)
- PostgreSQL running with migration tooling

## Success Criteria

- [ ] User can register with email + password
- [ ] User can login and receive tokens
- [ ] Protected endpoints require valid access token
- [ ] Access token auto-refreshes via refresh cookie when expired
- [ ] Logout invalidates refresh token
- [ ] All unit and integration tests pass
- [ ] Frontend auth flow end-to-end works
- [ ] README and ARQUITECTURA.md updated
