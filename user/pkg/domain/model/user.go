package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrLoginExists  = errors.New("login already exists")
	ErrEmailExists  = errors.New("email already exists")
)

type User struct {
	ID        uuid.UUID
	Login     string
	Email     string
	Tg        *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(user *User) error
	Find(id uuid.UUID) (*User, error)
	FindByLogin(login string) (*User, error)
	FindByEmail(email string) (*User, error)
	Delete(id uuid.UUID) error
}
