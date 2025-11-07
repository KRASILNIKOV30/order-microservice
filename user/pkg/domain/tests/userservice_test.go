package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"user/pkg/domain/model"
	"user/pkg/domain/service"
)

type testFixture struct {
	userService     service.User
	repo            *mockUserRepository
	eventDispatcher *mockEventDispatcher
}

func setup() testFixture {
	repo := &mockUserRepository{store: make(map[uuid.UUID]*model.User)}
	eventDispatcher := &mockEventDispatcher{}
	userService := service.NewUserService(repo, eventDispatcher)

	return testFixture{
		userService:     userService,
		repo:            repo,
		eventDispatcher: eventDispatcher,
	}
}

func TestUserService(t *testing.T) {
	login := "testuser"
	email := "test@example.com"
	tg := "testtg"

	t.Run("Create user", func(t *testing.T) {
		f := setup()

		userID, err := f.userService.CreateUser(login, email, toPtr(tg))

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[userID])
		require.Equal(t, login, f.repo.store[userID].Login)
		require.Equal(t, email, f.repo.store[userID].Email)
		require.NotNil(t, f.repo.store[userID].Tg)
		require.Equal(t, tg, *f.repo.store[userID].Tg)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.UserCreated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Delete user", func(t *testing.T) {
		f := setup()
		userID, _ := f.userService.CreateUser(login, email, nil)
		f.eventDispatcher.events = nil

		err := f.userService.DeleteUser(userID)

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[userID].DeletedAt)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.UserDeleted{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Update user", func(t *testing.T) {
		f := setup()
		userID, _ := f.userService.CreateUser(login, email, nil)
		f.eventDispatcher.events = nil

		newLogin := "newlogin"
		newEmail := "new@email.com"
		err := f.userService.UpdateUser(userID, newLogin, newEmail, toPtr(tg))

		require.NoError(t, err)
		require.Equal(t, newLogin, f.repo.store[userID].Login)
		require.Equal(t, newEmail, f.repo.store[userID].Email)
		require.NotNil(t, f.repo.store[userID].Tg)
		require.Equal(t, tg, *f.repo.store[userID].Tg)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.UserUpdated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Fail to create user with duplicate login", func(t *testing.T) {
		f := setup()
		_, _ = f.userService.CreateUser(login, email, nil)
		f.eventDispatcher.events = nil

		_, err := f.userService.CreateUser(login, "another@email.com", nil)

		require.ErrorIs(t, err, model.ErrLoginExists)
		require.Empty(t, f.eventDispatcher.events)
	})

	t.Run("Fail to create user with duplicate email", func(t *testing.T) {
		f := setup()
		_, _ = f.userService.CreateUser(login, email, nil)
		f.eventDispatcher.events = nil

		_, err := f.userService.CreateUser("anotherlogin", email, nil)

		require.ErrorIs(t, err, model.ErrEmailExists)
		require.Empty(t, f.eventDispatcher.events)
	})
}

var _ model.UserRepository = &mockUserRepository{}

type mockUserRepository struct {
	store map[uuid.UUID]*model.User
}

func (m *mockUserRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockUserRepository) Store(user *model.User) error {
	m.store[user.ID] = user
	return nil
}

func (m *mockUserRepository) Find(id uuid.UUID) (*model.User, error) {
	if user, ok := m.store[id]; ok && user.DeletedAt == nil {
		return user, nil
	}
	return nil, model.ErrUserNotFound
}

func (m *mockUserRepository) FindByLogin(login string) (*model.User, error) {
	for _, user := range m.store {
		if user.Login == login && user.DeletedAt == nil {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(email string) (*model.User, error) {
	for _, user := range m.store {
		if user.Email == email && user.DeletedAt == nil {
			return user, nil
		}
	}
	return nil, model.ErrUserNotFound
}

func (m *mockUserRepository) Delete(id uuid.UUID) error {
	if user, ok := m.store[id]; ok && user.DeletedAt == nil {
		user.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrUserNotFound
}

var _ service.EventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(event service.Event) error {
	m.events = append(m.events, event)
	return nil
}

func toPtr[V any](v V) *V {
	return &v
}
