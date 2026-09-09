package service

import (
	"context"
	"verschlechtert/backend/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
}

type UserService struct {
	userRepo UserRepository
}

func (s *UserService) Upsert(ctx context.Context, firebaseUID, displayName, photoURL, email, locale string) (*domain.User, error) {
	user, err := domain.NewUser(firebaseUID, displayName, photoURL, email, locale)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateLocale(ctx context.Context, user *domain.User, locale string) error {
	user.UpdateLocale(locale)

	err := s.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) Deactivate(ctx context.Context, user *domain.User) error {
	user.Deactivate()

	err := s.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}
