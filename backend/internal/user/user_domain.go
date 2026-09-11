package user

import "errors"

type User struct {
	id          int64
	firebaseUID string
	displayName string
	photoURL    string
	email       string
	locale      string
	active      bool // default true, set to false if user is banned or deleted
}

func (u *User) GetID() int64 { return u.id }

func (u *User) GetFirebaseUID() string { return u.firebaseUID }

func (u *User) GetDisplayName() string { return u.displayName }

func (u *User) GetPhotoURL() string { return u.photoURL }

func (u *User) GetEmail() string { return u.email }

func (u *User) GetLocale() string { return u.locale }

func (u *User) IsActive() bool { return u.active }

func normalizeLocale(locale string) string {
	switch locale {
	case "de", "en":
		return locale
	default:
		return "de"
	}
}

func (u *User) UpdateLocale(locale string) { u.locale = normalizeLocale(locale) }

func (u *User) Deactivate() { u.active = false }

func NewUser(firebaseUID, displayName, photoURL, email, locale string) (*User, error) {
	if firebaseUID == "" {
		return nil, errors.New("firebaseUID cannot be empty")
	}

	if displayName == "" {
		return nil, errors.New("displayName cannot be empty")
	}

	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	locale = normalizeLocale(locale)

	return &User{
		firebaseUID: firebaseUID,
		displayName: displayName,
		photoURL:    photoURL,
		email:       email,
		locale:      locale,
		active:      true,
	}, nil
}
