package service

import (
	"context"
	"errors"
	"testing"
	"verschlechtert/backend/internal/domain"

	"github.com/stretchr/testify/assert"
)

type mockUserRepository struct {
	upsertError error
}

func (m *mockUserRepository) Save(ctx context.Context, user *domain.User) error {
	if m.upsertError != nil {
		return m.upsertError
	}
	return nil
}

func TestUserService_Upsert_Success(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	service := NewUserService(mockRepo)

	// when
	user, err := service.Upsert(t.Context(), "firebaseUID", "displayName", "photoURL", "email", "locale")

	//then
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserService_Upsert_DatabaseError(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	mockRepo.upsertError = errors.New("database error")
	service := NewUserService(mockRepo)

	// when
	user, err := service.Upsert(t.Context(), "firebaseUID", "displayName", "photoURL", "email", "locale")

	//then
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_Upsert_MissingMandatoryFieldError(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	service := NewUserService(mockRepo)

	// when
	user, err := service.Upsert(t.Context(), "", "displayName", "photoURL", "email", "locale")

	//then
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserService_UpdateLocale_Success(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	service := NewUserService(mockRepo)

	// when
	user, _ := domain.NewUser("f", "d", "p", "e", "")
	err := service.UpdateLocale(t.Context(), user, "en")

	//then
	assert.NoError(t, err)
}

func TestUserService_UpdateLocale_Error(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	mockRepo.upsertError = errors.New("error")
	service := NewUserService(mockRepo)

	// when
	user, _ := domain.NewUser("f", "d", "p", "e", "")
	err := service.UpdateLocale(t.Context(), user, "en")

	//then
	assert.Error(t, err)
}

func TestUserService_Deactivate_Success(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	service := NewUserService(mockRepo)

	// when
	user, _ := domain.NewUser("f", "d", "p", "e", "")
	err := service.Deactivate(t.Context(), user)

	//then
	assert.NoError(t, err)
}

func TestUserService_Deactivate_Error(t *testing.T) {
	// given
	mockRepo := &mockUserRepository{}
	mockRepo.upsertError = errors.New("error")
	service := NewUserService(mockRepo)

	// when
	user, _ := domain.NewUser("f", "d", "p", "e", "")
	err := service.Deactivate(t.Context(), user)

	//then
	assert.Error(t, err)
}
