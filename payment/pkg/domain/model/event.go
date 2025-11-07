package model

import "github.com/google/uuid"

type PaymentInitiated struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	UserID    uuid.UUID
	Amount    float64
}

func (e PaymentInitiated) Type() string {
	return "PaymentInitiated"
}

type PaymentCompleted struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	UserID    uuid.UUID
}

func (e PaymentCompleted) Type() string {
	return "PaymentCompleted"
}

type PaymentFailed struct {
	PaymentID     uuid.UUID
	OrderID       uuid.UUID
	UserID        uuid.UUID
	FailureReason string
}

func (e PaymentFailed) Type() string {
	return "PaymentFailed"
}
