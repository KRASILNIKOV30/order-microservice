package service

import (
	"context"
	"order/pkg/domain/model"
)

type RepositoryProvider interface {
	OrderRepository(ctx context.Context) model.OrderRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockName string, f func(provider RepositoryProvider) error) error
}

type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
