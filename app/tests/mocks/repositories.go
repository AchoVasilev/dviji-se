package mocks

import (
	"context"
	"database/sql"
	"errors"
	"server/internal/domain/category"
	"server/internal/domain/posts"
	"server/internal/domain/user"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

// MockUserRepository is an in-memory mock of the user repository
type MockUserRepository struct {
	mu          sync.RWMutex
	users       map[uuid.UUID]user.User
	usersByEmail map[string]uuid.UUID
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:       make(map[uuid.UUID]user.User),
		usersByEmail: make(map[string]uuid.UUID),
	}
}

func (r *MockUserRepository) Create(u user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByEmail[u.Email]; exists {
		return ErrExists
	}

	r.users[u.Id] = u
	r.usersByEmail[u.Email] = u.Id
	return nil
}

func (r *MockUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usersByEmail[email]
	if !exists {
		return user.User{}, sql.ErrNoRows
	}

	u, exists := r.users[id]
	if !exists || u.IsDeleted {
		return user.User{}, sql.ErrNoRows
	}

	return u, nil
}

func (r *MockUserRepository) FindById(ctx context.Context, userId string) (user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, err := uuid.Parse(userId)
	if err != nil {
		return user.User{}, err
	}

	u, exists := r.users[id]
	if !exists || u.IsDeleted {
		return user.User{}, sql.ErrNoRows
	}

	return u, nil
}

func (r *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usersByEmail[email]
	if !exists {
		return false, nil
	}

	u, exists := r.users[id]
	return exists && !u.IsDeleted, nil
}

func (r *MockUserRepository) UpdatePassword(ctx context.Context, userId string, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := uuid.Parse(userId)
	if err != nil {
		return err
	}

	u, exists := r.users[id]
	if !exists {
		return sql.ErrNoRows
	}

	u.Password = hashedPassword
	u.UpdatedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	r.users[id] = u
	return nil
}

// AddUser is a helper for tests to add users directly
func (r *MockUserRepository) AddUser(u user.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Id] = u
	r.usersByEmail[u.Email] = u.Id
}

// MockPasswordResetTokenRepository is an in-memory mock
type MockPasswordResetTokenRepository struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]user.PasswordResetToken
	byHash map[string]uuid.UUID
}

func NewMockPasswordResetTokenRepository() *MockPasswordResetTokenRepository {
	return &MockPasswordResetTokenRepository{
		tokens: make(map[uuid.UUID]user.PasswordResetToken),
		byHash: make(map[string]uuid.UUID),
	}
}

func (r *MockPasswordResetTokenRepository) Create(ctx context.Context, userId uuid.UUID, tokenHash string) (*user.PasswordResetToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token := &user.PasswordResetToken{
		Id:        uuid.New(),
		UserId:    userId,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(user.TokenExpirationDuration),
		CreatedAt: time.Now().UTC(),
	}

	r.tokens[token.Id] = *token
	r.byHash[tokenHash] = token.Id
	return token, nil
}

func (r *MockPasswordResetTokenRepository) FindValidByHash(ctx context.Context, tokenHash string) (*user.PasswordResetToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byHash[tokenHash]
	if !exists {
		return nil, sql.ErrNoRows
	}

	token, exists := r.tokens[id]
	if !exists {
		return nil, sql.ErrNoRows
	}

	if token.UsedAt.Valid || token.ExpiresAt.Before(time.Now()) {
		return nil, sql.ErrNoRows
	}

	return &token, nil
}

func (r *MockPasswordResetTokenRepository) MarkAsUsed(ctx context.Context, tokenId uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, exists := r.tokens[tokenId]
	if !exists {
		return sql.ErrNoRows
	}

	token.UsedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	r.tokens[tokenId] = token
	return nil
}

func (r *MockPasswordResetTokenRepository) InvalidateAllForUser(ctx context.Context, userId uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, token := range r.tokens {
		if token.UserId == userId && !token.UsedAt.Valid {
			token.UsedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
			r.tokens[id] = token
		}
	}
	return nil
}

func (r *MockPasswordResetTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var count int64
	for id, token := range r.tokens {
		if token.ExpiresAt.Before(time.Now()) {
			delete(r.tokens, id)
			delete(r.byHash, token.TokenHash)
			count++
		}
	}
	return count, nil
}

// MockCategoryRepository is an in-memory mock
type MockCategoryRepository struct {
	mu         sync.RWMutex
	categories map[uuid.UUID]category.Category
	bySlug     map[string]uuid.UUID
}

func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		categories: make(map[uuid.UUID]category.Category),
		bySlug:     make(map[string]uuid.UUID),
	}
}

func (r *MockCategoryRepository) FindAll(ctx context.Context) ([]category.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []category.Category
	for _, c := range r.categories {
		if !c.IsDeleted {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *MockCategoryRepository) FindBySlug(ctx context.Context, slug string) (*category.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.bySlug[slug]
	if !exists {
		return nil, sql.ErrNoRows
	}

	c, exists := r.categories[id]
	if !exists || c.IsDeleted {
		return nil, sql.ErrNoRows
	}

	return &c, nil
}

func (r *MockCategoryRepository) Create(ctx context.Context, c *category.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.categories[c.Id] = *c
	r.bySlug[c.Slug] = c.Id
	return nil
}

// AddCategory is a helper for tests
func (r *MockCategoryRepository) AddCategory(c category.Category) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[c.Id] = c
	r.bySlug[c.Slug] = c.Id
}

// MockPostRepository is an in-memory mock
type MockPostRepository struct {
	mu      sync.RWMutex
	posts   map[uuid.UUID]posts.Post
	bySlug  map[string]uuid.UUID
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts:  make(map[uuid.UUID]posts.Post),
		bySlug: make(map[string]uuid.UUID),
	}
}

func (r *MockPostRepository) Create(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bySlug[p.Slug]; exists {
		return nil, ErrExists
	}

	r.posts[p.Id] = *p
	r.bySlug[p.Slug] = p.Id
	return p, nil
}

func (r *MockPostRepository) Update(ctx context.Context, p *posts.Post) (*posts.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[p.Id]; !exists {
		return nil, sql.ErrNoRows
	}

	r.posts[p.Id] = *p
	r.bySlug[p.Slug] = p.Id
	return p, nil
}

func (r *MockPostRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy string) error {
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

func (r *MockPostRepository) FindById(ctx context.Context, id uuid.UUID) (*posts.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.posts[id]
	if !exists || p.IsDeleted {
		return nil, sql.ErrNoRows
	}

	return &p, nil
}

func (r *MockPostRepository) FindBySlug(ctx context.Context, slug string) (*posts.PostWithAuthor, error) {
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

func (r *MockPostRepository) FindPublished(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error) {
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

func (r *MockPostRepository) FindAll(ctx context.Context, limit, offset int) ([]posts.PostWithAuthor, int, error) {
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

func (r *MockPostRepository) FindRecent(ctx context.Context, limit int) ([]posts.PostWithAuthor, error) {
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

func (r *MockPostRepository) ExistsBySlug(ctx context.Context, slug string, excludeId *uuid.UUID) (bool, error) {
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

func (r *MockPostRepository) FindByCategory(ctx context.Context, categorySlug string, limit, offset int) ([]posts.PostWithAuthor, int, error) {
	return []posts.PostWithAuthor{}, 0, nil
}

func (r *MockPostRepository) FindByStatus(ctx context.Context, status posts.PostStatus, limit, offset int) ([]posts.PostWithAuthor, int, error) {
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

// AddPost is a helper for tests
func (r *MockPostRepository) AddPost(p posts.Post) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.posts[p.Id] = p
	r.bySlug[p.Slug] = p.Id
}

// MockEmailService is a mock for email sending
type MockEmailService struct {
	mu         sync.Mutex
	SentEmails []SentEmail
}

type SentEmail struct {
	To    string
	Token string
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]SentEmail, 0),
	}
}

func (s *MockEmailService) SendPasswordResetEmail(toEmail, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SentEmails = append(s.SentEmails, SentEmail{To: toEmail, Token: token})
	return nil
}

func (s *MockEmailService) GetSentEmails() []SentEmail {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.SentEmails
}

func (s *MockEmailService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SentEmails = make([]SentEmail, 0)
}
