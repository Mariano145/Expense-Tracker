# Spendly

**Personal expense tracking, done privately.** Spendly is a multi-user app
to record your income and expenses, organize them by category, and see where
your money actually goes — synced across your devices, with your data living
on your own server.

## What it does

| Area | Status |
|------|--------|
| Accounts — register, login, sessions that survive restarts, theft-proof token rotation | ✅ Ready |
| Expenses & categories — record income/spending, organize by category | 🚧 Next |
| Dashboards — monthly summaries, charts, filters | 🔜 Planned |
| Android app | 🔜 Planned |
| Hosted web app | 🔜 Planned |

Currently Spendly runs as a **self-hosted web app** (API + web + database).
Public Android/web releases will land here once they exist; in the meantime,
to run your own instance see **[docs/setup.md](docs/setup.md)**.

## Tech stack

| Layer     | Technology                             |
|-----------|----------------------------------------|
| Backend   | Go (net/http, no framework)            |
| Frontend  | Next.js 16 (App Router, TypeScript)    |
| Database  | PostgreSQL 16                          |
| Auth      | JWT + bcrypt, rotating refresh tokens  |
| State     | Zustand                                |
| Testing   | testify (Go), Vitest + RTL (JS)        |
| CI/CD     | GitHub Actions                         |

## Documentation

Everything precise lives in [`docs/`](docs/):

- **[docs/setup.md](docs/setup.md)** — install, run, test, env vars, Makefile
- **[docs/architecture.md](docs/architecture.md)** — layers, patterns, security model, conventions
- **[docs/api/auth.md](docs/api/auth.md)** — every endpoint with request/response examples and error codes
- **[openspec/](openspec/)** — specs and design docs per feature (SDD artifacts, including the full authentication spec)

## Project layout

```
spendly/
├── backend/       # Go API
├── frontend/      # Next.js app
├── docs/          # setup, architecture, API reference
├── openspec/      # feature specs & designs
└── .github/       # CI
```

## Contributing

Clone a feature branch off `develop`, follow Conventional Commits, keep tests
green (`go test -p 1 ./...`, `npm test`). Details in
[docs/architecture.md](docs/architecture.md#conventions).

## License

MIT — see [LICENSE](./LICENSE)
