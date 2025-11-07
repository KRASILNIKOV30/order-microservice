package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"notification/pkg/domain/model"
	"notification/pkg/domain/service"
)

type testFixture struct {
	notificationService service.Notification
	repo                *mockNotificationRepository
	eventDispatcher     *mockEventDispatcher
}

func setup() testFixture {
	repo := &mockNotificationRepository{
		notificationStore: make(map[uuid.UUID]*model.Notification),
		recipientStore:    make(map[uuid.UUID]*model.Recipient),
	}
	eventDispatcher := &mockEventDispatcher{}
	notificationService := service.NewNotificationService(repo, eventDispatcher)

	return testFixture{
		notificationService: notificationService,
		repo:                repo,
		eventDispatcher:     eventDispatcher,
	}
}

func TestNotificationService(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	email := "test@example.com"
	tg := "test_tg"
	message := "Your order has been shipped!"

	t.Run("Register recipient", func(t *testing.T) {
		f := setup()
		err := f.notificationService.RegisterRecipient(userID, toPtr(email), toPtr(tg))

		require.NoError(t, err)
		recipient, ok := f.repo.recipientStore[userID]
		require.True(t, ok)
		require.Equal(t, email, *recipient.Email)
		require.Equal(t, tg, *recipient.Tg)
	})

	t.Run("Schedule notification successfully", func(t *testing.T) {
		f := setup()
		_ = f.notificationService.RegisterRecipient(userID, toPtr(email), nil)

		notificationID, err := f.notificationService.ScheduleNotification(userID, model.Email, message)

		require.NoError(t, err)
		notification, ok := f.repo.notificationStore[notificationID]
		require.True(t, ok)
		require.Equal(t, model.Pending, notification.Status)
		require.Equal(t, message, notification.Message)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.NotificationCreated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Send notification successfully", func(t *testing.T) {
		f := setup()
		_ = f.notificationService.RegisterRecipient(userID, toPtr(email), nil)
		notificationID, _ := f.notificationService.ScheduleNotification(userID, model.Email, message)
		f.eventDispatcher.events = nil

		err := f.notificationService.SendNotification(notificationID)

		require.NoError(t, err)
		notification := f.repo.notificationStore[notificationID]
		require.Equal(t, model.Completed, notification.Status)
		require.NotNil(t, notification.SentAt)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.NotificationSent{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Fail to schedule for unsupported channel", func(t *testing.T) {
		f := setup()
		_ = f.notificationService.RegisterRecipient(userID, toPtr(email), nil) // No Telegram registered

		_, err := f.notificationService.ScheduleNotification(userID, model.Telegram, message)

		require.ErrorIs(t, err, model.ErrUnsupportedChannel)
	})

	t.Run("Fail to schedule for unknown recipient", func(t *testing.T) {
		f := setup()

		_, err := f.notificationService.ScheduleNotification(userID, model.Email, message)

		require.ErrorIs(t, err, model.ErrRecipientNotFound)
	})

	t.Run("Mark notification as failed", func(t *testing.T) {
		f := setup()
		_ = f.notificationService.RegisterRecipient(userID, toPtr(email), nil)
		notificationID, _ := f.notificationService.ScheduleNotification(userID, model.Email, message)
		f.eventDispatcher.events = nil
		reason := "External gateway timeout"

		err := f.notificationService.MarkAsFailed(notificationID, reason)

		require.NoError(t, err)
		notification := f.repo.notificationStore[notificationID]
		require.Equal(t, model.Failed, notification.Status)
		require.Equal(t, reason, *notification.FailureReason)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.NotificationFailed{}.Type(), f.eventDispatcher.events[0].Type())
	})
}

var _ model.NotificationRepository = &mockNotificationRepository{}

type mockNotificationRepository struct {
	notificationStore map[uuid.UUID]*model.Notification
	recipientStore    map[uuid.UUID]*model.Recipient
}

func (m *mockNotificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockNotificationRepository) StoreNotification(notification *model.Notification) error {
	m.notificationStore[notification.ID] = notification
	return nil
}

func (m *mockNotificationRepository) FindNotification(id uuid.UUID) (*model.Notification, error) {
	if n, ok := m.notificationStore[id]; ok {
		return n, nil
	}
	return nil, model.ErrNotificationNotFound
}

func (m *mockNotificationRepository) StoreRecipient(recipient *model.Recipient) error {
	m.recipientStore[recipient.UserID] = recipient
	return nil
}

func (m *mockNotificationRepository) FindRecipientByUserID(userID uuid.UUID) (*model.Recipient, error) {
	if r, ok := m.recipientStore[userID]; ok {
		return r, nil
	}
	return nil, model.ErrRecipientNotFound
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
