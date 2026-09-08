package domain

import "errors"

type User struct {
	ID          int64  `json:"id"`
	FirebaseUID string `json:"firebaseUid"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoUrl"`
	Email       string `json:"email"`
	Locale      string `json:"locale"`
	Active      bool   `json:"active"` // default true, set to false if user is banned or deleted
}

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

	if locale == "" {
		locale = "de" // default to German if locale is not provided
	}

	return &User{
		FirebaseUID: firebaseUID,
		DisplayName: displayName,
		PhotoURL:    photoURL,
		Email:       email,
		Locale:      locale,
		// Active:      true, - we comment it out as on the DB level it defaults to true
	}, nil
}
