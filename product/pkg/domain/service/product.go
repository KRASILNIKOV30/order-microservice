package service

import (
	"time"

	"github.com/google/uuid"

	"product/pkg/domain/model"
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Product interface {
	CreateProduct(name string, price float64) (uuid.UUID, error)
	UpdateProduct(productID uuid.UUID, name string, price float64) error
	DeleteProduct(productID uuid.UUID) error
}

func NewProductService(repo model.ProductRepository, dispatcher EventDispatcher) Product {
	return &productService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type productService struct {
	repo       model.ProductRepository
	dispatcher EventDispatcher
}

func (s *productService) CreateProduct(name string, price float64) (uuid.UUID, error) {
	if _, err := s.repo.FindByName(name); err == nil {
		return uuid.Nil, model.ErrProductNameExists
	}

	productID, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	product := &model.Product{
		ID:        productID,
		Name:      name,
		Price:     price,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := s.repo.Store(product); err != nil {
		return uuid.Nil, err
	}

	return productID, s.dispatcher.Dispatch(model.ProductCreated{
		ProductID: productID,
		Name:      name,
		Price:     price,
	})
}

func (s *productService) UpdateProduct(productID uuid.UUID, name string, price float64) error {
	product, err := s.repo.Find(productID)
	if err != nil {
		return err
	}

	if product.Name != name {
		if _, err := s.repo.FindByName(name); err == nil {
			return model.ErrProductNameExists
		}
		product.Name = name
	}

	product.Price = price
	product.UpdatedAt = time.Now()

	if err := s.repo.Store(product); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.ProductUpdated{
		ProductID: productID,
	})
}

func (s *productService) DeleteProduct(productID uuid.UUID) error {
	if err := s.repo.Delete(productID); err != nil {
		return err
	}

	return s.dispatcher.Dispatch(model.ProductDeleted{
		ProductID: productID,
	})
}
