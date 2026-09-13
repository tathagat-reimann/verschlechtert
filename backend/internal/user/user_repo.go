package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repo *UserRepository) Save(ctx context.Context, user *User) (*User, error) {
	ctx, span := otel.Tracer("verschlechtert/backend").Start(ctx, "user.repo.Save",
		trace.WithAttributes(
			attribute.String("user.firebase_uid", user.GetFirebaseUID()),
			attribute.String("user.locale", user.GetLocale()),
		),
	)
	defer span.End()

	persistedUser := &User{}
	err := repo.pool.QueryRow(ctx, `
	INSERT INTO appuser (
		firebase_uid,
		display_name,
		photo_url,
		email,
		locale
	)
	VALUES (
		$1,
		COALESCE(NULLIF($2, ''), 'Anonymous user'),
		NULLIF($3, ''),
		NULLIF($4, ''),
		COALESCE(NULLIF($5, ''), 'de')
	)
	ON CONFLICT (firebase_uid) DO UPDATE SET
		display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), appuser.display_name),
		photo_url    = COALESCE(NULLIF(EXCLUDED.photo_url, ''), appuser.photo_url),
		email        = COALESCE(NULLIF(EXCLUDED.email, ''), appuser.email),
		locale       = COALESCE(NULLIF(EXCLUDED.locale, ''), appuser.locale),
		updated_at   = now(),
		last_seen_at = now()
	RETURNING id, firebase_uid, display_name, COALESCE(photo_url, ''), COALESCE(email, ''), locale, active
	`, user.GetFirebaseUID(), user.GetDisplayName(), user.GetPhotoURL(), user.GetEmail(), user.GetLocale()).Scan(
		&persistedUser.id,
		&persistedUser.firebaseUID,
		&persistedUser.displayName,
		&persistedUser.photoURL,
		&persistedUser.email,
		&persistedUser.locale,
		&persistedUser.active,
	)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	return persistedUser, nil
}

func (repo *UserRepository) Deactivate(ctx context.Context, userID int64) error {
	ctx, span := otel.Tracer("verschlechtert/backend").Start(ctx, "user.repo.Deactivate",
		trace.WithAttributes(
			attribute.Int64("user.id", userID),
		),
	)
	defer span.End()

	_, err := repo.pool.Exec(ctx, `
		UPDATE appuser
		SET active = false, updated_at = now()
		WHERE id = $1
	`, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
