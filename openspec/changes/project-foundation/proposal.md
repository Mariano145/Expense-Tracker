# Proposal: Project Foundation

## Intent

Establish the initial project scaffolding for **billetera**, an expense tracking web app. This change creates the monorepo structure, development environment, CI pipeline, and testing infrastructure so that feature work (auth, expense CRUD, dashboards) can begin on solid ground.

## Scope

### In Scope
- Monorepo layout with `backend/` (Go) and `frontend/` (Next.js) directories
- Go backend: `net/http` server, basic routing, health check endpoint (`GET /health`)
- Next.js frontend: App Router, root layout, placeholder home page
- PostgreSQL via Docker Compose for local development
- Database migration system using `golang-migrate/migrate`
- Testing frameworks configured: `testify` (Go), Vitest + React Testing Library (frontend)
- GitHub Actions CI pipeline: lint, test, build for both stacks
- README with setup instructions and environment variable reference

### Out of Scope
- Authentication implementation (JWT, bcrypt, login/register endpoints)
- Expense/category CRUD API or database schema
- Dashboard UI, charts, or data visualization
- Production deployment configuration (Dokku, Hetzner)
- OAuth integration
- Seed data or predefined categories

## Capabilities

### New Capabilities
None — this is pure scaffolding and infrastructure. No product-level behavior is introduced.

### Modified Capabilities
None — no existing specs or capabilities are altered.

## Approach

1. **Backend**: Initialize Go module in `backend/`, create `cmd/server/main.go` with a minimal `net/http` server and `GET /health` handler. Add `testify` dependency and a sample test.
2. **Frontend**: Bootstrap Next.js in `frontend/` with TypeScript, App Router, and Tailwind CSS. Add Vitest and React Testing Library. Create a root layout and placeholder page.
3. **Database & Migrations**: Add `docker-compose.yml` with PostgreSQL service. Install `golang-migrate` CLI via Docker or binary. Create an initial migration folder structure.
4. **CI**: Add `.github/workflows/ci.yml` with jobs for Go (lint, test, build) and Node.js (lint, test, build).
5. **Docs**: Write `README.md` covering prerequisites, local setup (`docker-compose up`), running migrations, and running tests.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/` | New | Go module, entrypoint, router, health handler, tests |
| `frontend/` | New | Next.js app, layout, placeholder page, test config |
| `docker-compose.yml` | New | PostgreSQL + optional backend/frontend services |
| `backend/migrations/` | New | Migration directory + initial baseline |
| `.github/workflows/ci.yml` | New | Lint, test, build pipeline |
| `README.md` | New | Setup instructions and env var reference |
| `openspec/config.yaml` | Modified | Update context with detected stack |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Go module path confusion in monorepo | Low | Use explicit module path (`github.com/user/billetera/backend`) and document it |
| Next.js + Docker Compose hot-reload issues | Low | Bind-mount source with `node_modules` volume override; document workaround |
| golang-migrate not available locally | Low | Provide Docker-based migrate command in README |
| CI setup fails on first push | Med | Test workflow locally with `act` or validate YAML before merge |

## Rollback Plan

This is an additive, greenfield change. Rollback is a full revert:
1. Revert the commit that introduced the scaffolding.
2. Delete untracked files (`backend/`, `frontend/`, `docker-compose.yml`, `.github/`).
3. Confirm `git status` shows a clean working tree.

No production data or users are affected.

## Dependencies

- Docker and Docker Compose installed locally
- Go 1.22+ installed locally (for backend development outside containers)
- Node.js 20+ installed locally (for frontend development outside containers)

## Success Criteria

- [ ] `docker-compose up` starts PostgreSQL without errors
- [ ] `go test ./...` in `backend/` passes (including health handler test)
- [ ] `npm test` in `frontend/` passes (including sample component test)
- [ ] GitHub Actions CI workflow runs and passes on the feature branch
- [ ] `README.md` instructions allow a new contributor to run the stack in under 10 minutes
- [ ] `openspec/config.yaml` stack context is updated
