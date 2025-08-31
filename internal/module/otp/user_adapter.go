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
	// Use the real user storage to check if user exists
	return a.userStorage.GetUserByEmail(ctx, email)
}

// CreateUser creates a new user
func (a *UserStorageAdapter) CreateUser(ctx context.Context, user dto.User) (dto.User, error) {
	// Use the real user storage to create user
	createdUser, err := a.userStorage.CreateUser(ctx, user)
	if err != nil {
		return dto.User{}, err
	}
	return createdUser, nil
}

// UpdateUserVerificationStatus updates the user's email verification status
func (a *UserStorageAdapter) UpdateUserVerificationStatus(ctx context.Context, userID uuid.UUID, isVerified bool) (dto.User, error) {
	// Use the real user storage to update verification status
	updatedUser, err := a.userStorage.UpdateUserVerificationStatus(ctx, userID, isVerified)
	if err != nil {
		return dto.User{}, err
	}
	return updatedUser, nil
}
