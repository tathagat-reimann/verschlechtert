package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Success(t *testing.T) {
	user, err := NewUser("firebase-uid", "Display Name", "photo-url", "user@example.com", "en")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "firebase-uid", user.GetFirebaseUID())
	assert.Equal(t, "Display Name", user.GetDisplayName())
	assert.Equal(t, "photo-url", user.GetPhotoURL())
	assert.Equal(t, "user@example.com", user.GetEmail())
	assert.Equal(t, "en", user.GetLocale())
	assert.True(t, user.IsActive())
}

func TestNewUser_DefaultsLocaleToGerman(t *testing.T) {
	user, err := NewUser("firebase-uid", "Display Name", "photo-url", "user@example.com", "")

	require.NoError(t, err)
	assert.Equal(t, "de", user.GetLocale())
}

func TestNewUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		firebaseUID  string
		displayName  string
		email        string
		expectedText string
	}{
		{
			name:         "missing firebase UID",
			displayName:  "Display Name",
			email:        "user@example.com",
			expectedText: "firebaseUID cannot be empty",
		},
		{
			name:         "missing display name",
			firebaseUID:  "firebase-uid",
			email:        "user@example.com",
			expectedText: "displayName cannot be empty",
		},
		{
			name:         "missing email",
			firebaseUID:  "firebase-uid",
			displayName:  "Display Name",
			expectedText: "email cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := NewUser(test.firebaseUID, test.displayName, "photo-url", test.email, "en")

			assert.Nil(t, user)
			require.EqualError(t, err, test.expectedText)
		})
	}
}

func TestUser_UpdateLocale(t *testing.T) {
	user, err := NewUser("firebase-uid", "Display Name", "photo-url", "user@example.com", "de")
	require.NoError(t, err)

	user.UpdateLocale("en")

	assert.Equal(t, "en", user.GetLocale())
}

func TestUser_UpdateLocale_RejectsUnsupportedLocale(t *testing.T) {
	user, err := NewUser("firebase-uid", "Display Name", "photo-url", "user@example.com", "de")
	require.NoError(t, err)

	user.UpdateLocale("fr")

	assert.Equal(t, "de", user.GetLocale())
}

func TestUser_Deactivate(t *testing.T) {
	user, err := NewUser("firebase-uid", "Display Name", "photo-url", "user@example.com", "de")
	require.NoError(t, err)
	assert.True(t, user.IsActive())

	user.Deactivate()

	assert.False(t, user.IsActive())
}
