package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gentle-ai/billetera/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshTokenRepository defines the data access contract for RefreshToken entities.
// Tokens are stored as SHA-256 hashes — the repository never sees plain-text tokens.
type RefreshTokenRepository interface {
	// Create persists a new refresh token. The token's ID and CreatedAt
	// fields are populated by the database and written back into the struct.
	Create(ctx context.Context, token *model.RefreshToken) error

	// FindByHash retrieves a refresh token by its SHA-256 hash.
	// Returns model.ErrTokenNotFound if no matching token exists.
	FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error)

	// RevokeByHash marks a single token as revoked by its hash.
	// Returns model.ErrTokenNotFound if no matching token exists.
	RevokeByHash(ctx context.Context, hash string) error

	// RevokeByFamilyID marks all tokens in a family as revoked.
	// Used for theft detection — when a rotated token is reused, the entire
	// family is revoked to force re-authentication.
	RevokeByFamilyID(ctx context.Context, familyID string) error

	// DeleteExpired removes tokens that have expired.
	// Used for periodic cleanup of stale records.
	DeleteExpired(ctx context.Context) (int64, error)
}

// pgxRefreshTokenRepo is the pgx-backed implementation of RefreshTokenRepository.
type pgxRefreshTokenRepo struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepository returns a RefreshTokenRepository backed by the given pool.
func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &pgxRefreshTokenRepo{pool: pool}
}

func (r *pgxRefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, token.UserID, token.TokenHash, token.FamilyID, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *pgxRefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	token := &model.RefreshToken{}
	query := `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	err := r.pool.QueryRow(ctx, query, hash).
		Scan(&token.ID, &token.UserID, &token.TokenHash, &token.FamilyID,
			&token.ExpiresAt, &token.Revoked, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find refresh token by hash: %w", err)
	}
	return token, nil
}

func (r *pgxRefreshTokenRepo) RevokeByHash(ctx context.Context, hash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token_hash = $1
	`
	tag, err := r.pool.Exec(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("revoke refresh token by hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrTokenNotFound
	}
	return nil
}

func (r *pgxRefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE family_id = $1
	`
	_, err := r.pool.Exec(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens by family: %w", err)
	}
	return nil
}

func (r *pgxRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1
	`
	tag, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return tag.RowsAffected(), nil
}
