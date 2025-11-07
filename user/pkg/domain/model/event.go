package model

import "github.com/google/uuid"

type UserCreated struct {
	UserID uuid.UUID
	Login  string
	Email  string
}

func (e UserCreated) Type() string {
	return "UserCreated"
}

type UserUpdated struct {
	UserID uuid.UUID
}

func (e UserUpdated) Type() string {
	return "UserUpdated"
}

type UserDeleted struct {
	UserID uuid.UUID
}

func (e UserDeleted) Type() string {
	return "UserDeleted"
}
