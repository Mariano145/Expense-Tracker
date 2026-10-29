// Package repository provides data access implementations for the domain layer.
// Each repository wraps a pgxpool.Pool and maps database errors to domain sentinel errors.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mariano145/spendly/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines the data access contract for User entities.
// The interface allows the service layer to be tested with mock implementations.
type UserRepository interface {
	// Create persists a new user. The user's ID, CreatedAt, and UpdatedAt
	// fields are populated by the database and written back into the struct.
	// Returns model.ErrEmailExists if the email is already taken.
	Create(ctx context.Context, user *model.User) error

	// FindByEmail retrieves a user by their email address.
	// Returns model.ErrUserNotFound if no matching user exists.
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID retrieves a user by their UUID.
	// Returns model.ErrUserNotFound if no matching user exists.
	FindByID(ctx context.Context, id string) (*model.User, error)
}

// pgxUserRepo is the pgx-backed implementation of UserRepository.
type pgxUserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepository returns a UserRepository backed by the given connection pool.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgxUserRepo{pool: pool}
}

func (r *pgxUserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// PostgreSQL unique violation — the email is already taken.
		if isUniqueViolation(err, "users_email_key") {
			return fmt.Errorf("%w: %s", model.ErrEmailExists, user.Email)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *pgxUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.pool.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", model.ErrUserNotFound, email)
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *pgxUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", model.ErrUserNotFound, id)
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
// for the specified constraint name.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraint
	}
	return false
}
