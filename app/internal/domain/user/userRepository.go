package user

import (
	"database/sql"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) Create(user User) (uuid, error) {
	query := `INSERT INTO users (id, email, password, status, created_at) VALUES ($1, $2, $3, $4, $5)`

	res, err := repo.db.Exec(query, user.Id, user.Email, user.Password, user.Status, user.CreatedAt)

	return err
}
