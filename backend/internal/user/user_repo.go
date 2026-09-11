package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repo *UserRepository) Save(ctx context.Context, user *User) (*User, error) {
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
		return nil, err
	}
	return persistedUser, nil
}

func (repo *UserRepository) Deactivate(ctx context.Context, userID int64) error {
	_, err := repo.pool.Exec(ctx, `
		UPDATE appuser
		SET active = false, updated_at = now()
		WHERE id = $1
	`, userID)
	return err
}
