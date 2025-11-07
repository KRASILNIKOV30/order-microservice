package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrRecipientNotFound    = errors.New("recipient not found")
	ErrUnsupportedChannel   = errors.New("recipient does not support this channel")
)

type NotificationStatus int

const (
	Pending NotificationStatus = iota
	Completed
	Failed
)

type NotificationChannel int

const (
	Email NotificationChannel = iota
	Telegram
)

type Notification struct {
	ID            uuid.UUID
	RecipientID   uuid.UUID // Ссылка на ID пользователя
	Channel       NotificationChannel
	Message       string
	Status        NotificationStatus
	FailureReason *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SentAt        *time.Time
}

type Recipient struct {
	UserID    uuid.UUID
	Email     *string
	Tg        *string
	UpdatedAt time.Time
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	StoreNotification(notification *Notification) error
	FindNotification(id uuid.UUID) (*Notification, error)
	StoreRecipient(recipient *Recipient) error
	FindRecipientByUserID(userID uuid.UUID) (*Recipient, error)
}
