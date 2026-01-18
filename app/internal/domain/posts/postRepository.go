package posts

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostRepository struct {
	Db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{Db: db}
}

func (r *PostRepository) Create(ctx context.Context, post Post) (*Post, error) {
	query := `
		INSERT INTO posts (id, title, slug, content, excerpt, cover_image_url, status, published_at,
			meta_description, reading_time_minutes, category_id, creator_user_id, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, title, slug, content, excerpt, cover_image_url, status, published_at,
			meta_description, reading_time_minutes, category_id, creator_user_id, created_at, updated_at, updated_by, is_deleted`

	var createdPost Post
	var excerpt, coverImageUrl, metaDescription, updatedBy sql.NullString

	err := r.Db.QueryRowContext(ctx, query,
		post.Id, post.Title, post.Slug, post.Content, toNullString(post.Excerpt),
		toNullString(post.CoverImageUrl), post.Status, post.PublishedAt,
		toNullString(post.MetaDescription), post.ReadingTimeMinutes, post.CategoryId,
		post.CreatorUserId, post.CreatedAt,
	).Scan(
		&createdPost.Id, &createdPost.Title, &createdPost.Slug, &createdPost.Content,
		&excerpt, &coverImageUrl, &createdPost.Status, &createdPost.PublishedAt,
		&metaDescription, &createdPost.ReadingTimeMinutes, &createdPost.CategoryId,
		&createdPost.CreatorUserId, &createdPost.CreatedAt, &createdPost.UpdatedAt,
		&updatedBy, &createdPost.IsDeleted,
	)

	createdPost.Excerpt = excerpt.String
	createdPost.CoverImageUrl = coverImageUrl.String
	createdPost.MetaDescription = metaDescription.String
	createdPost.UpdatedBy = updatedBy.String

	return &createdPost, err
}

func (r *PostRepository) Update(ctx context.Context, post Post) (*Post, error) {
	query := `
		UPDATE posts SET title = $1, slug = $2, content = $3, excerpt = $4, cover_image_url = $5,
			status = $6, published_at = $7, meta_description = $8, reading_time_minutes = $9,
			category_id = $10, updated_at = NOW(), updated_by = $11
		WHERE id = $12 AND is_deleted = FALSE
		RETURNING id, title, slug, content, excerpt, cover_image_url, status, published_at,
			meta_description, reading_time_minutes, category_id, creator_user_id, created_at, updated_at, updated_by, is_deleted`

	var updatedPost Post
	var excerpt, coverImageUrl, metaDescription, updatedBy sql.NullString

	err := r.Db.QueryRowContext(ctx, query,
		post.Title, post.Slug, post.Content, toNullString(post.Excerpt),
		toNullString(post.CoverImageUrl), post.Status, post.PublishedAt,
		toNullString(post.MetaDescription), post.ReadingTimeMinutes, post.CategoryId,
		post.UpdatedBy, post.Id,
	).Scan(
		&updatedPost.Id, &updatedPost.Title, &updatedPost.Slug, &updatedPost.Content,
		&excerpt, &coverImageUrl, &updatedPost.Status, &updatedPost.PublishedAt,
		&metaDescription, &updatedPost.ReadingTimeMinutes, &updatedPost.CategoryId,
		&updatedPost.CreatorUserId, &updatedPost.CreatedAt, &updatedPost.UpdatedAt,
		&updatedBy, &updatedPost.IsDeleted,
	)

	updatedPost.Excerpt = excerpt.String
	updatedPost.CoverImageUrl = coverImageUrl.String
	updatedPost.MetaDescription = metaDescription.String
	updatedPost.UpdatedBy = updatedBy.String

	return &updatedPost, err
}

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
	query := `UPDATE posts SET is_deleted = TRUE, updated_at = NOW(), updated_by = $1 WHERE id = $2`
	_, err := r.Db.ExecContext(ctx, query, deletedBy, id)
	return err
}

func (r *PostRepository) FindById(ctx context.Context, id uuid.UUID) (*Post, error) {
	query := `
		SELECT id, title, slug, content, excerpt, cover_image_url, status, published_at,
			meta_description, reading_time_minutes, category_id, creator_user_id, created_at, updated_at, updated_by, is_deleted
		FROM posts WHERE id = $1 AND is_deleted = FALSE`

	var post Post
	var excerpt, coverImageUrl, metaDescription, updatedBy sql.NullString

	err := r.Db.QueryRowContext(ctx, query, id).Scan(
		&post.Id, &post.Title, &post.Slug, &post.Content,
		&excerpt, &coverImageUrl, &post.Status, &post.PublishedAt,
		&metaDescription, &post.ReadingTimeMinutes, &post.CategoryId,
		&post.CreatorUserId, &post.CreatedAt, &post.UpdatedAt,
		&updatedBy, &post.IsDeleted,
	)

	post.Excerpt = excerpt.String
	post.CoverImageUrl = coverImageUrl.String
	post.MetaDescription = metaDescription.String
	post.UpdatedBy = updatedBy.String

	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindBySlug(ctx context.Context, slug string) (*PostWithAuthor, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.slug = $1 AND p.is_deleted = FALSE`

	var post PostWithAuthor
	var excerpt, coverImageUrl, metaDescription, updatedBy, firstName, lastName sql.NullString

	err := r.Db.QueryRowContext(ctx, query, slug).Scan(
		&post.Id, &post.Title, &post.Slug, &post.Content,
		&excerpt, &coverImageUrl, &post.Status, &post.PublishedAt,
		&metaDescription, &post.ReadingTimeMinutes, &post.CategoryId,
		&post.CreatorUserId, &post.CreatedAt, &post.UpdatedAt,
		&updatedBy, &post.IsDeleted,
		&firstName, &lastName, &post.CategoryName, &post.CategorySlug,
	)

	post.Excerpt = excerpt.String
	post.CoverImageUrl = coverImageUrl.String
	post.MetaDescription = metaDescription.String
	post.UpdatedBy = updatedBy.String
	post.AuthorFirstName = firstName.String
	post.AuthorLastName = lastName.String

	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindPublished(ctx context.Context, limit, offset int) ([]PostWithAuthor, int, error) {
	countQuery := `SELECT COUNT(*) FROM posts WHERE status = 'published' AND is_deleted = FALSE`
	var total int
	if err := r.Db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'published' AND p.is_deleted = FALSE
		ORDER BY p.published_at DESC
		LIMIT $1 OFFSET $2`

	return r.queryPostsWithAuthor(ctx, query, total, limit, offset)
}

func (r *PostRepository) FindByCategory(ctx context.Context, categorySlug string, limit, offset int) ([]PostWithAuthor, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM posts p
		JOIN categories c ON p.category_id = c.id
		WHERE c.slug = $1 AND p.status = 'published' AND p.is_deleted = FALSE`
	var total int
	if err := r.Db.QueryRowContext(ctx, countQuery, categorySlug).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE c.slug = $1 AND p.status = 'published' AND p.is_deleted = FALSE
		ORDER BY p.published_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.Db.QueryContext(ctx, query, categorySlug, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts, err := r.scanPostsWithAuthor(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *PostRepository) FindAll(ctx context.Context, limit, offset int) ([]PostWithAuthor, int, error) {
	countQuery := `SELECT COUNT(*) FROM posts WHERE is_deleted = FALSE`
	var total int
	if err := r.Db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.is_deleted = FALSE
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	return r.queryPostsWithAuthor(ctx, query, total, limit, offset)
}

func (r *PostRepository) FindByStatus(ctx context.Context, status PostStatus, limit, offset int) ([]PostWithAuthor, int, error) {
	countQuery := `SELECT COUNT(*) FROM posts WHERE status = $1 AND is_deleted = FALSE`
	var total int
	if err := r.Db.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.status = $1 AND p.is_deleted = FALSE
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.Db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts, err := r.scanPostsWithAuthor(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *PostRepository) FindRecent(ctx context.Context, limit int) ([]PostWithAuthor, error) {
	query := `
		SELECT p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url, p.status, p.published_at,
			p.meta_description, p.reading_time_minutes, p.category_id, p.creator_user_id, p.created_at, p.updated_at, p.updated_by, p.is_deleted,
			u.first_name, u.last_name, c.name, c.slug
		FROM posts p
		JOIN users u ON p.creator_user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'published' AND p.is_deleted = FALSE
		ORDER BY p.published_at DESC
		LIMIT $1`

	rows, err := r.Db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPostsWithAuthor(rows)
}

func (r *PostRepository) ExistsBySlug(ctx context.Context, slug string, excludeId *uuid.UUID) (bool, error) {
	var query string
	var args []interface{}

	if excludeId != nil {
		query = `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND id != $2 AND is_deleted = FALSE)`
		args = []interface{}{slug, *excludeId}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND is_deleted = FALSE)`
		args = []interface{}{slug}
	}

	var exists bool
	err := r.Db.QueryRowContext(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (r *PostRepository) queryPostsWithAuthor(ctx context.Context, query string, total int, limit, offset int) ([]PostWithAuthor, int, error) {
	rows, err := r.Db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts, err := r.scanPostsWithAuthor(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *PostRepository) scanPostsWithAuthor(rows *sql.Rows) ([]PostWithAuthor, error) {
	var posts []PostWithAuthor
	for rows.Next() {
		var post PostWithAuthor
		var excerpt, coverImageUrl, metaDescription, updatedBy, firstName, lastName sql.NullString

		err := rows.Scan(
			&post.Id, &post.Title, &post.Slug, &post.Content,
			&excerpt, &coverImageUrl, &post.Status, &post.PublishedAt,
			&metaDescription, &post.ReadingTimeMinutes, &post.CategoryId,
			&post.CreatorUserId, &post.CreatedAt, &post.UpdatedAt,
			&updatedBy, &post.IsDeleted,
			&firstName, &lastName, &post.CategoryName, &post.CategorySlug,
		)
		if err != nil {
			return nil, err
		}

		post.Excerpt = excerpt.String
		post.CoverImageUrl = coverImageUrl.String
		post.MetaDescription = metaDescription.String
		post.UpdatedBy = updatedBy.String
		post.AuthorFirstName = firstName.String
		post.AuthorLastName = lastName.String

		posts = append(posts, post)
	}
	return posts, nil
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
