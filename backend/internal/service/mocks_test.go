package service

import (
	"context"

	"github.com/gentle-ai/billetera/backend/internal/model"
)

// mockUserRepo is a test double for repository.UserRepository.
type mockUserRepo struct {
	createFn     func(ctx context.Context, user *model.User) error
	findByEmailFn func(ctx context.Context, email string) (*model.User, error)
	findByIDFn   func(ctx context.Context, id string) (*model.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	return m.findByIDFn(ctx, id)
}

// mockRefreshTokenRepo is a test double for repository.RefreshTokenRepository.
type mockRefreshTokenRepo struct {
	createFn          func(ctx context.Context, token *model.RefreshToken) error
	findByHashFn      func(ctx context.Context, hash string) (*model.RefreshToken, error)
	revokeByHashFn    func(ctx context.Context, hash string) error
	revokeByFamilyFn  func(ctx context.Context, familyID string) error
	deleteExpiredFn   func(ctx context.Context) (int64, error)
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	return m.createFn(ctx, token)
}

func (m *mockRefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	return m.findByHashFn(ctx, hash)
}

func (m *mockRefreshTokenRepo) RevokeByHash(ctx context.Context, hash string) error {
	return m.revokeByHashFn(ctx, hash)
}

func (m *mockRefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID string) error {
	return m.revokeByFamilyFn(ctx, familyID)
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return m.deleteExpiredFn(ctx)
}
