package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"product/pkg/domain/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

func (r *ProductRepository) Store(product *model.Product) error {
	query := `
		INSERT INTO products (id, name, price, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			price = VALUES(price),
			updated_at = VALUES(updated_at),
			deleted_at = VALUES(deleted_at)
	`

	deletedAt := (*time.Time)(nil)
	if product.DeletedAt != nil {
		deletedAt = product.DeletedAt
	}

	_, err := r.db.Exec(query,
		product.ID.String(),
		product.Name,
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store product: %w", err)
	}

	return nil
}

func (r *ProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at, deleted_at
		FROM products
		WHERE id = ?
	`

	var product ProductRow
	var _ sql.NullTime

	err := r.db.Get(&product, query, id.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return r.rowToProduct(&product), nil
}

func (r *ProductRepository) FindByName(name string) (*model.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at, deleted_at
		FROM products
		WHERE name = ? AND deleted_at IS NULL
	`

	var product ProductRow
	var _ sql.NullTime

	err := r.db.Get(&product, query, name)
	if err == sql.ErrNoRows {
		return nil, model.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find product by name: %w", err)
	}

	return r.rowToProduct(&product), nil
}

func (r *ProductRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE products 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// CheckProductNameExists проверяет, существует ли продукт с таким именем
func (r *ProductRepository) CheckProductNameExists(name string) (bool, error) {
	query := "SELECT COUNT(*) FROM products WHERE name = ? AND deleted_at IS NULL"
	var count int
	err := r.db.Get(&count, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check product name exists: %w", err)
	}
	return count > 0, nil
}

// GetAllActiveProducts получает все активные продукты
func (r *ProductRepository) GetAllActiveProducts() ([]*model.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at, deleted_at
		FROM products
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var products []ProductRow
	err := r.db.Select(&products, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active products: %w", err)
	}

	result := make([]*model.Product, len(products))
	for i, productRow := range products {
		result[i] = r.rowToProduct(&productRow)
	}

	return result, nil
}

// GetProductsByIDs получает продукты по списку ID
func (r *ProductRepository) GetProductsByIDs(ids []uuid.UUID) ([]*model.Product, error) {
	if len(ids) == 0 {
		return []*model.Product{}, nil
	}

	// Создаем плейсхолдеры для IN запроса
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id.String()
	}

	query := fmt.Sprintf(`
		SELECT id, name, price, created_at, updated_at, deleted_at
		FROM products
		WHERE id IN (%s) AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ", "))

	var products []ProductRow
	err := r.db.Select(&products, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by IDs: %w", err)
	}

	result := make([]*model.Product, len(products))
	for i, productRow := range products {
		result[i] = r.rowToProduct(&productRow)
	}

	return result, nil
}

type ProductRow struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Price     float64      `db:"price"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (r *ProductRepository) rowToProduct(row *ProductRow) *model.Product {
	productID, _ := uuid.Parse(row.ID)

	deletedAt := (*time.Time)(nil)
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}

	return &model.Product{
		ID:        productID,
		Name:      row.Name,
		Price:     row.Price,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
