package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrWalletNotFound  = errors.New("wallet not found")
)

type PaymentStatus int

const (
	Pending PaymentStatus = iota
	Completed
	Failed
)

type Payment struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	UserID        uuid.UUID
	Amount        float64
	Status        PaymentStatus
	FailureReason *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PaymentRepository interface {
	NextID() (uuid.UUID, error)
	StorePayment(payment *Payment) error
	FindPayment(id uuid.UUID) (*Payment, error)
	StoreWallet(wallet *Wallet) error
	FindWalletByUserID(userID uuid.UUID) (*Wallet, error)
}
