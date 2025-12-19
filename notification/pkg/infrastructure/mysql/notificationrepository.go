package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"notification/pkg/domain/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

func (r *NotificationRepository) StoreNotification(notification *model.Notification) error {
	query := `
		INSERT INTO notifications (id, recipient_id, channel, message, status, failure_reason, created_at, updated_at, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			recipient_id = VALUES(recipient_id),
			channel = VALUES(channel),
			message = VALUES(message),
			status = VALUES(status),
			failure_reason = VALUES(failure_reason),
			updated_at = VALUES(updated_at),
			sent_at = VALUES(sent_at)
	`

	failureReason := (*string)(nil)
	if notification.FailureReason != nil {
		failureReason = notification.FailureReason
	}

	sentAt := (*time.Time)(nil)
	if notification.SentAt != nil {
		sentAt = notification.SentAt
	}

	_, err := r.db.Exec(query,
		notification.ID.String(),
		notification.RecipientID.String(),
		int(notification.Channel),
		notification.Message,
		int(notification.Status),
		failureReason,
		notification.CreatedAt,
		notification.UpdatedAt,
		sentAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store notification: %w", err)
	}

	return nil
}

func (r *NotificationRepository) FindNotification(id uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT id, recipient_id, channel, message, status, failure_reason, created_at, updated_at, sent_at
		FROM notifications
		WHERE id = ?
	`

	var notification NotificationRow

	err := r.db.Get(notification, query, id.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrNotificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	return r.rowToNotification(toPtr(notification)), nil
}

func (r *NotificationRepository) StoreRecipient(recipient *model.Recipient) error {
	query := `
		INSERT INTO recipients (user_id, email, tg, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			email = VALUES(email),
			tg = VALUES(tg),
			updated_at = VALUES(updated_at)
	`

	email := (*string)(nil)
	if recipient.Email != nil {
		email = recipient.Email
	}

	tg := (*string)(nil)
	if recipient.Tg != nil {
		tg = recipient.Tg
	}

	_, err := r.db.Exec(query,
		recipient.UserID.String(),
		email,
		tg,
		recipient.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store recipient: %w", err)
	}

	return nil
}

func (r *NotificationRepository) FindRecipientByUserID(userID uuid.UUID) (*model.Recipient, error) {
	query := `
		SELECT user_id, email, tg, updated_at
		FROM recipients
		WHERE user_id = ?
	`

	var recipient RecipientRow

	err := r.db.Get(&recipient, query, userID.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrRecipientNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find recipient: %w", err)
	}

	return r.rowToRecipient(&recipient), nil
}

// GetPendingNotifications получает ожидающие отправки уведомления
func (r *NotificationRepository) GetPendingNotifications() ([]*model.Notification, error) {
	query := `
		SELECT id, recipient_id, channel, message, status, failure_reason, created_at, updated_at, sent_at
		FROM notifications
		WHERE status = ?
		ORDER BY created_at ASC
	`

	var notifications []NotificationRow
	err := r.db.Select(notifications, query, int(model.Pending))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}

	result := make([]*model.Notification, len(notifications))
	for i, notificationRow := range notifications {
		result[i] = r.rowToNotification(toPtr(notificationRow))
	}

	return result, nil
}

// UpdateNotificationStatus обновляет статус уведомления
func (r *NotificationRepository) UpdateNotificationStatus(id uuid.UUID, status model.NotificationStatus, failureReason *string, sentAt *time.Time) error {
	query := `
		UPDATE notifications 
		SET status = ?, failure_reason = ?, sent_at = ?, updated_at = NOW()
		WHERE id = ?
	`

	failureReasonValue := (*string)(nil)
	if failureReason != nil {
		failureReasonValue = failureReason
	}

	sentAtValue := (*time.Time)(nil)
	if sentAt != nil {
		sentAtValue = sentAt
	}

	_, err := r.db.Exec(query, int(status), failureReasonValue, sentAtValue, id.String())
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// GetNotificationsByRecipientID получает уведомления по ID получателя
func (r *NotificationRepository) GetNotificationsByRecipientID(recipientID uuid.UUID) ([]*model.Notification, error) {
	query := `
		SELECT id, recipient_id, channel, message, status, failure_reason, created_at, updated_at, sent_at
		FROM notifications
		WHERE recipient_id = ?
		ORDER BY created_at DESC
	`

	var notifications []NotificationRow
	err := r.db.Select(notifications, query, recipientID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by recipient: %w", err)
	}

	result := make([]*model.Notification, len(notifications))
	for i, notificationRow := range notifications {
		result[i] = r.rowToNotification(toPtr(notificationRow))
	}

	return result, nil
}

type NotificationRow struct {
	ID            string         `db:"id"`
	RecipientID   string         `db:"recipient_id"`
	Channel       int            `db:"channel"`
	Message       string         `db:"message"`
	Status        int            `db:"status"`
	FailureReason sql.NullString `db:"failure_reason"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	SentAt        sql.NullTime   `db:"sent_at"`
}

type RecipientRow struct {
	UserID    string         `db:"user_id"`
	Email     sql.NullString `db:"email"`
	Tg        sql.NullString `db:"tg"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (r *NotificationRepository) rowToNotification(row *NotificationRow) *model.Notification {
	notificationID, _ := uuid.Parse(row.ID)
	recipientID, _ := uuid.Parse(row.RecipientID)

	failureReason := (*string)(nil)
	if row.FailureReason.Valid {
		failureReasonStr := row.FailureReason.String
		failureReason = &failureReasonStr
	}

	sentAt := (*time.Time)(nil)
	if row.SentAt.Valid {
		sentAt = &row.SentAt.Time
	}

	return &model.Notification{
		ID:            notificationID,
		RecipientID:   recipientID,
		Channel:       model.NotificationChannel(row.Channel),
		Message:       row.Message,
		Status:        model.NotificationStatus(row.Status),
		FailureReason: failureReason,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		SentAt:        sentAt,
	}
}

func (r *NotificationRepository) rowToRecipient(row *RecipientRow) *model.Recipient {
	userID, _ := uuid.Parse(row.UserID)

	email := (*string)(nil)
	if row.Email.Valid {
		emailStr := row.Email.String
		email = &emailStr
	}

	tg := (*string)(nil)
	if row.Tg.Valid {
		tgStr := row.Tg.String
		tg = &tgStr
	}

	return &model.Recipient{
		UserID:    userID,
		Email:     email,
		Tg:        tg,
		UpdatedAt: row.UpdatedAt,
	}
}

func toPtr[V any](v V) *V {
	return &v
}
