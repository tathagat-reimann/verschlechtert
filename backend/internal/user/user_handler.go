package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type IUserService interface {
	Upsert(ctx context.Context, firebaseUID, displayName, photoURL, email, locale string) (*User, error)
	UpdateLocale(ctx context.Context, user *User, locale string) error
}

type IAuthMiddleware interface {
	UserFromContext(ctx context.Context) (*auth.Token, bool)
}

type UserHandler struct {
	userService IUserService
	authmw      IAuthMiddleware
}

func NewUserHandler(userService IUserService, authmw IAuthMiddleware) *UserHandler {
	return &UserHandler{userService: userService, authmw: authmw}
}

type userResponse struct {
	ID          int64  `json:"id"`
	FirebaseUID string `json:"firebaseUid"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoUrl"`
	Email       string `json:"email"`
	Locale      string `json:"locale"`
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "could not load user", http.StatusInternalServerError)
		return
	}
	writeUserJSON(w, user)
}

func (h *UserHandler) UpdateLocale(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		http.Error(w, "could not load user", http.StatusInternalServerError)
		return
	}

	var request struct {
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !localePattern.MatchString(request.Locale) {
		http.Error(w, "invalid locale", http.StatusBadRequest)
		return
	}
	if err := h.userService.UpdateLocale(r.Context(), user, request.Locale); err != nil {
		http.Error(w, "could not update user locale", http.StatusInternalServerError)
		return
	}
	writeUserJSON(w, user)
}

func (h *UserHandler) currentUser(r *http.Request) (*User, error) {
	firebaseUser, ok := h.authmw.UserFromContext(r.Context())
	if !ok {
		return nil, errUnauthorized
	}
	user, err := h.userService.Upsert(
		r.Context(),
		firebaseUser.UID,
		claimString(firebaseUser.Claims, "name"),
		claimString(firebaseUser.Claims, "picture"),
		claimString(firebaseUser.Claims, "email"),
		requestLocale(r),
	)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, errInactiveUser
	}
	return user, nil
}

func writeUserJSON(w http.ResponseWriter, user *User) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userResponse{
		ID:          user.GetID(),
		FirebaseUID: user.GetFirebaseUID(),
		DisplayName: user.GetDisplayName(),
		PhotoURL:    user.GetPhotoURL(),
		Email:       user.GetEmail(),
		Locale:      user.GetLocale(),
	})
}

var localePattern = regexp.MustCompile(`^(de|en)$`)

func requestLocale(r *http.Request) string {
	locale := r.URL.Query().Get("locale")
	if locale == "en" {
		return "en"
	}
	return "de"
}

func claimString(claims map[string]interface{}, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

var errUnauthorized = errors.New("unauthorized")
var errInactiveUser = errors.New("user is inactive")
