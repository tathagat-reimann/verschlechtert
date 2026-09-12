package user

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"firebase.google.com/go/v4/auth"
)

type userServiceMock struct {
	user *User
}

func (u *userServiceMock) Upsert(ctx context.Context, firebaseUID, displayName, photoURL, email, locale string) (*User, error) {
	return u.user, nil
}

func (u *userServiceMock) UpdateLocale(ctx context.Context, user *User, locale string) error {
	return nil
}

var authorizedUser = &auth.Token{
	UID: "test-uid",
	Claims: map[string]interface{}{
		"name":    "Test User",
		"picture": "http://example.com/photo.jpg",
		"email":   "test@example.com",
	},
}
var userOk = true

var userFromContextMock = func(ctx context.Context) (*auth.Token, bool) {
	return authorizedUser, userOk
}

func TestGetMe(t *testing.T) {
	// given
	userServiceMock := &userServiceMock{
		user: &User{
			active: true,
		},
	}
	// authmwMock := &authmwMock{
	// 	firebaseUser: &auth.Token{
	// 		UID: "test-uid",
	// 		Claims: map[string]interface{}{
	// 			"name":    "Test User",
	// 			"picture": "http://example.com/photo.jpg",
	// 			"email":   "test@example.com",
	// 		},
	// 	},
	// 	ok: true,
	// }
	handler := NewUserHandler(userServiceMock, userFromContextMock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/me", nil)

	// set test context in r with key "firebaseUser" and value firebaseUser
	// firebaseUser :=
	// 	&auth.Token{
	// 		UID: "test-uid",
	// 		Claims: map[string]interface{}{
	// 			"name":    "Test User",
	// 			"picture": "http://example.com/photo.jpg",
	// 			"email":   "test@example.com",
	// 		},
	// 	}
	// ctx := context.WithValue(r.Context(), "firebaseUser", firebaseUser)
	// r = r.WithContext(ctx)

	// when
	handler.GetMe(w, r)

	// then
	if w.Code != 200 {
		t.Errorf("expected status code 200, got %d", w.Code)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	// given
	// authmwMock := &authmwMock{
	// 	ok: false,
	// }
	handler := NewUserHandler(nil, userFromContextMock)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/me", nil)

	userOk = false

	// when
	handler.GetMe(w, r)

	// then
	if w.Code != 500 {
		t.Errorf("expected status code 500, got %d", w.Code)
	}
}

func TestUpdateLocale(t *testing.T) {
	// given
	userServiceMock := &userServiceMock{
		user: &User{
			active: true,
		},
	}
	// authmwMock := &authmwMock{
	// 	firebaseUser: &auth.Token{
	// 		UID: "test-uid",
	// 		Claims: map[string]interface{}{
	// 			"name":    "Test User",
	// 			"picture": "http://example.com/photo.jpg",
	// 			"email":   "test@example.com",
	// 		},
	// 	},
	// 	ok: true,
	// }
	handler := NewUserHandler(userServiceMock, userFromContextMock)

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"locale": "en"}`)
	r := httptest.NewRequest("PATCH", "/me/locale", body)

	userOk = true

	// set test context in r with key "firebaseUser" and value firebaseUser
	// firebaseUser :=
	// 	&auth.Token{
	// 		UID: "test-uid",
	// 		Claims: map[string]interface{}{
	// 			"name":    "Test User",
	// 			"picture": "http://example.com/photo.jpg",
	// 			"email":   "test@example.com",
	// 		},
	// 	}
	// ctx := context.WithValue(r.Context(), "firebaseUser", firebaseUser)
	// r = r.WithContext(ctx)

	// when
	handler.UpdateLocale(w, r)

	// then
	if w.Code != 200 {
		t.Errorf("expected status code 200, got %d", w.Code)
	}
}
