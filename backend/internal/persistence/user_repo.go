package persistence

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"verschlechtert/backend/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repo *UserRepository) Save(ctx context.Context, user *domain.User) error {
	_, err := repo.pool.Exec(ctx, `
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
		active = COALESCE(EXCLUDED.active, appuser.active),
		updated_at   = now(),
		last_seen_at = now()
	`, user.FirebaseUID, user.DisplayName, user.PhotoURL, user.Email, user.Locale)

	// TODO - return the updated user object if needed

	if err != nil {
		return err
	}
	return nil
}
