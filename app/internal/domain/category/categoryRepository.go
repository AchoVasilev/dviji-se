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
	query := `INSERT INTO categories (id, name, slug, image_url, created_at)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, name, slug, image_url, created_at, updated_at, is_deleted`
	var createdCategory Category
	err := repository.Db.QueryRowContext(ctx, query, category.Id, category.Name, category.Slug, category.ImageUrl, category.CreatedAt).Scan(
		&createdCategory.Id, &createdCategory.Name, &createdCategory.Slug, &createdCategory.ImageUrl, &createdCategory.CreatedAt,
		&createdCategory.UpdatedAt, &createdCategory.IsDeleted)

	return &createdCategory, err
}

func (repository *CategoryRepository) FindAll(ctx context.Context) ([]Category, error) {
	rows, err := repository.Db.QueryContext(ctx, `SELECT id, name, slug, image_url, created_at, updated_at, is_deleted FROM categories WHERE is_deleted = FALSE`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Id, &category.Name, &category.Slug, &category.ImageUrl, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (repository *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*Category, error) {
	query := `SELECT id, name, slug, image_url, created_at, updated_at, is_deleted FROM categories WHERE slug = $1 AND is_deleted = FALSE`
	var category Category
	err := repository.Db.QueryRowContext(ctx, query, slug).Scan(
		&category.Id, &category.Name, &category.Slug, &category.ImageUrl, &category.CreatedAt,
		&category.UpdatedAt, &category.IsDeleted)

	if err != nil {
		return nil, err
	}
	return &category, nil
}
