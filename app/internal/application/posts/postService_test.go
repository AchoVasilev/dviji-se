package posts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server/internal/domain/posts"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockPostRepository implements postRepository interface for testing
type mockPostRepository struct {
	mu      sync.RWMutex
	posts   map[uuid.UUID]posts.Post
	bySlug  map[string]uuid.UUID
	errOnOp error // Set this to simulate errors
}

func newMockPostRepository() *mockPostRepository {
	return &mockPostRepository{
		posts:  make(map[uuid.UUID]posts.Post),
		bySlug: make(map[string]uuid.UUID),
	}
}

func (r *mockPostRepository) Create(ctx context.Context, p posts.Post) (*posts.Post, error) {
	if r.errOnOp != nil {
		return nil, r.errOnOp
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.posts[p.Id] = p
	r.bySlug[p.Slug] = p.Id
	return &p, nil
}

func (r *mockPostRepository) Update(ctx context.Context, p posts.Post) (*posts.Post, error) {
	if r.errOnOp != nil {
		return nil, r.errOnOp
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.posts[p.Id]; !exists {
		return nil, sql.ErrNoRows
	}
	r.posts[p.Id] = p
	r.bySlug[p.Slug] = p.Id
	return &p, nil
}

func (r *mockPostRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
	if r.errOnOp != nil {
		return r.errOnOp
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	p, exists := r.posts[id]
	if !exists {
		return sql.ErrNoRows
	}
	p.IsDeleted = true
	r.posts[id] = p
	return nil
}

func (r *mockPostRepository) FindById(ctx context.Context, id uuid.UUID) (*posts.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.posts[id]
	if !exists || p.IsDeleted {
		return nil, sql.ErrNoRows
	}
	return &p, nil
}

func (r *mockPostRepository) FindBySlug(ctx context.Context, slug string) (*posts.PostWithAuthor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.bySlug[slug]
	if !exists {
		return nil, sql.ErrNoRows
	}
	p, exists := r.posts[id]
	if !exists || p.IsDeleted {
		return nil, sql.ErrNoRows
	}
	return &posts.PostWithAuthor{Post: p}, nil
}

func (r *mockPostRepository) FindPublished(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []posts.PostWithAuthor
	for _, p := range r.posts {
		if !p.IsDeleted && p.Status == posts.PostStatusPublished {
			result = append(result, posts.PostWithAuthor{Post: p})
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []posts.PostWithAuthor{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockPostRepository) FindByCategory(ctx context.Context, categorySlug string, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	return []posts.PostWithAuthor{}, 0, nil
}

func (r *mockPostRepository) FindAll(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []posts.PostWithAuthor
	for _, p := range r.posts {
		if !p.IsDeleted {
			result = append(result, posts.PostWithAuthor{Post: p})
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []posts.PostWithAuthor{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockPostRepository) FindByStatus(ctx context.Context, status posts.PostStatus, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []posts.PostWithAuthor
	for _, p := range r.posts {
		if !p.IsDeleted && p.Status == status {
			result = append(result, posts.PostWithAuthor{Post: p})
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []posts.PostWithAuthor{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockPostRepository) FindRecent(ctx context.Context, limit int) ([]posts.PostWithAuthor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []posts.PostWithAuthor
	for _, p := range r.posts {
		if !p.IsDeleted && p.Status == posts.PostStatusPublished {
			result = append(result, posts.PostWithAuthor{Post: p})
		}
	}
	if limit > len(result) {
		limit = len(result)
	}
	return result[:limit], nil
}

func (r *mockPostRepository) ExistsBySlug(ctx context.Context, slug string, excludeId *uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.bySlug[slug]
	if !exists {
		return false, nil
	}
	if excludeId != nil && id == *excludeId {
		return false, nil
	}
	p, exists := r.posts[id]
	return exists && !p.IsDeleted, nil
}

func (r *mockPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []posts.PostWithAuthor
	for _, p := range r.posts {
		if !p.IsDeleted && p.Status == posts.PostStatusPublished &&
			(strings.Contains(strings.ToLower(p.Title), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(p.Content), strings.ToLower(query))) {
			result = append(result, posts.PostWithAuthor{Post: p})
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []posts.PostWithAuthor{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (r *mockPostRepository) addPost(p posts.Post) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.posts[p.Id] = p
	r.bySlug[p.Slug] = p.Id
}

// Tests

func TestGenerateSlug(t *testing.T) {
	service := NewPostService(newMockPostRepository())

	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"simple english", "Hello World", "hello-world"},
		{"with numbers", "Top 10 Tips", "top-10-tips"},
		{"cyrillic", "Привет мир", "privet-mir"},
		{"bulgarian cyrillic", "Как да отслабнем", "kak-da-otslabnem"},
		{"mixed cyrillic and latin", "Hello Привет", "hello-privet"},
		{"special characters", "Hello! World? Test.", "hello-world-test"},
		{"multiple spaces", "Hello    World", "hello-world"},
		{"leading trailing spaces", "  Hello World  ", "hello-world"},
		{"with dashes", "Hello-World-Test", "hello-world-test"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"numbers only", "12345", "12345"},
		{"cyrillic ш", "Шоколад", "shokolad"},
		{"cyrillic щ", "Щастие", "shtastie"},
		{"cyrillic ю", "Юни", "yuni"},
		{"cyrillic я", "Ябълка", "yabalka"},
		{"cyrillic ъ", "Българин", "balgarin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GenerateSlug(tt.title)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestGenerateSlug_LongTitle(t *testing.T) {
	service := NewPostService(newMockPostRepository())

	// Create a title longer than 100 characters
	longTitle := strings.Repeat("word ", 30) // 150+ characters
	slug := service.GenerateSlug(longTitle)

	if len(slug) > 100 {
		t.Errorf("GenerateSlug() should limit slug to 100 chars, got %d", len(slug))
	}

	// Should cut at word boundary (last dash)
	if strings.HasSuffix(slug, "-") {
		t.Error("GenerateSlug() should not end with dash")
	}
}

func TestCalculateReadingTime(t *testing.T) {
	service := NewPostService(newMockPostRepository())

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"empty content", "", 1},
		{"few words", "Hello world", 1},
		{"200 words", strings.Repeat("word ", 200), 1},
		{"400 words", strings.Repeat("word ", 400), 2},
		{"600 words", strings.Repeat("word ", 600), 3},
		{"1000 words", strings.Repeat("word ", 1000), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CalculateReadingTime(tt.content)
			if got != tt.expected {
				t.Errorf("CalculateReadingTime() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	creatorId := uuid.New()
	categoryId := uuid.New()

	t.Run("create post with default status", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		input := CreatePostInput{
			Title:           "Test Post",
			Content:         "This is test content with some words.",
			Excerpt:         "Test excerpt",
			CoverImageUrl:   "https://example.com/image.jpg",
			CategoryId:      categoryId,
			MetaDescription: "Test meta",
			Status:          "", // Empty should default to created
		}

		post, err := service.Create(ctx, input, creatorId)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if post.Title != input.Title {
			t.Errorf("Create() Title = %v, want %v", post.Title, input.Title)
		}

		if post.Slug != "test-post" {
			t.Errorf("Create() Slug = %v, want test-post", post.Slug)
		}

		if post.Status != posts.PostStatusCreated {
			t.Errorf("Create() Status = %v, want %v", post.Status, posts.PostStatusCreated)
		}

		if post.CreatorUserId != creatorId {
			t.Errorf("Create() CreatorUserId = %v, want %v", post.CreatorUserId, creatorId)
		}

		if post.ReadingTimeMinutes < 1 {
			t.Error("Create() ReadingTimeMinutes should be at least 1")
		}

		if post.PublishedAt.Valid {
			t.Error("Create() PublishedAt should not be set for non-published post")
		}
	})

	t.Run("create published post sets published_at", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		input := CreatePostInput{
			Title:      "Published Post",
			Content:    "Content",
			CategoryId: categoryId,
			Status:     posts.PostStatusPublished,
		}

		post, err := service.Create(ctx, input, creatorId)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if post.Status != posts.PostStatusPublished {
			t.Errorf("Create() Status = %v, want %v", post.Status, posts.PostStatusPublished)
		}

		if !post.PublishedAt.Valid {
			t.Error("Create() PublishedAt should be set for published post")
		}

		if post.PublishedAt.Time.After(time.Now().Add(time.Second)) {
			t.Error("Create() PublishedAt should be approximately now")
		}
	})

	t.Run("create post with metadata", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		metadata := json.RawMessage(`{"gpxFileUrl":"https://example.com/route.gpx"}`)

		input := CreatePostInput{
			Title:      "Hiking Post",
			Content:    "A trail through the mountains.",
			CategoryId: categoryId,
			Status:     posts.PostStatusPublished,
			Metadata:   metadata,
		}

		post, err := service.Create(ctx, input, creatorId)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if post.Metadata == nil {
			t.Fatal("Create() Metadata should not be nil")
		}

		var m map[string]interface{}
		if err := json.Unmarshal(post.Metadata, &m); err != nil {
			t.Fatalf("Create() Metadata unmarshal error = %v", err)
		}

		if m["gpxFileUrl"] != "https://example.com/route.gpx" {
			t.Errorf("Create() Metadata gpxFileUrl = %v, want https://example.com/route.gpx", m["gpxFileUrl"])
		}
	})

	t.Run("create post without metadata", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		input := CreatePostInput{
			Title:      "Simple Post",
			Content:    "No metadata here.",
			CategoryId: categoryId,
		}

		post, err := service.Create(ctx, input, creatorId)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if post.Metadata != nil && string(post.Metadata) != "null" {
			var m map[string]interface{}
			if err := json.Unmarshal(post.Metadata, &m); err == nil && len(m) > 0 {
				t.Error("Create() Metadata should be empty when not provided")
			}
		}
	})

	t.Run("create with duplicate slug appends uuid", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		// Add existing post with the same slug
		existingPost := posts.Post{
			Id:   uuid.New(),
			Slug: "test-post",
		}
		repo.addPost(existingPost)

		input := CreatePostInput{
			Title:      "Test Post",
			Content:    "Content",
			CategoryId: categoryId,
		}

		post, err := service.Create(ctx, input, creatorId)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if post.Slug == "test-post" {
			t.Error("Create() should append unique suffix when slug exists")
		}

		if !strings.HasPrefix(post.Slug, "test-post-") {
			t.Errorf("Create() Slug = %v, should start with test-post-", post.Slug)
		}
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	creatorId := uuid.New()
	categoryId := uuid.New()

	t.Run("update post title changes slug", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		existingPost := posts.Post{
			Id:            uuid.New(),
			Title:         "Original Title",
			Slug:          "original-title",
			Content:       "Content",
			CategoryId:    categoryId,
			CreatorUserId: creatorId,
			Status:        posts.PostStatusCreated,
		}
		repo.addPost(existingPost)

		input := UpdatePostInput{
			Title:      "New Title",
			Content:    "Updated content",
			CategoryId: categoryId,
			Status:     posts.PostStatusCreated,
		}

		post, err := service.Update(ctx, existingPost.Id, input, creatorId.String())
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if post.Title != "New Title" {
			t.Errorf("Update() Title = %v, want New Title", post.Title)
		}

		if post.Slug != "new-title" {
			t.Errorf("Update() Slug = %v, want new-title", post.Slug)
		}
	})

	t.Run("update to published sets published_at", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		existingPost := posts.Post{
			Id:            uuid.New(),
			Title:         "Draft Post",
			Slug:          "draft-post",
			Content:       "Content",
			CategoryId:    categoryId,
			CreatorUserId: creatorId,
			Status:        posts.PostStatusDraft,
		}
		repo.addPost(existingPost)

		input := UpdatePostInput{
			Title:      "Draft Post",
			Content:    "Content",
			CategoryId: categoryId,
			Status:     posts.PostStatusPublished,
		}

		post, err := service.Update(ctx, existingPost.Id, input, creatorId.String())
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if !post.PublishedAt.Valid {
			t.Error("Update() PublishedAt should be set when publishing")
		}
	})

	t.Run("update already published keeps original published_at", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		originalPublishedAt := time.Now().UTC().Add(-24 * time.Hour)
		existingPost := posts.Post{
			Id:            uuid.New(),
			Title:         "Published Post",
			Slug:          "published-post",
			Content:       "Content",
			CategoryId:    categoryId,
			CreatorUserId: creatorId,
			Status:        posts.PostStatusPublished,
			PublishedAt:   sql.NullTime{Time: originalPublishedAt, Valid: true},
		}
		repo.addPost(existingPost)

		input := UpdatePostInput{
			Title:      "Updated Published Post",
			Content:    "Updated content",
			CategoryId: categoryId,
			Status:     posts.PostStatusPublished,
		}

		post, err := service.Update(ctx, existingPost.Id, input, creatorId.String())
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if post.PublishedAt.Time.Sub(originalPublishedAt) > time.Second {
			t.Error("Update() should keep original PublishedAt for already published posts")
		}
	})

	t.Run("update post metadata", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		existingPost := posts.Post{
			Id:            uuid.New(),
			Title:         "Hiking Post",
			Slug:          "hiking-post",
			Content:       "Content",
			CategoryId:    categoryId,
			CreatorUserId: creatorId,
			Status:        posts.PostStatusCreated,
			Metadata:      json.RawMessage(`{}`),
		}
		repo.addPost(existingPost)

		metadata := json.RawMessage(`{"gpxFileUrl":"https://example.com/new-route.gpx"}`)

		input := UpdatePostInput{
			Title:      "Hiking Post",
			Content:    "Updated content",
			CategoryId: categoryId,
			Status:     posts.PostStatusCreated,
			Metadata:   metadata,
		}

		post, err := service.Update(ctx, existingPost.Id, input, creatorId.String())
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(post.Metadata, &m); err != nil {
			t.Fatalf("Update() Metadata unmarshal error = %v", err)
		}

		if m["gpxFileUrl"] != "https://example.com/new-route.gpx" {
			t.Errorf("Update() Metadata gpxFileUrl = %v, want https://example.com/new-route.gpx", m["gpxFileUrl"])
		}
	})

	t.Run("update non-existent post returns error", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		input := UpdatePostInput{
			Title:      "Test",
			Content:    "Content",
			CategoryId: categoryId,
		}

		_, err := service.Update(ctx, uuid.New(), input, creatorId.String())
		if err == nil {
			t.Error("Update() should return error for non-existent post")
		}
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("delete existing post", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		existingPost := posts.Post{
			Id:   uuid.New(),
			Slug: "test-post",
		}
		repo.addPost(existingPost)

		err := service.Delete(ctx, existingPost.Id, "user123")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify post is marked as deleted
		_, err = service.GetById(ctx, existingPost.Id)
		if err == nil {
			t.Error("Delete() should mark post as deleted")
		}
	})

	t.Run("delete non-existent post returns error", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		err := service.Delete(ctx, uuid.New(), "user123")
		if err == nil {
			t.Error("Delete() should return error for non-existent post")
		}
	})
}

func TestGetBySlug(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing post by slug", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		existingPost := posts.Post{
			Id:    uuid.New(),
			Title: "Test Post",
			Slug:  "test-post",
		}
		repo.addPost(existingPost)

		post, err := service.GetBySlug(ctx, "test-post")
		if err != nil {
			t.Fatalf("GetBySlug() error = %v", err)
		}

		if post.Title != "Test Post" {
			t.Errorf("GetBySlug() Title = %v, want Test Post", post.Title)
		}
	})

	t.Run("get non-existent slug returns error", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		_, err := service.GetBySlug(ctx, "non-existent")
		if err == nil {
			t.Error("GetBySlug() should return error for non-existent slug")
		}
	})
}

func TestGetPublished(t *testing.T) {
	ctx := context.Background()

	repo := newMockPostRepository()
	service := NewPostService(repo)

	// Add mix of published and draft posts
	for i := 0; i < 5; i++ {
		repo.addPost(posts.Post{
			Id:     uuid.New(),
			Slug:   strings.ReplaceAll(uuid.New().String(), "-", ""),
			Status: posts.PostStatusPublished,
		})
	}
	for i := 0; i < 3; i++ {
		repo.addPost(posts.Post{
			Id:     uuid.New(),
			Slug:   strings.ReplaceAll(uuid.New().String(), "-", ""),
			Status: posts.PostStatusDraft,
		})
	}

	result, total, err := service.GetPublished(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPublished() error = %v", err)
	}

	if total != 5 {
		t.Errorf("GetPublished() total = %d, want 5", total)
	}

	if len(result) != 5 {
		t.Errorf("GetPublished() returned %d posts, want 5", len(result))
	}
}

func TestGetAll_Pagination(t *testing.T) {
	ctx := context.Background()

	repo := newMockPostRepository()
	service := NewPostService(repo)

	// Add 15 posts
	for i := 0; i < 15; i++ {
		repo.addPost(posts.Post{
			Id:     uuid.New(),
			Slug:   strings.ReplaceAll(uuid.New().String(), "-", ""),
			Status: posts.PostStatusPublished,
		})
	}

	// Page 1 with 10 items
	result, total, err := service.GetAll(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if total != 15 {
		t.Errorf("GetAll() total = %d, want 15", total)
	}

	if len(result) != 10 {
		t.Errorf("GetAll() page 1 returned %d posts, want 10", len(result))
	}

	// Page 2 with 10 items (should get 5)
	result2, _, err := service.GetAll(ctx, 2, 10)
	if err != nil {
		t.Fatalf("GetAll() page 2 error = %v", err)
	}

	if len(result2) != 5 {
		t.Errorf("GetAll() page 2 returned %d posts, want 5", len(result2))
	}
}

func TestSearchPublished(t *testing.T) {
	ctx := context.Background()

	t.Run("finds posts by title", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "Hiking in the Mountains",
			Slug:    "hiking-in-the-mountains",
			Content: "Some content",
			Status:  posts.PostStatusPublished,
		})
		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "Best Recipes for Summer",
			Slug:    "best-recipes-for-summer",
			Content: "Recipe content",
			Status:  posts.PostStatusPublished,
		})

		result, total, err := service.SearchPublished(ctx, "Hiking", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 1 {
			t.Errorf("SearchPublished() total = %d, want 1", total)
		}

		if len(result) != 1 {
			t.Fatalf("SearchPublished() returned %d posts, want 1", len(result))
		}

		if result[0].Title != "Hiking in the Mountains" {
			t.Errorf("SearchPublished() Title = %v, want 'Hiking in the Mountains'", result[0].Title)
		}
	})

	t.Run("finds posts by content", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "My Post",
			Slug:    "my-post",
			Content: "This post contains protein shake recipes",
			Status:  posts.PostStatusPublished,
		})

		result, total, err := service.SearchPublished(ctx, "protein", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 1 {
			t.Errorf("SearchPublished() total = %d, want 1", total)
		}

		if len(result) != 1 {
			t.Errorf("SearchPublished() returned %d posts, want 1", len(result))
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "YOGA for Beginners",
			Slug:    "yoga-for-beginners",
			Content: "Content about yoga",
			Status:  posts.PostStatusPublished,
		})

		result, total, err := service.SearchPublished(ctx, "yoga", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 1 {
			t.Errorf("SearchPublished() total = %d, want 1", total)
		}

		if len(result) != 1 {
			t.Errorf("SearchPublished() returned %d posts, want 1", len(result))
		}
	})

	t.Run("excludes draft posts", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "Draft Fitness Guide",
			Slug:    "draft-fitness-guide",
			Content: "Fitness content",
			Status:  posts.PostStatusDraft,
		})
		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "Published Fitness Guide",
			Slug:    "published-fitness-guide",
			Content: "Fitness content",
			Status:  posts.PostStatusPublished,
		})

		result, total, err := service.SearchPublished(ctx, "Fitness", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 1 {
			t.Errorf("SearchPublished() total = %d, want 1 (should exclude draft)", total)
		}

		if len(result) != 1 {
			t.Fatalf("SearchPublished() returned %d posts, want 1", len(result))
		}

		if result[0].Title != "Published Fitness Guide" {
			t.Errorf("SearchPublished() returned wrong post: %v", result[0].Title)
		}
	})

	t.Run("excludes deleted posts", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:        uuid.New(),
			Title:     "Deleted Running Tips",
			Slug:      "deleted-running-tips",
			Content:   "Running content",
			Status:    posts.PostStatusPublished,
			IsDeleted: true,
		})

		result, total, err := service.SearchPublished(ctx, "Running", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 0 {
			t.Errorf("SearchPublished() total = %d, want 0 (should exclude deleted)", total)
		}

		if len(result) != 0 {
			t.Errorf("SearchPublished() returned %d posts, want 0", len(result))
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		repo.addPost(posts.Post{
			Id:      uuid.New(),
			Title:   "Hiking Guide",
			Slug:    "hiking-guide",
			Content: "Mountain content",
			Status:  posts.PostStatusPublished,
		})

		result, total, err := service.SearchPublished(ctx, "nonexistentterm", 1, 10)
		if err != nil {
			t.Fatalf("SearchPublished() error = %v", err)
		}

		if total != 0 {
			t.Errorf("SearchPublished() total = %d, want 0", total)
		}

		if len(result) != 0 {
			t.Errorf("SearchPublished() returned %d posts, want 0", len(result))
		}
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		repo := newMockPostRepository()
		service := NewPostService(repo)

		for i := 0; i < 5; i++ {
			repo.addPost(posts.Post{
				Id:      uuid.New(),
				Title:   fmt.Sprintf("Workout Plan %d", i),
				Slug:    fmt.Sprintf("workout-plan-%d", i),
				Content: "Workout content",
				Status:  posts.PostStatusPublished,
			})
		}

		// Page 1, 2 per page
		result, total, err := service.SearchPublished(ctx, "Workout", 1, 2)
		if err != nil {
			t.Fatalf("SearchPublished() page 1 error = %v", err)
		}

		if total != 5 {
			t.Errorf("SearchPublished() total = %d, want 5", total)
		}

		if len(result) != 2 {
			t.Errorf("SearchPublished() page 1 returned %d posts, want 2", len(result))
		}

		// Page 3, 2 per page (should get 1)
		result2, _, err := service.SearchPublished(ctx, "Workout", 3, 2)
		if err != nil {
			t.Fatalf("SearchPublished() page 3 error = %v", err)
		}

		if len(result2) != 1 {
			t.Errorf("SearchPublished() page 3 returned %d posts, want 1", len(result2))
		}
	})
}

func TestTransliterate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"english only", "hello world", "hello world"},
		{"cyrillic basic", "привет", "privet"},
		{"cyrillic Ж", "журнал", "zhurnal"},
		{"cyrillic Ц", "цена", "tsena"},
		{"cyrillic Ч", "час", "chas"},
		{"cyrillic Ш", "шапка", "shapka"},
		{"cyrillic Щ", "щастие", "shtastie"},
		{"cyrillic Ъ", "ъгъл", "agal"},
		{"cyrillic Ь", "вьюга", "vyuga"},
		{"cyrillic Ю", "юла", "yula"},
		{"cyrillic Я", "ябълка", "yabalka"},
		{"mixed case cyrillic", "ПРИВЕТ Мир", "PRIVET Mir"},
		{"numbers preserved", "тест123", "test123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transliterate(tt.input)
			if got != tt.expected {
				t.Errorf("transliterate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
