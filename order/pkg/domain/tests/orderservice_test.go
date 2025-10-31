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
	if order, ok := m.store[id]; ok && order.DeletedAt != nil {
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
