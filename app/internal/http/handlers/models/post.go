package models

import (
	"server/internal/domain/posts"
	"time"

	"github.com/google/uuid"
)

type CreatePostResource struct {
	Title           string `json:"title" validate:"required,max=100"`
	Content         string `json:"content" validate:"required"`
	Excerpt         string `json:"excerpt" validate:"max=500"`
	CoverImageUrl   string `json:"coverImageUrl" validate:"max=500"`
	CategoryId      string `json:"categoryId" validate:"required,uuid"`
	MetaDescription string `json:"metaDescription" validate:"max=160"`
	Status          string `json:"status" validate:"omitempty,oneof=created draft published archived"`
}

type UpdatePostResource struct {
	Title           string `json:"title" validate:"required,max=100"`
	Content         string `json:"content" validate:"required"`
	Excerpt         string `json:"excerpt" validate:"max=500"`
	CoverImageUrl   string `json:"coverImageUrl" validate:"max=500"`
	CategoryId      string `json:"categoryId" validate:"required,uuid"`
	MetaDescription string `json:"metaDescription" validate:"max=160"`
	Status          string `json:"status" validate:"required,oneof=created draft published archived"`
}

type PostResponseResource struct {
	Id                 uuid.UUID  `json:"id"`
	Title              string     `json:"title"`
	Slug               string     `json:"slug"`
	Content            string     `json:"content"`
	Excerpt            string     `json:"excerpt"`
	CoverImageUrl      string     `json:"coverImageUrl"`
	Status             string     `json:"status"`
	PublishedAt        *time.Time `json:"publishedAt"`
	MetaDescription    string     `json:"metaDescription"`
	ReadingTimeMinutes int        `json:"readingTimeMinutes"`
	CategoryId         uuid.UUID  `json:"categoryId"`
	CategoryName       string     `json:"categoryName"`
	CategorySlug       string     `json:"categorySlug"`
	AuthorFirstName    string     `json:"authorFirstName"`
	AuthorLastName     string     `json:"authorLastName"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
}

type PostListItem struct {
	Id                 uuid.UUID  `json:"id"`
	Title              string     `json:"title"`
	Slug               string     `json:"slug"`
	Excerpt            string     `json:"excerpt"`
	CoverImageUrl      string     `json:"coverImageUrl"`
	Status             string     `json:"status"`
	PublishedAt        *time.Time `json:"publishedAt"`
	ReadingTimeMinutes int        `json:"readingTimeMinutes"`
	CategoryName       string     `json:"categoryName"`
	CategorySlug       string     `json:"categorySlug"`
	AuthorFirstName    string     `json:"authorFirstName"`
	AuthorLastName     string     `json:"authorLastName"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type PaginatedResponse[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

func PostResponseFromDomain(p *posts.PostWithAuthor) PostResponseResource {
	var publishedAt *time.Time
	if p.PublishedAt.Valid {
		publishedAt = &p.PublishedAt.Time
	}

	var updatedAt *time.Time
	if p.UpdatedAt.Valid {
		updatedAt = &p.UpdatedAt.Time
	}

	return PostResponseResource{
		Id:                 p.Id,
		Title:              p.Title,
		Slug:               p.Slug,
		Content:            p.Content,
		Excerpt:            p.Excerpt,
		CoverImageUrl:      p.CoverImageUrl,
		Status:             string(p.Status),
		PublishedAt:        publishedAt,
		MetaDescription:    p.MetaDescription,
		ReadingTimeMinutes: p.ReadingTimeMinutes,
		CategoryId:         p.CategoryId,
		CategoryName:       p.CategoryName,
		CategorySlug:       p.CategorySlug,
		AuthorFirstName:    p.AuthorFirstName,
		AuthorLastName:     p.AuthorLastName,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          updatedAt,
	}
}

func PostListItemFromDomain(p *posts.PostWithAuthor) PostListItem {
	var publishedAt *time.Time
	if p.PublishedAt.Valid {
		publishedAt = &p.PublishedAt.Time
	}

	return PostListItem{
		Id:                 p.Id,
		Title:              p.Title,
		Slug:               p.Slug,
		Excerpt:            p.Excerpt,
		CoverImageUrl:      p.CoverImageUrl,
		Status:             string(p.Status),
		PublishedAt:        publishedAt,
		ReadingTimeMinutes: p.ReadingTimeMinutes,
		CategoryName:       p.CategoryName,
		CategorySlug:       p.CategorySlug,
		AuthorFirstName:    p.AuthorFirstName,
		AuthorLastName:     p.AuthorLastName,
		CreatedAt:          p.CreatedAt,
	}
}

func PostListFromDomain(domainPosts []posts.PostWithAuthor) []PostListItem {
	items := make([]PostListItem, len(domainPosts))
	for i, p := range domainPosts {
		items[i] = PostListItemFromDomain(&p)
	}
	return items
}
