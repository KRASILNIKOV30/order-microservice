package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"notification/pkg/domain/model"
)

var (
	ErrNotificationAlreadySent = errors.New("notification has already been sent")
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Notification interface {
	RegisterRecipient(userID uuid.UUID, email, tg *string) error
	ScheduleNotification(userID uuid.UUID, channel model.NotificationChannel, message string) (uuid.UUID, error)
	SendNotification(notificationID uuid.UUID) error
	MarkAsFailed(notificationID uuid.UUID, reason string) error
}

func NewNotificationService(repo model.NotificationRepository, dispatcher EventDispatcher) Notification {
	return &notificationService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type notificationService struct {
	repo       model.NotificationRepository
	dispatcher EventDispatcher
}

func (s *notificationService) RegisterRecipient(userID uuid.UUID, email, tg *string) error {
	recipient := &model.Recipient{
		UserID:    userID,
		Email:     email,
		Tg:        tg,
		UpdatedAt: time.Now(),
	}
	return s.repo.StoreRecipient(recipient)
}

func (s *notificationService) ScheduleNotification(userID uuid.UUID, channel model.NotificationChannel, message string) (uuid.UUID, error) {
	recipient, err := s.repo.FindRecipientByUserID(userID)
	if err != nil {
		return uuid.Nil, err
	}

	if (channel == model.Email && recipient.Email == nil) || (channel == model.Telegram && recipient.Tg == nil) {
		return uuid.Nil, model.ErrUnsupportedChannel
	}

	notificationID, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	notification := &model.Notification{
		ID:          notificationID,
		RecipientID: userID,
		Channel:     channel,
		Message:     message,
		Status:      model.Pending,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	if err := s.repo.StoreNotification(notification); err != nil {
		return uuid.Nil, err
	}

	return notificationID, s.dispatcher.Dispatch(model.NotificationCreated{
		NotificationID: notificationID,
		RecipientID:    userID,
		Channel:        channel,
	})
}

func (s *notificationService) SendNotification(notificationID uuid.UUID) error {
	notification, err := s.repo.FindNotification(notificationID)
	if err != nil {
		return err
	}
	if notification.Status != model.Pending {
		return ErrNotificationAlreadySent
	}

	// TODO: someExternalGateway.Send(notification.Channel, ...)

	notification.Status = model.Completed
	notification.SentAt = toPtr(time.Now())
	notification.UpdatedAt = *notification.SentAt

	if err := s.repo.StoreNotification(notification); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.NotificationSent{
		NotificationID: notificationID,
		RecipientID:    notification.RecipientID,
	})
}

func (s *notificationService) MarkAsFailed(notificationID uuid.UUID, reason string) error {
	notification, err := s.repo.FindNotification(notificationID)
	if err != nil {
		return err
	}
	if notification.Status != model.Pending {
		return ErrNotificationAlreadySent
	}

	notification.Status = model.Failed
	notification.FailureReason = &reason
	notification.UpdatedAt = time.Now()

	if err := s.repo.StoreNotification(notification); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.NotificationFailed{
		NotificationID: notificationID,
		RecipientID:    notification.RecipientID,
		Reason:         reason,
	})
}

func toPtr[V any](v V) *V {
	return &v
}
