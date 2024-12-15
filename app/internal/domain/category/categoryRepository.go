package category

import (
	"context"
	"database/sql"
	"server/internal/infrastructure/database"
)

type CategoryRepository struct {
	Db *sql.DB
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{Db: database.Db}
}

func (repository *CategoryRepository) Create(ctx context.Context, category Category) error {
	query := `INSERT INTO categories (id, name, image_url, created_at) VALUES($1, $2, $3, $4)`
	_, err := repository.Db.ExecContext(ctx, query, category.Id, category.Name, category.ImageUrl, category.CreatedAt)

	return err
}

func (repository *CategoryRepository) FindAll(ctx context.Context) ([]Category, error) {
	rows, err := repository.Db.QueryContext(ctx, `SELECT * FROM categories c WHERE c.is_deleted = FALSE`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}
