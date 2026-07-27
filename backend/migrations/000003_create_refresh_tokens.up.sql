-- Refresh tokens are stored as SHA-256 hashes, never as plain text.
-- Family-based rotation enables theft detection: if a revoked token is reused,
-- the entire family is revoked, forcing re-authentication.
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index on user_id for fast lookups when revoking all tokens for a user.
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Index on token_hash for fast lookups during token validation.
-- The UNIQUE constraint already creates an index, but being explicit clarifies intent.
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Index on family_id for family-wide revocation during theft detection.
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
