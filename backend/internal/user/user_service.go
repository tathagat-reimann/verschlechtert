package user

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type IUserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	Deactivate(ctx context.Context, userID int64) error
}

type UserService struct {
	userRepo IUserRepository
}

func (s *UserService) Upsert(ctx context.Context, authClient *auth.Client, firebaseUID, displayName, photoURL, email, locale string) (*User, error) {
	ctx, span := otel.Tracer("verschlechtert/backend").Start(ctx, "user.service.Upsert",
		trace.WithAttributes(
			attribute.String("user.firebase_uid", firebaseUID),
			attribute.String("user.locale", locale),
		),
	)
	defer span.End()

	user, err := NewUser(firebaseUID, displayName, photoURL, email, locale)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	persistedUser, err := s.userRepo.Save(ctx, user)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if authClient != nil {
		claims := map[string]interface{}{
			"userID": persistedUser.GetID(),
			"locale": persistedUser.GetLocale(),
		}
		if err := authClient.SetCustomUserClaims(ctx, firebaseUID, claims); err != nil {
			span.RecordError(err)
		}
	}

	return persistedUser, nil
}

func (s *UserService) UpdateLocale(ctx context.Context, authClient *auth.Client, user *User, locale string) error {
	ctx, span := otel.Tracer("verschlechtert/backend").Start(ctx, "user.service.UpdateLocale",
		trace.WithAttributes(
			attribute.Int64("user.id", user.GetID()),
			attribute.String("user.locale", locale),
		),
	)
	defer span.End()

	user.UpdateLocale(locale)

	_, err := s.userRepo.Save(ctx, user)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if authClient != nil {
		claims := map[string]interface{}{
			"userID": user.GetID(),
			"locale": user.GetLocale(),
		}
		if err := authClient.SetCustomUserClaims(ctx, user.GetFirebaseUID(), claims); err != nil {
			span.RecordError(err)
		}
	}

	return nil
}

func (s *UserService) Deactivate(ctx context.Context, user *User) error {
	ctx, span := otel.Tracer("verschlechtert/backend").Start(ctx, "user.service.Deactivate",
		trace.WithAttributes(
			attribute.Int64("user.id", user.GetID()),
		),
	)
	defer span.End()

	user.Deactivate()

	err := s.userRepo.Deactivate(ctx, user.GetID())
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func NewUserService(userRepo IUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}
