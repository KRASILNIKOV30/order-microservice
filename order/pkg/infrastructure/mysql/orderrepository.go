package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"order/pkg/domain/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

func (r *OrderRepository) Store(order *model.Order) error {
	// Сохраняем заказ
	query := `
		INSERT INTO orders (id, customer_id, status, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			customer_id = VALUES(customer_id),
			status = VALUES(status),
			updated_at = VALUES(updated_at),
			deleted_at = VALUES(deleted_at)
	`

	deletedAt := (*time.Time)(nil)
	if order.DeletedAt != nil {
		deletedAt = order.DeletedAt
	}

	_, err := r.db.Exec(query,
		order.ID.String(),
		order.CustomerID.String(),
		int(order.Status),
		order.CreatedAt,
		order.UpdatedAt,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store order: %w", err)
	}

	// Удаляем старые элементы заказа
	_, err = r.db.Exec("DELETE FROM order_items WHERE order_id = ?", order.ID.String())
	if err != nil {
		return fmt.Errorf("failed to delete old order items: %w", err)
	}

	// Сохраняем элементы заказа
	for _, item := range order.Items {
		_, err = r.db.Exec(`
			INSERT INTO order_items (id, order_id, product_id, price)
			VALUES (?, ?, ?, ?)
		`, item.ID.String(), order.ID.String(), item.ProductID.String(), item.Price)
		if err != nil {
			return fmt.Errorf("failed to store order item: %w", err)
		}
	}

	return nil
}

func (r *OrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	// Получаем заказ
	orderQuery := `
		SELECT id, customer_id, status, created_at, updated_at, deleted_at
		FROM orders
		WHERE id = ?
	`

	var order model.Order
	var deletedAt sql.NullTime

	err := r.db.Get(&order, orderQuery, id.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// Получаем элементы заказа
	itemsQuery := `
		SELECT id, product_id, price
		FROM order_items
		WHERE order_id = ?
	`

	var items []ItemRow
	err = r.db.Select(&items, itemsQuery, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}

	// Конвертируем элементы
	order.Items = make([]model.Item, len(items))
	for i, item := range items {
		order.Items[i] = model.Item{
			ID:        uuid.Must(uuid.Parse(item.ProductID)),
			ProductID: uuid.Must(uuid.Parse(item.ProductID)),
			Price:     item.Price,
		}
		order.Items[i].ID = uuid.Must(uuid.Parse(item.ID))
	}

	if deletedAt.Valid {
		deletedTime := deletedAt.Time
		order.DeletedAt = &deletedTime
	}

	return &order, nil
}

func (r *OrderRepository) Delete(id uuid.UUID) error {
	// Удаляем элементы заказа
	_, err := r.db.Exec("DELETE FROM order_items WHERE order_id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Удаляем заказ
	_, err = r.db.Exec("DELETE FROM orders WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

// Helper structs для работы с БД
type ItemRow struct {
	ID        string  `db:"id"`
	ProductID string  `db:"product_id"`
	Price     float64 `db:"price"`
}

// FindByCustomerID получает заказы по ID клиента
func (r *OrderRepository) FindByCustomerID(customerID uuid.UUID) ([]*model.Order, error) {
	query := `
		SELECT id, customer_id, status, created_at, updated_at, deleted_at
		FROM orders
		WHERE customer_id = ?
		ORDER BY created_at DESC
	`

	var orders []OrderRow
	err := r.db.Select(&orders, query, customerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by customer: %w", err)
	}

	result := make([]*model.Order, len(orders))
	for i, orderRow := range orders {
		order, err := r.rowToOrder(&orderRow)
		if err != nil {
			return nil, fmt.Errorf("failed to convert order row: %w", err)
		}
		result[i] = order
	}

	return result, nil
}

// FindByStatus получает заказы по статусу
func (r *OrderRepository) FindByStatus(status model.OrderStatus) ([]*model.Order, error) {
	query := `
		SELECT id, customer_id, status, created_at, updated_at, deleted_at
		FROM orders
		WHERE status = ?
		ORDER BY created_at DESC
	`

	var orders []OrderRow
	err := r.db.Select(&orders, query, int(status))
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by status: %w", err)
	}

	result := make([]*model.Order, len(orders))
	for i, orderRow := range orders {
		order, err := r.rowToOrder(&orderRow)
		if err != nil {
			return nil, fmt.Errorf("failed to convert order row: %w", err)
		}
		result[i] = order
	}

	return result, nil
}

type OrderRow struct {
	ID         string       `db:"id"`
	CustomerID string       `db:"customer_id"`
	Status     int          `db:"status"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	DeletedAt  sql.NullTime `db:"deleted_at"`
}

func (r *OrderRepository) rowToOrder(row *OrderRow) (*model.Order, error) {
	orderID, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	customerID, err := uuid.Parse(row.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	order := &model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.OrderStatus(row.Status),
		Items:      []model.Item{},
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}

	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		order.DeletedAt = &deletedAt
	}

	// Получаем элементы заказа
	itemsQuery := `
		SELECT id, product_id, price
		FROM order_items
		WHERE order_id = ?
	`

	var items []ItemRow
	err = r.db.Select(&items, itemsQuery, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}

	order.Items = make([]model.Item, len(items))
	for i, item := range items {
		itemID, err := uuid.Parse(item.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid item ID: %w", err)
		}

		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}

		order.Items[i] = model.Item{
			ID:        itemID,
			ProductID: productID,
			Price:     item.Price,
		}
	}

	return order, nil
}
