package service

import (
	"time"

	"github.com/google/uuid"

	"user/pkg/domain/model"
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type User interface {
	CreateUser(login, email string, tg *string) (uuid.UUID, error)
	UpdateUser(userID uuid.UUID, login, email string, tg *string) error
	DeleteUser(userID uuid.UUID) error
}

func NewUserService(repo model.UserRepository, dispatcher EventDispatcher) User {
	return &userService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type userService struct {
	repo       model.UserRepository
	dispatcher EventDispatcher
}

func (s *userService) CreateUser(login, email string, tg *string) (uuid.UUID, error) {
	if _, err := s.repo.FindByLogin(login); err == nil {
		return uuid.Nil, model.ErrLoginExists
	}
	if _, err := s.repo.FindByEmail(email); err == nil {
		return uuid.Nil, model.ErrEmailExists
	}

	userID, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	user := &model.User{
		ID:        userID,
		Login:     login,
		Email:     email,
		Tg:        tg,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := s.repo.Store(user); err != nil {
		return uuid.Nil, err
	}

	return userID, s.dispatcher.Dispatch(model.UserCreated{
		UserID: userID,
		Login:  login,
		Email:  email,
	})
}

func (s *userService) UpdateUser(userID uuid.UUID, login, email string, tg *string) error {
	user, err := s.repo.Find(userID)
	if err != nil {
		return err
	}

	if user.Login != login {
		if _, err := s.repo.FindByLogin(login); err == nil {
			return model.ErrLoginExists
		}
		user.Login = login
	}

	if user.Email != email {
		if _, err := s.repo.FindByEmail(email); err == nil {
			return model.ErrEmailExists
		}
		user.Email = email
	}

	user.Tg = tg
	user.UpdatedAt = time.Now()

	if err := s.repo.Store(user); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.UserUpdated{
		UserID: userID,
	})
}

func (s *userService) DeleteUser(userID uuid.UUID) error {
	if err := s.repo.Delete(userID); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.UserDeleted{
		UserID: userID,
	})
}
