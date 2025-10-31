package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"order/pkg/domain/model"
)

var (
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Order interface {
	CreateOrder(customerID uuid.UUID) (uuid.UUID, error)
	DeleteOrder(orderID uuid.UUID) error
	SetStatus(orderID uuid.UUID, status model.OrderStatus) error

	AddItem(orderID uuid.UUID, productID uuid.UUID, price float64) (uuid.UUID, error)
	DeleteItem(orderID uuid.UUID, itemID uuid.UUID) error
}

func NewOrderService(repo model.OrderRepository, dispatcher EventDispatcher) Order {
	return &orderService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type orderService struct {
	repo       model.OrderRepository
	dispatcher EventDispatcher
}

func (o *orderService) CreateOrder(customerID uuid.UUID) (uuid.UUID, error) {
	orderID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = o.repo.Store(&model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Open,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, o.dispatcher.Dispatch(model.OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}

func (o *orderService) DeleteOrder(orderID uuid.UUID) error {
	err := o.repo.Delete(orderID)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderDeleted{
		OrderID: orderID,
	})
}

func (o *orderService) SetStatus(orderID uuid.UUID, status model.OrderStatus) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return err
	}

	oldStatus := order.Status
	if oldStatus == status {
		return nil
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderStatusChanged{
		OrderID:   orderID,
		OldStatus: oldStatus,
		NewStatus: status,
	})
}

func (o *orderService) AddItem(orderID, productID uuid.UUID, price float64) (uuid.UUID, error) {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return uuid.Nil, err
	}

	if order.Status != model.Open {
		return uuid.Nil, ErrInvalidOrderStatus
	}

	itemID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}
	order.Items = append(order.Items, model.Item{
		ID:        itemID,
		ProductID: productID,
		Price:     price,
	})
	order.UpdatedAt = time.Now()

	err = o.repo.Store(order)
	if err != nil {
		return uuid.Nil, err
	}

	return itemID, o.dispatcher.Dispatch(model.OrderItemChanged{
		OrderID:    orderID,
		AddedItems: []uuid.UUID{itemID},
	})
}

func (o *orderService) DeleteItem(orderID, itemID uuid.UUID) error {
	order, err := o.repo.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status != model.Open {
		return ErrInvalidOrderStatus
	}

	itemIndex, found := findItemIndex(order.Items, itemID)
	if !found {
		return model.ErrItemNotFound
	}

	order.Items = append(order.Items[:itemIndex], order.Items[itemIndex+1:]...)
	order.UpdatedAt = time.Now()

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderItemChanged{
		OrderID:      orderID,
		RemovedItems: []uuid.UUID{itemID},
	})
}

func findItemIndex(items []model.Item, itemID uuid.UUID) (int, bool) {
	for i, item := range items {
		if item.ID == itemID {
			return i, true
		}
	}
	return -1, false
}
