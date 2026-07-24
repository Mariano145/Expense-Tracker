# Tasks: Project Foundation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1100–1500 (greenfield: Next.js, Go, CI, Docker, docs) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (Backend) → PR 2 (DB) → PR 3 (Frontend) → PR 4 (CI+docs) |
| Delivery strategy | ask-on-risk |
| Chain strategy | feature-branch-chain |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | PR | Notes |
|------|------|----|-------|
| 1 | Backend boots, `/health` 200, `go test` green | PR 1 | Base: `develop`. Runnable alone. |
| 2 | `docker compose up postgres` works, migrate wired | PR 2 | Base: PR 1. |
| 3 | Next.js scaffold, Vitest green, `npm test` passes | PR 3 | Base: PR 2. |
| 4 | CI green; root README enables <10min setup | PR 4 | Base: PR 3. Final merge to `develop`. |

PR 1 targets `develop`; PR 2/3/4 stack on the previous branch.

## Phase 1: Git & Monorepo Foundation

- [x] 1.1 `git init` at root; create empty `main` and `develop` branches
- [x] 1.2 Root `.gitignore` (Go, Node, `.env*`, IDE, OS)
- [x] 1.3 Update `openspec/config.yaml` `context` with detected stack
- [x] 1.4 Root `LICENSE` placeholder (or skip)

## Phase 2: Backend Foundation (Go)

- [x] 2.1 `go mod init github.com/<user>/billetera/backend` in `backend/`
- [x] 2.2 `backend/cmd/server/main.go`: `http.Server` + graceful shutdown, port from env
- [x] 2.3 `backend/internal/handler/health.go`: `GET /health` → 200 `{"status":"ok"}`
- [x] 2.4 `backend/internal/router/router.go`: `http.ServeMux` (Go 1.22+ method patterns)
- [x] 2.5 Add `testify`; `health_test.go` table-driven (200 case)
- [x] 2.6 `backend/.env.example` (`PORT`, `DATABASE_URL`, `JWT_SECRET`, `CORS_ORIGIN`) + `backend/README.md`

## Phase 3: Database & Migrations

- [x] 3.1 Root `docker-compose.yml` with `postgres:16` (named volume, env, healthcheck)
- [x] 3.2 `backend/migrations/` + `README.md` with `migrate` CLI usage
- [x] 3.3 Empty baseline pair `000001_baseline.up.sql` / `.down.sql` (comment-only)
- [x] 3.4 Root `Makefile`: `db-up`, `db-down`, `migrate-up`, `migrate-down` (migrate image)

## Phase 4: Frontend Foundation (Next.js)

- [x] 4.1 `create-next-app` in `frontend/` (TypeScript, App Router, Tailwind, ESLint, `src/`)
- [x] 4.2 Install + configure Vitest, RTL, jsdom; add `vitest.config.ts` + `vitest.setup.ts`
- [x] 4.3 `package.json` scripts: `test`, `test:watch`, `test:coverage`
- [x] 4.4 `frontend/src/components/Hello.tsx` + `Hello.test.tsx` smoke test
- [x] 4.5 `frontend/.env.example` (`NEXT_PUBLIC_API_URL`) + `frontend/Dockerfile` (Node 20 multi-stage)

## Phase 5: CI & Final Documentation

- [x] 5.1 `.github/workflows/ci.yml`: `backend-ci` (vet/test/build) + `frontend-ci` (lint/test/build); trigger on push to `develop` and PRs
- [x] 5.2 Root `README.md`: prereqs, `docker compose up`, migrate, test commands, layout
- [x] 5.3 Local validation: `go test ./...`, `npm test`, `docker compose config` clean
- [x] 5.4 Open PR chain: PR 1 → develop; PR 2/3/4 stacked; final merge to develop
