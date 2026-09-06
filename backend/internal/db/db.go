package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_initial_schema.sql
var initialSchema []byte

type User struct {
	ID          int64
	FirebaseUID string
	DisplayName string
	PhotoURL    string
	Email       string
	Locale      string
}

// NewPool creates a Postgres connection pool from a DATABASE_URL connection string.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}

// ApplyMigrations creates the database schema required by the application.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("creating schema migrations table: %w", err)
	}

	var applied bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM schema_migrations WHERE version = 1
		)
	`).Scan(&applied); err != nil {
		return fmt.Errorf("checking initial schema migration: %w", err)
	}
	if applied {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting initial schema migration: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(initialSchema)); err != nil {
		return fmt.Errorf("applying initial schema migration: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES (1)`); err != nil {
		return fmt.Errorf("recording initial schema migration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing initial schema migration: %w", err)
	}

	return nil
}

// UpsertUser creates or refreshes the local profile for a verified Firebase user.
func UpsertUser(ctx context.Context, pool *pgxpool.Pool, firebaseUID, displayName, photoURL, email string) (User, error) {
	var user User
	err := pool.QueryRow(ctx, `
		INSERT INTO users (firebase_uid, display_name, photo_url, email)
		VALUES ($1, COALESCE(NULLIF($2, ''), 'Anonymous user'), NULLIF($3, ''), NULLIF($4, ''))
		ON CONFLICT (firebase_uid) DO UPDATE SET
			display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
			photo_url = COALESCE(EXCLUDED.photo_url, users.photo_url),
			email = COALESCE(EXCLUDED.email, users.email),
			updated_at = now(),
			last_seen_at = now()
		RETURNING id, firebase_uid, display_name, COALESCE(photo_url, ''), COALESCE(email, ''), locale
	`, firebaseUID, displayName, photoURL, email).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.DisplayName,
		&user.PhotoURL,
		&user.Email,
		&user.Locale,
	)
	if err != nil {
		return User{}, fmt.Errorf("upserting user: %w", err)
	}
	return user, nil
}

// UpdateUserLocale changes the application language preference for a user.
func UpdateUserLocale(ctx context.Context, pool *pgxpool.Pool, firebaseUID, locale string) (User, error) {
	var user User
	err := pool.QueryRow(ctx, `
		UPDATE users
		SET locale = $2, updated_at = now()
		WHERE firebase_uid = $1
		RETURNING id, firebase_uid, display_name, COALESCE(photo_url, ''), COALESCE(email, ''), locale
	`, firebaseUID, locale).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.DisplayName,
		&user.PhotoURL,
		&user.Email,
		&user.Locale,
	)
	if err != nil {
		return User{}, fmt.Errorf("updating user locale: %w", err)
	}
	return user, nil
}
