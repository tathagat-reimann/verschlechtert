package user

import (
	"context"
)

type IUserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	Deactivate(ctx context.Context, userID int64) error
}

type UserService struct {
	userRepo IUserRepository
}

func (s *UserService) Upsert(ctx context.Context, firebaseUID, displayName, photoURL, email, locale string) (*User, error) {
	user, err := NewUser(firebaseUID, displayName, photoURL, email, locale)
	if err != nil {
		return nil, err
	}

	persistedUser, err := s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	return persistedUser, nil
}

func (s *UserService) UpdateLocale(ctx context.Context, user *User, locale string) error {
	user.UpdateLocale(locale)

	_, err := s.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Deactivate(ctx context.Context, user *User) error {
	user.Deactivate()

	err := s.userRepo.Deactivate(ctx, user.GetID())
	if err != nil {
		return err
	}

	return nil
}

func NewUserService(userRepo IUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}
