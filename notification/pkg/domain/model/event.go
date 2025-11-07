package model

import "github.com/google/uuid"

type NotificationCreated struct {
	NotificationID uuid.UUID
	RecipientID    uuid.UUID
	Channel        NotificationChannel
}

func (e NotificationCreated) Type() string {
	return "NotificationCreated"
}

type NotificationSent struct {
	NotificationID uuid.UUID
	RecipientID    uuid.UUID
}

func (e NotificationSent) Type() string {
	return "NotificationSent"
}

type NotificationFailed struct {
	NotificationID uuid.UUID
	RecipientID    uuid.UUID
	Reason         string
}

func (e NotificationFailed) Type() string {
	return "NotificationFailed"
}
