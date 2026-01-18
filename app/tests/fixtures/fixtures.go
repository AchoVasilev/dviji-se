package fixtures

import (
	"database/sql"
	"server/internal/domain/category"
	"server/internal/domain/posts"
	"server/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

// sql is used for NullTime and NullString in user fixtures
var _ = sql.NullString{}

// User fixtures
var (
	TestUserID       = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	TestAdminID      = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	TestUserRoleID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	TestAdminRoleID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestUser() user.User {
	return user.User{
		Id:        TestUserID,
		Email:     "test@example.com",
		Password:  "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.G1vO0yL0B9JQHi", // "password123"
		FirstName: sql.NullString{String: "Test", Valid: true},
		LastName:  sql.NullString{String: "User", Valid: true},
		Status:    "ACTIVE",
		CreatedAt: time.Now().UTC(),
		IsDeleted: false,
		Roles:     []user.Role{{Id: TestUserRoleID, Name: "USER"}},
	}
}

func TestAdmin() user.User {
	return user.User{
		Id:        TestAdminID,
		Email:     "admin@example.com",
		Password:  "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.G1vO0yL0B9JQHi", // "password123"
		FirstName: sql.NullString{String: "Admin", Valid: true},
		LastName:  sql.NullString{String: "User", Valid: true},
		Status:    "ACTIVE",
		CreatedAt: time.Now().UTC(),
		IsDeleted: false,
		Roles:     []user.Role{{Id: TestAdminRoleID, Name: "ADMIN"}},
	}
}

func TestUserWithPassword(password string) user.User {
	u := TestUser()
	u.Password = password
	return u
}

// Category fixtures
var (
	TestCategoryID = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
)

func TestCategory() category.Category {
	return category.Category{
		Id:        TestCategoryID,
		Name:      "Test Category",
		Slug:      "test-category",
		ImageUrl:  "https://example.com/image.jpg",
		CreatedAt: time.Now().UTC(),
		IsDeleted: false,
	}
}

// Post fixtures
var (
	TestPostID = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
)

func TestPost() posts.Post {
	return posts.Post{
		Id:                 TestPostID,
		Title:              "Test Post Title",
		Slug:               "test-post-title",
		Content:            "This is the content of the test post. It has some words to test reading time calculation.",
		Excerpt:            "Test excerpt",
		Status:             posts.PostStatusPublished,
		CategoryId:         TestCategoryID,
		CreatorUserId:      TestUserID,
		ReadingTimeMinutes: 1,
		CreatedAt:          time.Now().UTC(),
		IsDeleted:          false,
	}
}

func TestDraftPost() posts.Post {
	p := TestPost()
	p.Id = uuid.New()
	p.Slug = "draft-post"
	p.Status = posts.PostStatusDraft
	return p
}

func TestPublishedPost() posts.Post {
	p := TestPost()
	p.Status = posts.PostStatusPublished
	now := time.Now().UTC()
	p.PublishedAt = sql.NullTime{Time: now, Valid: true}
	return p
}

// JWT test keys
const (
	TestJWTKey        = "test-jwt-secret-key-for-testing-only"
	TestJWTRefreshKey = "test-jwt-refresh-secret-key-for-testing"
)
