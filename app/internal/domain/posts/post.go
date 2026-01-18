package posts

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PostStatus string

const (
	PostStatusCreated   PostStatus = "created"
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

type Post struct {
	Id                 uuid.UUID    `json:"id"`
	Title              string       `json:"title"`
	Slug               string       `json:"slug"`
	Content            string       `json:"content"`
	Excerpt            string       `json:"excerpt"`
	CoverImageUrl      string       `json:"cover_image_url"`
	Status             PostStatus   `json:"status"`
	PublishedAt        sql.NullTime `json:"published_at"`
	MetaDescription    string       `json:"meta_description"`
	ReadingTimeMinutes int          `json:"reading_time_minutes"`
	CategoryId         uuid.UUID    `json:"category_id"`
	CreatorUserId      uuid.UUID    `json:"creator_user_id"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          sql.NullTime `json:"updated_at"`
	UpdatedBy          string       `json:"updated_by"`
	IsDeleted          bool         `json:"is_deleted"`
}

func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished
}

type PostWithAuthor struct {
	Post
	AuthorFirstName string `json:"author_first_name"`
	AuthorLastName  string `json:"author_last_name"`
	CategoryName    string `json:"category_name"`
	CategorySlug    string `json:"category_slug"`
}
