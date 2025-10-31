package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"order/pkg/domain/model"
	"order/pkg/domain/service"
)

type testFixture struct {
	orderService    service.Order
	repo            *mockOrderRepository
	eventDispatcher *mockEventDispatcher
}

func setup() testFixture {
	repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
	eventDispatcher := &mockEventDispatcher{}
	orderService := service.NewOrderService(repo, eventDispatcher)

	return testFixture{
		orderService:    orderService,
		repo:            repo,
		eventDispatcher: eventDispatcher,
	}
}

func TestOrderService(t *testing.T) {
	customerID := uuid.Must(uuid.NewV7())
	productID := uuid.Must(uuid.NewV7())

	t.Run("Create order", func(t *testing.T) {
		f := setup()

		orderID, err := f.orderService.CreateOrder(customerID)

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[orderID])
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.OrderCreated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Delete order", func(t *testing.T) {
		f := setup()
		orderID, _ := f.orderService.CreateOrder(customerID)
		f.eventDispatcher.events = nil

		err := f.orderService.DeleteOrder(orderID)

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[orderID].DeletedAt)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.OrderDeleted{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Set status", func(t *testing.T) {
		f := setup()
		orderID, _ := f.orderService.CreateOrder(customerID)
		f.eventDispatcher.events = nil

		err := f.orderService.SetStatus(orderID, model.Paid)

		require.NoError(t, err)
		require.Equal(t, model.Paid, f.repo.store[orderID].Status)
		require.Len(t, f.eventDispatcher.events, 1)
		event := f.eventDispatcher.events[0].(model.OrderStatusChanged)
		require.Equal(t, model.OrderStatusChanged{}.Type(), event.Type())
		require.Equal(t, model.Open, event.OldStatus)
		require.Equal(t, model.Paid, event.NewStatus)
	})

	t.Run("Add item to order", func(t *testing.T) {
		f := setup()
		orderID, _ := f.orderService.CreateOrder(customerID)
		f.eventDispatcher.events = nil

		itemID, err := f.orderService.AddItem(orderID, productID, 99.99)

		require.NoError(t, err)
		require.Len(t, f.repo.store[orderID].Items, 1)
		require.Equal(t, itemID, f.repo.store[orderID].Items[0].ID)
		require.Len(t, f.eventDispatcher.events, 1)
		event := f.eventDispatcher.events[0].(model.OrderItemChanged)
		require.Equal(t, model.OrderItemChanged{}.Type(), event.Type())
		require.Equal(t, []uuid.UUID{itemID}, event.AddedItems)
	})

	t.Run("Delete item from order", func(t *testing.T) {
		f := setup()
		orderID, _ := f.orderService.CreateOrder(customerID)
		itemID, _ := f.orderService.AddItem(orderID, productID, 99.99)
		f.eventDispatcher.events = nil

		err := f.orderService.DeleteItem(orderID, itemID)

		require.NoError(t, err)
		require.Empty(t, f.repo.store[orderID].Items)
		require.Len(t, f.eventDispatcher.events, 1)
		event := f.eventDispatcher.events[0].(model.OrderItemChanged)
		require.Equal(t, model.OrderItemChanged{}.Type(), event.Type())
		require.Equal(t, []uuid.UUID{itemID}, event.RemovedItems)
	})

	t.Run("Fail to add item to non-open order", func(t *testing.T) {
		f := setup()
		orderID, _ := f.orderService.CreateOrder(customerID)
		_ = f.orderService.SetStatus(orderID, model.Paid)
		f.eventDispatcher.events = nil

		_, err := f.orderService.AddItem(orderID, productID, 99.99)

		require.ErrorIs(t, err, service.ErrInvalidOrderStatus)
		require.Empty(t, f.eventDispatcher.events)
	})
}

var _ model.OrderRepository = &mockOrderRepository{}

type mockOrderRepository struct {
	store map[uuid.UUID]*model.Order
}

func (m *mockOrderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockOrderRepository) Store(order *model.Order) error {
	m.store[order.ID] = order
	return nil
}

func (m *mockOrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		return order, nil
	}
	return nil, model.ErrOrderNotFound
}

func (m *mockOrderRepository) Delete(id uuid.UUID) error {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		order.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrOrderNotFound
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
