# Token Management Specification

## Purpose

Define JWT access token and refresh token lifecycle, validation, rotation, and theft detection.

## Requirements

### Requirement: Access Token Generation

The system MUST issue short-lived JWT access tokens upon successful authentication.

- **Acceptance Criteria**: JWT is signed with HS256, contains user ID and email, expires in 15 minutes.
- **Edge Cases**: Clock skew, empty claims.

#### Scenario: Token issued at login

- GIVEN a successful login
- WHEN the session is established
- THEN the system returns a JWT access token signed with HS256
- AND the token MUST expire in 15 minutes

#### Scenario: Token issued at refresh

- GIVEN a valid refresh token
- WHEN the user requests token refresh
- THEN the system returns a new JWT access token with a fresh 15-minute expiry

### Requirement: Access Token Validation

The system MUST verify the access token on protected endpoints.

- **Acceptance Criteria**: Valid token permits access; expired, tampered, or missing token returns 401.
- **Edge Cases**: Token with wrong signing key, token with missing required claims.

#### Scenario: Valid token

- GIVEN a request with a valid, unexpired access token in the Authorization header
- WHEN the request reaches a protected endpoint
- THEN the endpoint processes the request and extracts the user identity

#### Scenario: Expired token

- GIVEN a request with an expired access token
- WHEN the request reaches a protected endpoint
- THEN the system responds with HTTP 401 Unauthorized
- AND the error indicates the token has expired

#### Scenario: Tampered token

- GIVEN a request with an invalid signature or malformed JWT
- WHEN the request reaches a protected endpoint
- THEN the system responds with HTTP 401 Unauthorized

### Requirement: Refresh Token Generation and Storage

The system MUST issue long-lived refresh tokens stored securely server-side.

- **Acceptance Criteria**: Token is a cryptographically random string; stored as SHA-256 hash; 7-day expiry.
- **Edge Cases**: Token collision (extremely unlikely), database storage failure.

#### Scenario: Token creation

- GIVEN a successful login
- WHEN the session is established
- THEN the system generates a cryptographically random refresh token
- AND stores a SHA-256 hash in PostgreSQL with a 7-day expiry and a family identifier

### Requirement: Refresh Token Rotation

The system MUST rotate refresh tokens on every use.

- **Acceptance Criteria**: Each refresh produces a new access token and a new refresh token in the same family; old refresh token is invalidated.
- **Edge Cases**: Concurrent refresh requests, network interruption after token issuance.

#### Scenario: Successful rotation

- GIVEN a valid refresh token cookie
- WHEN the user requests POST /api/auth/refresh
- THEN the system issues a new access token
- AND replaces the old refresh token with a new one belonging to the same family
- AND sets the new refresh token in an httpOnly Secure cookie

#### Scenario: Reused rotated token

- GIVEN a refresh token that was already rotated and invalidated
- WHEN a request uses that old token
- THEN the system revokes the entire token family
- AND responds with HTTP 401 Unauthorized

### Requirement: Token Theft Detection

The system MUST detect and mitigate refresh token replay attacks via family revocation.

- **Acceptance Criteria**: Reuse of an already-rotated refresh token triggers family revocation, forcing re-authentication.
- **Edge Cases**: Attacker reuse after legitimate rotation, network retry causing false positive.

#### Scenario: Family revocation on reuse

- GIVEN an attacker possesses a stolen refresh token
- WHEN the attacker uses the token after the legitimate user has rotated it
- THEN the system revokes the entire token family
- AND all refresh tokens for that user family become invalid
- AND the user MUST re-authenticate

#### Scenario: Legitimate user after theft detection

- GIVEN the token family has been revoked due to detected reuse
- WHEN the legitimate user tries to refresh with their latest token
- THEN the system responds with HTTP 401 Unauthorized
- AND the user MUST log in again

### Requirement: Token Expiration Handling

The system MUST reject expired refresh tokens.

- **Acceptance Criteria**: Refresh tokens past 7 days are rejected; cleanup of expired tokens MAY run periodically.
- **Edge Cases**: Token expires mid-request, system clock drift.

#### Scenario: Expired refresh token

- GIVEN a refresh token past its 7-day expiry
- WHEN the user attempts to refresh
- THEN the system responds with HTTP 401 Unauthorized
- AND the user MUST re-authenticate

#### Scenario: Cleanup of stale tokens

- GIVEN refresh tokens expired more than 30 days ago
- WHEN a background cleanup task runs (if implemented)
- THEN those token records MAY be deleted from the database
