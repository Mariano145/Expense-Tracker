# Feature: Authentication (living spec)

**Status:** implemented on `feature/authentication`, awaiting human QA + merge to `develop`.
Full historical specs (design, G/W/T scenarios, task log): [`openspec/changes/archive/2026-10-31-authentication/`](../openspec/changes/archive/2026-10-31-authentication/).
API contract reference: [docs/api/auth.md](../api/auth.md).

## What the system does today

- Users register with email + password; passwords bcrypt-hashed (cost 12).
- Register/login reject malformed emails and passwords <8 chars (400) with
  human-readable messages; duplicate email → 409; bad credentials → generic 401.
- Sessions: JWT access token (HS256, 15 min) + rotating refresh token
  (httpOnly Secure SameSite=Strict cookie, 7 days) with family-based rotation
  and theft detection (reuse of a revoked token revokes the whole family).
- Protected routes redirect unauthenticated users to /login; authenticated
  users skip login/register.

## Open known gaps (candidates for next stories)

- [ ] Forms show raw axios errors instead of the API's `error.message` (the 400 story).
- [ ] Session-expiry UX: silent 15-min token lapse relies on auto-refresh only.

## Conventions this area must keep

- Auth changes get an integration test against real Postgres (`-p 1` — shared test DB).
- Token/cookie changes require the theft-detection scenario re-verified by hand.
