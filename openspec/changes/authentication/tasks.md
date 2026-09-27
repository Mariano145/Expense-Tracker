# Tasks: User Authentication (JWT + bcrypt)

## Review Workload Forecast

Estimated changed lines: ~2,100. 400-line budget risk: High. Chained PRs recommended: Yes.

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units (base = `develop`)

| Unit | Goal | PR |
|------|------|----|
| 1 | Backend foundation | PR 1 |
| 2 | Repository layer | PR 2 |
| 3 | Service layer (size:exception ~450) | PR 3 |
| 4 | HTTP layer | PR 4 |
| 5 | Backend wiring | PR 5 |
| 6 | Frontend (size:exception ~530) | PR 6 |
| 7 | Documentation | PR 7 |

Note: Next.js 16.2.11 differs from training-data. Implementer MUST consult installed docs under `frontend/node_modules/next/dist/` before writing `middleware.ts`.

## Phase 1: Backend Foundation

- [x] 1.1 Create `backend/migrations/000002_create_users.{up,down}.sql` (table, email index, `updated_at` trigger).
- [x] 1.2 Create `backend/migrations/000003_create_refresh_tokens.{up,down}.sql` (table, 3 indexes).
- [x] 1.3 Create `backend/internal/config/config.go` (env loading, defaults, validation).
- [x] 1.4 Create `backend/internal/database/postgres.go` (pgxpool factory).
- [x] 1.5 Create `backend/internal/model/{user.go, refresh_token.go, errors.go}`.
- [x] 1.6 Update `backend/.env.example` (JWT_ACCESS_EXPIRY, JWT_REFRESH_EXPIRY, BCRYPT_COST).

## Phase 2: Repository Layer

- [x] 2.1 Implement `UserRepository` and `RefreshTokenRepository` in `internal/repository/`.
- [x] 2.2 Repository tests against a Postgres test DB.

## Phase 3: Service Layer

- [x] 3.1 Add `internal/auth/{jwt.go, password.go, tokens.go}` (HS256, bcrypt 12, SHA-256).
- [x] 3.2 Implement `AuthService` (Register, Login, Refresh with family rotation + reuse detection, Logout, ValidateToken).
- [x] 3.3 Service unit tests with mocks covering every scenario in `specs/user-auth.md` and `specs/token-management.md`.

## Phase 4: HTTP Layer

- [x] 4.1 Create `internal/handler/{dto.go, auth.go}` (`AuthHandler` + 5 endpoints).
- [x] 4.2 Create `internal/middleware/{auth.go, cors.go}` (Bearer validate, user context, CORS).
- [x] 4.3 Handler and middleware tests (valid, expired, tampered, missing).

## Phase 5: Backend Wiring

- [x] 5.1 Update `backend/cmd/server/main.go` (config → pgxpool → DI → graceful shutdown).
- [x] 5.2 Update `backend/internal/router/router.go` (register `/api/auth/*`, apply middleware).
- [x] 5.3 Add `backend/internal/integration/auth_flow_test.go` (end-to-end).

## Phase 6: Frontend Foundation

- [x] 6.1 Add `zustand` and `axios` to `frontend/package.json`.
- [x] 6.2 Create `frontend/src/lib/api.ts` (Axios, `withCredentials`, `baseURL`).
- [x] 6.3 Interceptors: request (Bearer) + response (on 401 → refresh once → retry; on fail → logout + redirect).
- [x] 6.4 Create `frontend/src/store/auth-store.ts` (Zustand + persist).

## Phase 7: Frontend Pages

- [x] 7.1 Create `frontend/src/app/(auth)/{layout.tsx, login/page.tsx, register/page.tsx}`.
- [x] 7.2 Create `frontend/src/app/(dashboard)/{layout.tsx, page.tsx}` (protected).
- [x] 7.3 Create `frontend/src/proxy.ts` (route protection redirects — Next.js 16 renamed middleware→proxy).

## Phase 8: Frontend Tests

- [ ] 8.1 Vitest tests for `auth-store.ts` (state, persist).
- [ ] 8.2 RTL tests for login and register pages.
- [ ] 8.3 `middleware.ts` test + auto-refresh integration test (mock axios).

## Phase 9: Documentation

- [ ] 9.1 Update `README.md` (setup, env vars, migration, test user).
- [ ] 9.2 Update `ARQUITECTURA.md` (auth components, 3-layer diagram).
- [ ] 9.3 Add `docs/api/auth.md` (endpoint reference, examples, error codes).
