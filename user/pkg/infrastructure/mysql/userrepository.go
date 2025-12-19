package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"user/pkg/domain/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) NextID() (uuid.UUID, error) {
	return uuid.NewUUID()
}

func (r *UserRepository) Store(user *model.User) error {
	query := `
		INSERT INTO users (id, login, email, tg, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			login = VALUES(login),
			email = VALUES(email),
			tg = VALUES(tg),
			updated_at = VALUES(updated_at),
			deleted_at = VALUES(deleted_at)
	`

	deletedAt := (*time.Time)(nil)
	if user.DeletedAt != nil {
		deletedAt = user.DeletedAt
	}

	tg := (*string)(nil)
	if user.Tg != nil {
		tg = user.Tg
	}

	_, err := r.db.Exec(query,
		user.ID.String(),
		user.Login,
		user.Email,
		tg,
		user.CreatedAt,
		user.UpdatedAt,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	return nil
}

func (r *UserRepository) Find(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, login, email, tg, created_at, updated_at, deleted_at
		FROM users
		WHERE id = ?
	`

	var user UserRow

	err := r.db.Get(&user, query, id.String())
	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.rowToUser(&user), nil
}

func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	query := `
		SELECT id, login, email, tg, created_at, updated_at, deleted_at
		FROM users
		WHERE login = ?
	`

	var user UserRow

	err := r.db.Get(&user, query, login)
	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by login: %w", err)
	}

	return r.rowToUser(&user), nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, login, email, tg, created_at, updated_at, deleted_at
		FROM users
		WHERE email = ?
	`

	var user UserRow

	err := r.db.Get(&user, query, email)
	if err == sql.ErrNoRows {
		return nil, model.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return r.rowToUser(&user), nil
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *UserRepository) CheckLoginExists(login string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE login = ? AND deleted_at IS NULL"
	var count int
	err := r.db.Get(&count, query, login)
	if err != nil {
		return false, fmt.Errorf("failed to check login exists: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE email = ? AND deleted_at IS NULL"
	var count int
	err := r.db.Get(&count, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email exists: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepository) GetAllActiveUsers() ([]*model.User, error) {
	query := `
		SELECT id, login, email, tg, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var users []UserRow
	err := r.db.Select(&users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active users: %w", err)
	}

	result := make([]*model.User, len(users))
	for i, userRow := range users {
		result[i] = r.rowToUser(&userRow)
	}

	return result, nil
}

type UserRow struct {
	ID        string         `db:"id"`
	Login     string         `db:"login"`
	Email     string         `db:"email"`
	Tg        sql.NullString `db:"tg"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

func (r *UserRepository) rowToUser(row *UserRow) *model.User {
	userID, _ := uuid.Parse(row.ID)

	tg := (*string)(nil)
	if row.Tg.Valid {
		tgStr := row.Tg.String
		tg = &tgStr
	}

	deletedAt := (*time.Time)(nil)
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}

	return &model.User{
		ID:        userID,
		Login:     row.Login,
		Email:     row.Email,
		Tg:        tg,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
