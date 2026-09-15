package user

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type IUserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	Deactivate(ctx context.Context, userID int64) error
}

type UserService struct {
	userRepo IUserRepository
}

func (s *UserService) Upsert(ctx context.Context, authClient *auth.Client, firebaseUID, displayName, photoURL, email, locale string) (*User, error) {
	user, err := NewUser(firebaseUID, displayName, photoURL, email, locale)
	if err != nil {
		return nil, err
	}

	persistedUser, err := s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	if authClient != nil {
		claims := map[string]interface{}{
			"userID": persistedUser.GetID(),
			"locale": persistedUser.GetLocale(),
		}
		if err := authClient.SetCustomUserClaims(ctx, firebaseUID, claims); err != nil {
			return nil, err
		}
	}

	return persistedUser, nil
}

func (s *UserService) UpdateLocale(ctx context.Context, authClient *auth.Client, user *User, locale string) error {
	user.UpdateLocale(locale)

	_, err := s.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}

	if authClient != nil {
		claims := map[string]interface{}{
			"userID": user.GetID(),
			"locale": user.GetLocale(),
		}
		if err := authClient.SetCustomUserClaims(ctx, user.GetFirebaseUID(), claims); err != nil {
			return err
		}
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
