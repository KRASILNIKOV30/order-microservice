package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"payment/pkg/domain/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) NextID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

func (r *PaymentRepository) StorePayment(payment *model.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, amount, status, failure_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			order_id = VALUES(order_id),
			user_id = VALUES(user_id),
			amount = VALUES(amount),
			status = VALUES(status),
			failure_reason = VALUES(failure_reason),
			updated_at = VALUES(updated_at)
	`

	failureReason := (*string)(nil)
	if payment.FailureReason != nil {
		failureReason = payment.FailureReason
	}

	_, err := r.db.Exec(query,
		payment.ID.String(),
		payment.OrderID.String(),
		payment.UserID.String(),
		payment.Amount,
		int(payment.Status),
		failureReason,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}

	return nil
}

func (r *PaymentRepository) FindPayment(id uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, failure_reason, created_at, updated_at
		FROM payments
		WHERE id = ?
	`

	var payment PaymentRow

	err := r.db.Get(&payment, query, id.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return r.rowToPayment(&payment), nil
}

func (r *PaymentRepository) StoreWallet(wallet *model.Wallet) error {
	query := `
		INSERT INTO wallets (id, user_id, balance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			balance = VALUES(balance),
			updated_at = VALUES(updated_at)
	`

	_, err := r.db.Exec(query,
		wallet.ID.String(),
		wallet.UserID.String(),
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store wallet: %w", err)
	}

	return nil
}

func (r *PaymentRepository) FindWalletByUserID(userID uuid.UUID) (*model.Wallet, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM wallets
		WHERE user_id = ?
	`

	var wallet WalletRow
	err := r.db.Get(&wallet, query, userID.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrWalletNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	return r.rowToWallet(&wallet), nil
}

// GetPaymentsByOrderID получает платежи по ID заказа
func (r *PaymentRepository) GetPaymentsByOrderID(orderID uuid.UUID) ([]*model.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, failure_reason, created_at, updated_at
		FROM payments
		WHERE order_id = ?
		ORDER BY created_at DESC
	`

	var payments []PaymentRow
	err := r.db.Select(&payments, query, orderID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by order: %w", err)
	}

	result := make([]*model.Payment, len(payments))
	for i, paymentRow := range payments {
		result[i] = r.rowToPayment(&paymentRow)
	}

	return result, nil
}

// UpdatePaymentStatus обновляет статус платежа
func (r *PaymentRepository) UpdatePaymentStatus(id uuid.UUID, status model.PaymentStatus, failureReason *string) error {
	query := `
		UPDATE payments 
		SET status = ?, failure_reason = ?, updated_at = NOW()
		WHERE id = ?
	`

	failureReasonValue := (*string)(nil)
	if failureReason != nil {
		failureReasonValue = failureReason
	}

	_, err := r.db.Exec(query, int(status), failureReasonValue, id.String())
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// UpdateWalletBalance обновляет баланс кошелька
func (r *PaymentRepository) UpdateWalletBalance(userID uuid.UUID, newBalance float64) error {
	query := `
		UPDATE wallets 
		SET balance = ?, updated_at = NOW()
		WHERE user_id = ?
	`

	_, err := r.db.Exec(query, newBalance, userID.String())
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

type PaymentRow struct {
	ID            string         `db:"id"`
	OrderID       string         `db:"order_id"`
	UserID        string         `db:"user_id"`
	Amount        float64        `db:"amount"`
	Status        int            `db:"status"`
	FailureReason sql.NullString `db:"failure_reason"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

type WalletRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *PaymentRepository) rowToPayment(row *PaymentRow) *model.Payment {
	paymentID, _ := uuid.Parse(row.ID)
	orderID, _ := uuid.Parse(row.OrderID)
	userID, _ := uuid.Parse(row.UserID)

	failureReason := (*string)(nil)
	if row.FailureReason.Valid {
		failureReasonStr := row.FailureReason.String
		failureReason = &failureReasonStr
	}

	return &model.Payment{
		ID:            paymentID,
		OrderID:       orderID,
		UserID:        userID,
		Amount:        row.Amount,
		Status:        model.PaymentStatus(row.Status),
		FailureReason: failureReason,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func (r *PaymentRepository) rowToWallet(row *WalletRow) *model.Wallet {
	walletID, _ := uuid.Parse(row.ID)
	userID, _ := uuid.Parse(row.UserID)

	return &model.Wallet{
		ID:        walletID,
		UserID:    userID,
		Balance:   row.Balance,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
