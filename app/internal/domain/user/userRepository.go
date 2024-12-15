package user

import (
	"context"
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user User) error {
	query := `INSERT INTO users (id, email, password, created_at) VALUES ($1, $2, $3, $4)`
	_, err := repo.db.ExecContext(ctx, query)
	repo.db.BeginTx(ctx, &sql.TxOptions{})
	return err
}
