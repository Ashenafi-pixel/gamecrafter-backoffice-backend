package otp

import (
	"context"

	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/storage"
)

// UserStorageAdapter adapts the storage.User interface to the OTP module's UserStorage interface
type UserStorageAdapter struct {
	userStorage storage.User
}

// NewUserStorageAdapter creates a new user storage adapter
func NewUserStorageAdapter(userStorage storage.User) UserStorage {
	return &UserStorageAdapter{
		userStorage: userStorage,
	}
}

// GetUserByEmail checks if a user exists by email
func (a *UserStorageAdapter) GetUserByEmail(ctx context.Context, email string) (dto.User, bool, error) {
	return a.userStorage.GetUserByEmail(ctx, email)
}

// CreateUser creates a new user
func (a *UserStorageAdapter) CreateUser(ctx context.Context, user dto.User) (dto.User, error) {
	return a.userStorage.CreateUser(ctx, user)
}

// UpdateUserVerificationStatus updates the user's email verification status
func (a *UserStorageAdapter) UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, isVerified bool) (dto.User, error) {
	return a.userStorage.UpdateUserVerificationStatus(ctx, userID, isVerified)
}
