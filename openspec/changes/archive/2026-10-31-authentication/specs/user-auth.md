# User Authentication Specification

## Purpose

Define email/password registration, login, logout, and profile access for the expense tracker.

## Requirements

### Requirement: User Registration

The system MUST allow a user to register with a unique email and a password meeting strength rules.

- **Acceptance Criteria**: User is created with bcrypt-hashed password; duplicate emails are rejected; weak passwords are rejected.
- **Edge Cases**: Concurrent requests with same email, malformed email format, empty password.

#### Scenario: Successful registration

- GIVEN a valid email and password with minimum 8 characters
- WHEN the user submits the registration request
- THEN a new user is created with a bcrypt-hashed password (cost 12)
- AND the response returns the user ID and email with HTTP 201

#### Scenario: Duplicate email

- GIVEN an email already exists in the database
- WHEN the user submits registration with that email
- THEN the system responds with HTTP 409 Conflict
- AND no new user is created

#### Scenario: Weak password

- GIVEN a password shorter than 8 characters
- WHEN the user submits registration
- THEN the system responds with HTTP 400 Bad Request
- AND the error message indicates password is too short

#### Scenario: Malformed email

- GIVEN an email without "@" or domain
- WHEN the user submits registration
- THEN the system responds with HTTP 400 Bad Request

### Requirement: User Login

The system MUST validate email and password, then initiate an authenticated session.

- **Acceptance Criteria**: Valid credentials return access token and set refresh cookie; invalid credentials return 401 without field disclosure.
- **Edge Cases**: Non-existent email, wrong password, unregistered user.

#### Scenario: Successful login

- GIVEN a registered user with correct credentials
- WHEN the user submits login
- THEN the system returns an access token (JWT, HS256, 15 min expiry)
- AND sets an httpOnly Secure SameSite=Strict refresh token cookie

#### Scenario: Invalid credentials

- GIVEN incorrect email or password
- WHEN the user submits login
- THEN the system responds with HTTP 401 Unauthorized
- AND MUST NOT reveal whether the email or password is invalid

### Requirement: User Logout

The system MUST invalidate the user's refresh token upon logout.

- **Acceptance Criteria**: Refresh token is revoked in the database and cookie is cleared.
- **Edge Cases**: Logout without cookie, logout with already-revoked token.

#### Scenario: Successful logout

- GIVEN an authenticated user with a valid refresh cookie
- WHEN the user requests logout
- THEN the refresh token is revoked in the database
- AND the refresh cookie is cleared in the response

#### Scenario: Logout without refresh cookie

- GIVEN a logout request with no refresh cookie
- WHEN the request is processed
- THEN the system responds with HTTP 400 Bad Request

### Requirement: Get Current User

The system MUST return the authenticated user's profile.

- **Acceptance Criteria**: Valid access token returns user profile; missing or invalid token returns 401.
- **Edge Cases**: Expired access token, malformed Authorization header.

#### Scenario: Retrieve profile

- GIVEN a request with a valid access token in the Authorization header
- WHEN the user requests GET /api/auth/me
- THEN the system returns the user ID and email with HTTP 200

#### Scenario: Missing or invalid token

- GIVEN a request without a valid access token
- WHEN the user requests GET /api/auth/me
- THEN the system responds with HTTP 401 Unauthorized

### Requirement: Password Validation

The system MUST enforce password strength at registration.

- **Acceptance Criteria**: Minimum 8 characters; no maximum enforced at spec time.
- **Edge Cases**: Empty password, exactly 8 characters, Unicode characters.

#### Scenario: Password too short

- GIVEN a password of 7 characters
- WHEN the user registers
- THEN the system rejects with HTTP 400 Bad Request

#### Scenario: Password exactly 8 characters

- GIVEN a password of exactly 8 characters
- WHEN the user registers
- THEN the system accepts the request

### Requirement: Email Uniqueness

The system MUST ensure no two users share the same email.

- **Acceptance Criteria**: Database unique constraint on users.email; API returns 409 on conflict.
- **Edge Cases**: Case-insensitive duplicates (implementation MAY normalize to lowercase).

#### Scenario: Concurrent duplicate registration

- GIVEN two simultaneous requests with the same new email
- WHEN both requests are processed
- THEN exactly one succeeds with HTTP 201 and the other returns HTTP 409
