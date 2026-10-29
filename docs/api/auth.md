# Auth API Reference

Base URL: `http://localhost:8080`

All request and response bodies are `application/json`. Authenticated
endpoints expect `Authorization: Bearer <access_token>`.

## Common Response Shapes

Success responses return the resource directly. Errors always use:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid email or password"
  }
}
```

| HTTP | `error.code`   | When                                        |
|------|----------------|---------------------------------------------|
| 400  | `BAD_REQUEST`  | Malformed JSON body, invalid email, weak password |
| 401  | `UNAUTHORIZED` | Bad credentials, expired/revoked/missing token |
| 409  | `CONFLICT`     | Email already registered                    |
| 500  | `INTERNAL_ERROR` | Unexpected server failure                 |

## Cookies

`POST /login` and `POST /refresh` set a `refresh_token` cookie:

| Attribute  | Value                                     |
|------------|-------------------------------------------|
| HttpOnly   | `true` (JS cannot read it)                |
| Secure     | `true`                                    |
| SameSite   | `Strict`                                  |
| Path       | `/api/auth`                               |
| Max-Age    | 604800 (7 days)                           |

The access token (JWT, HS256) lives 15 minutes and is returned in the body —
the client keeps it in memory / sessionStorage, never a persistent cookie.

---

## POST /api/auth/register

Create a new account.

**Request**

```json
{ "email": "user@example.com", "password": "supersecret123" }
```

Validation (rejected before any DB work):
- `email` must match a basic `local@domain.tld` shape → else `400`.
- `password` must be ≥ 8 characters → else `400`.

**201 Created**

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "email": "user@example.com",
  "created_at": "2026-09-27T01:14:22Z"
}
```

**Errors:** `400` invalid email/password, `409` email already registered.

```bash
curl -i -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"supersecret123"}'
```

---

## POST /api/auth/login

Authenticate and start a session.

**Request**

```json
{ "email": "user@example.com", "password": "supersecret123" }
```

**200 OK** (plus `Set-Cookie: refresh_token=...`)

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7", "email": "user@example.com" }
}
```

**Errors:** `400` invalid email format / empty password, `401` wrong
credentials (same generic message for unknown email and wrong password —
no field disclosure).

```bash
curl -i -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"supersecret123"}'
```

---

## POST /api/auth/refresh

Exchange the refresh cookie for a new access token. The presented refresh
token is revoked and replaced by a new one in the same family (rotation).

**Request:** no body. Requires the `refresh_token` cookie (send credentials).

**200 OK** (plus a new `Set-Cookie: refresh_token=...`)

```json
{ "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

**Errors:** `401` missing cookie, expired token, or revoked token.

**Theft detection:** if a previously-revoked refresh token is presented again,
the entire token family is revoked and the response is `401`. The legitimate
user must log in again.

```bash
curl -i -X POST http://localhost:8080/api/auth/refresh \
  --cookie 'refresh_token=<value>'
```

---

## POST /api/auth/logout

Revoke the current refresh-token family and clear the cookie.

**Request:** no body. Requires the `refresh_token` cookie.

**204 No Content** (plus `Set-Cookie: refresh_token=; Max-Age=-1`)

**Errors:** `401` missing cookie / unknown token.

```bash
curl -i -X POST http://localhost:8080/api/auth/logout \
  --cookie 'refresh_token=<value>'
```

---

## GET /api/auth/me

Return the profile of the bearer of a valid access token. Protected by auth
middleware.

**Request**

```
Authorization: Bearer <access_token>
```

**200 OK**

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "email": "user@example.com"
}
```

**Errors:** `401` missing/expired/tampered Bearer token.

```bash
curl -i http://localhost:8080/api/auth/me \
  -H 'Authorization: Bearer <access_token>'
```

---

## Token Lifecycle (summary)

```
register ──(no tokens issued)
login ─────→ access_token (15m) + refresh_token cookie (7d, family F)
refresh ───→ old refresh revoked, new refresh issued (same family F)
             reuse of a revoked token ⇒ family F fully revoked ⇒ re-login
logout ────→ family F revoked, cookie cleared
```
