package category

import (
	"context"
	"database/sql"
)

type CategoryRepository struct {
	Db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{Db: db}
}

func (repository *CategoryRepository) Create(ctx context.Context, category Category) (*Category, error) {
	query := `INSERT INTO categories (id, name, image_url, created_at) VALUES($1, $2, $3, $4)`
	var createdCategory Category
	err := repository.Db.QueryRowContext(ctx, query, category.Id, category.Name, category.ImageUrl, category.CreatedAt).Scan(
		&createdCategory.Id, &createdCategory.Name, &createdCategory.ImageUrl, &createdCategory.CreatedAt,
		&createdCategory.UpdatedAt, &createdCategory.IsDeleted)

	return &createdCategory, err
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
