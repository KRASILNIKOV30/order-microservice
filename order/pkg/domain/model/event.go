package model

import "github.com/google/uuid"

type OrderCreated struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
}

type OrderDeleted struct {
	OrderID uuid.UUID
}

func (e OrderCreated) Type() string {
	return "OrderCreated"
}
func (e OrderDeleted) Type() string {
	return "OrderDeleted"
}

type OrderItemChanged struct {
	OrderID      uuid.UUID
	AddedItems   []uuid.UUID
	RemovedItems []uuid.UUID
}

func (e OrderItemChanged) Type() string {
	return "OrderItemChanged"
}
