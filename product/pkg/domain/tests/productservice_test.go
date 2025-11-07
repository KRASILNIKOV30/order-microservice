package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"product/pkg/domain/model"
	"product/pkg/domain/service"
)

type testFixture struct {
	productService  service.Product
	repo            *mockProductRepository
	eventDispatcher *mockEventDispatcher
}

func setup() testFixture {
	repo := &mockProductRepository{store: make(map[uuid.UUID]*model.Product)}
	eventDispatcher := &mockEventDispatcher{}
	productService := service.NewProductService(repo, eventDispatcher)

	return testFixture{
		productService:  productService,
		repo:            repo,
		eventDispatcher: eventDispatcher,
	}
}

func TestProductService(t *testing.T) {
	name := "Digital Book"
	price := 19.99

	t.Run("Create product", func(t *testing.T) {
		f := setup()

		productID, err := f.productService.CreateProduct(name, price)

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[productID])
		require.Equal(t, name, f.repo.store[productID].Name)
		require.Equal(t, price, f.repo.store[productID].Price)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.ProductCreated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Delete product", func(t *testing.T) {
		f := setup()
		productID, _ := f.productService.CreateProduct(name, price)
		f.eventDispatcher.events = nil

		err := f.productService.DeleteProduct(productID)

		require.NoError(t, err)
		require.NotNil(t, f.repo.store[productID].DeletedAt)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.ProductDeleted{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Update product", func(t *testing.T) {
		f := setup()
		productID, _ := f.productService.CreateProduct(name, price)
		f.eventDispatcher.events = nil

		newName := "Digital Course"
		newPrice := 49.99
		err := f.productService.UpdateProduct(productID, newName, newPrice)

		require.NoError(t, err)
		require.Equal(t, newName, f.repo.store[productID].Name)
		require.Equal(t, newPrice, f.repo.store[productID].Price)
		require.Len(t, f.eventDispatcher.events, 1)
		require.Equal(t, model.ProductUpdated{}.Type(), f.eventDispatcher.events[0].Type())
	})

	t.Run("Fail to create product with duplicate name", func(t *testing.T) {
		f := setup()
		_, _ = f.productService.CreateProduct(name, price)
		f.eventDispatcher.events = nil

		_, err := f.productService.CreateProduct(name, 100.00)

		require.ErrorIs(t, err, model.ErrProductNameExists)
		require.Empty(t, f.eventDispatcher.events)
	})

	t.Run("Fail to update product to a duplicate name", func(t *testing.T) {
		f := setup()
		productID, _ := f.productService.CreateProduct("Product A", 10.0)
		_, _ = f.productService.CreateProduct("Product B", 20.0)
		f.eventDispatcher.events = nil

		err := f.productService.UpdateProduct(productID, "Product B", 30.0)

		require.ErrorIs(t, err, model.ErrProductNameExists)
		require.Empty(t, f.eventDispatcher.events)
	})
}

var _ model.ProductRepository = &mockProductRepository{}

type mockProductRepository struct {
	store map[uuid.UUID]*model.Product
}

func (m *mockProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockProductRepository) Store(product *model.Product) error {
	m.store[product.ID] = product
	return nil
}

func (m *mockProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	if product, ok := m.store[id]; ok && product.DeletedAt == nil {
		return product, nil
	}
	return nil, model.ErrProductNotFound
}

func (m *mockProductRepository) FindByName(name string) (*model.Product, error) {
	for _, product := range m.store {
		if product.Name == name && product.DeletedAt == nil {
			return product, nil
		}
	}
	return nil, model.ErrProductNotFound
}

func (m *mockProductRepository) Delete(id uuid.UUID) error {
	if product, ok := m.store[id]; ok && product.DeletedAt == nil {
		product.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrProductNotFound
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
