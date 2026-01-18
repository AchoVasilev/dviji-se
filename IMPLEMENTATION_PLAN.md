# Fitness Blog Implementation Plan

## Status: ✅ COMPLETE

All phases have been implemented.

---

## Overview

Transform the application into a fitness blog with admin panel for content management and public blog pages for readers.

---

## Phase 1: Database Schema ✅

**File:** `app/cmd/db/migrations/00001_init.up.sql`

### Posts Table
```sql
CREATE TABLE posts
(
  id UUID NOT NULL,
  title VARCHAR(100) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  content VARCHAR NOT NULL,
  excerpt VARCHAR(500),
  cover_image_url VARCHAR(500),
  status VARCHAR(20) NOT NULL DEFAULT 'created',
  published_at TIMESTAMPTZ,
  meta_description VARCHAR(160),
  reading_time_minutes INTEGER DEFAULT 0,
  category_id UUID NOT NULL,
  creator_user_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  updated_by VARCHAR,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_post_id PRIMARY KEY(id),
  CONSTRAINT fk_category_id FOREIGN KEY(category_id) REFERENCES categories(id),
  CONSTRAINT uq_posts_slug UNIQUE (slug),
  CONSTRAINT chk_posts_status CHECK (status IN ('created', 'draft', 'published', 'archived'))
);

CREATE INDEX idx_posts_status ON posts (status, published_at DESC) WHERE is_deleted = FALSE;
```

### Categories Table
```sql
CREATE TABLE categories
(
  id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) NOT NULL,
  image_url VARCHAR NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT(now() at time zone 'utc'),
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,

  CONSTRAINT pk_category_id PRIMARY KEY(id),
  CONSTRAINT uq_categories_slug UNIQUE (slug)
);
```

---

## Phase 2: Domain Layer ✅

### Post Model
**File:** `app/internal/domain/posts/post.go`

Fields: `Id`, `Title`, `Slug`, `Content`, `Excerpt`, `CoverImageUrl`, `Status`, `PublishedAt`, `MetaDescription`, `ReadingTimeMinutes`, `CategoryId`, `CreatorUserId`, timestamps

Includes `PostWithAuthor` struct for queries joining user/category data.

### Post Repository
**File:** `app/internal/domain/posts/postRepository.go`

Methods:
- `Create(ctx, post) (*Post, error)`
- `Update(ctx, post) (*Post, error)`
- `Delete(ctx, id, deletedBy) error`
- `FindById(ctx, id) (*Post, error)`
- `FindBySlug(ctx, slug) (*PostWithAuthor, error)`
- `FindPublished(ctx, limit, offset) ([]PostWithAuthor, total, error)`
- `FindByCategory(ctx, categorySlug, limit, offset) ([]PostWithAuthor, total, error)`
- `FindAll(ctx, limit, offset) ([]PostWithAuthor, total, error)`
- `FindRecent(ctx, limit) ([]PostWithAuthor, error)`
- `ExistsBySlug(ctx, slug, excludeId) (bool, error)`

### Category
**File:** `app/internal/domain/category/category.go` - Includes `Slug` field
**File:** `app/internal/domain/category/categoryRepository.go` - Includes `FindBySlug` method

---

## Phase 3: Cloudinary Service ✅

**File:** `app/internal/infrastructure/cloudinary/cloudinaryService.go`

```go
type CloudinaryService struct { client *cloudinary.Cloudinary }
func NewCloudinaryService() (*CloudinaryService, error)
func (s *CloudinaryService) Upload(ctx, file, filename) (*UploadResult, error)
func (s *CloudinaryService) Delete(ctx, publicId) error
```

**Configuration:** Uses centralized config package
- `config.CloudinaryCloudName()`
- `config.CloudinaryAPIKey()`
- `config.CloudinaryAPISecret()`
- `config.CloudinaryFolder()`

---

## Phase 4: Post Service ✅

**File:** `app/internal/application/posts/postService.go`

```go
type PostService struct { postRepository PostRepository }
func NewPostService(repo) *PostService
func (s *PostService) Create(ctx, input, creatorId) (*Post, error)
func (s *PostService) Update(ctx, id, input, updatedBy) (*Post, error)
func (s *PostService) Delete(ctx, id, deletedBy) error
func (s *PostService) GetBySlug(ctx, slug) (*PostWithAuthor, error)
func (s *PostService) GetPublished(ctx, page, pageSize) ([]PostWithAuthor, total, error)
func (s *PostService) GetByCategory(ctx, categorySlug, page, pageSize) ([]PostWithAuthor, total, error)
func (s *PostService) GetAll(ctx, page, pageSize) ([]PostWithAuthor, total, error)
func (s *PostService) GetRecent(ctx, limit) ([]PostWithAuthor, error)
func (s *PostService) GenerateSlug(title) string  // Handles Cyrillic transliteration
func (s *PostService) CalculateReadingTime(content) int  // ~200 words/min
```

**Tests:** `app/internal/application/posts/postService_test.go`

---

## Phase 5: Admin Middleware ✅

**File:** `app/internal/http/middleware/admin.go`

```go
func RequireAdmin(next http.Handler) http.Handler
func RequireAuth(next http.Handler) http.Handler
```

**Tests:** `app/internal/http/middleware/admin_test.go`

---

## Phase 6: Handler Models ✅

**File:** `app/internal/http/handlers/models/post.go`

- `CreatePostResource` - Input for creating posts
- `UpdatePostResource` - Input for updating posts
- `PostResponseResource` - Full post response with author/category
- `PostListItem` - Simplified for list views
- `PaginatedResponse[T]` - Generic pagination wrapper

**File:** `app/internal/http/handlers/models/category.go`
- Category DTOs

---

## Phase 7: Admin Handler ✅

**File:** `app/internal/http/handlers/adminHandler.go`

```go
type AdminHandler struct {
    postService       *posts.PostService
    categoryService   *categories.CategoryService
    cloudinaryService *cloudinary.CloudinaryService
}

func (h *AdminHandler) GetDashboard(w, r)   // GET /admin
func (h *AdminHandler) GetPosts(w, r)       // GET /admin/posts
func (h *AdminHandler) GetPostForm(w, r)    // GET /admin/posts/new, GET /admin/posts/{id}
func (h *AdminHandler) CreatePost(w, r)     // POST /admin/posts
func (h *AdminHandler) UpdatePost(w, r)     // PUT /admin/posts/{id}
func (h *AdminHandler) DeletePost(w, r)     // DELETE /admin/posts/{id}
func (h *AdminHandler) UploadImage(w, r)    // POST /admin/upload
```

---

## Phase 8: Blog Handler ✅

**File:** `app/internal/http/handlers/blogHandler.go`

```go
type BlogHandler struct {
    postService     *posts.PostService
    categoryService *categories.CategoryService
}

func (h *BlogHandler) GetBlogList(w, r)       // GET /blog
func (h *BlogHandler) GetBlogPost(w, r)       // GET /blog/{slug}
func (h *BlogHandler) GetBlogByCategory(w, r) // GET /blog/category/{slug}
```

---

## Phase 9: Routes ✅

**File:** `app/internal/http/routes/blogRoutes.go`
- `GET /blog` → BlogHandler.GetBlogList
- `GET /blog/{slug}` → BlogHandler.GetBlogPost
- `GET /blog/category/{slug}` → BlogHandler.GetBlogByCategory

**File:** `app/internal/http/routes/adminRoutes.go`
- All `/admin/*` routes wrapped with RequireAuth + RequireAdmin middleware
- `GET /admin` → AdminHandler.GetDashboard
- `GET /admin/posts` → AdminHandler.GetPosts
- `GET /admin/posts/new` → AdminHandler.GetPostForm
- `GET /admin/posts/{id}` → AdminHandler.GetPostForm (edit)
- `POST /admin/posts` → AdminHandler.CreatePost
- `PUT /admin/posts/{id}` → AdminHandler.UpdatePost
- `DELETE /admin/posts/{id}` → AdminHandler.DeletePost
- `POST /admin/upload` → AdminHandler.UploadImage

**File:** `app/internal/http/routes/registry.go`
- Includes `BlogRoutes(mux, db)` and `AdminRoutes(mux, db)` calls

---

## Phase 10: Templates ✅

### Admin Templates
**Directory:** `app/web/templates/admin/`

- `dashboard.templ` - Stats, recent posts, quick actions
- `posts-list.templ` - Table with all posts, status badges, edit/delete actions
- `post-form.templ` - Create/edit form with TinyMCE editor integration

### Blog Templates
**Directory:** `app/web/templates/`

- `blog-list.templ` - Grid of post cards with sidebar categories, pagination
- `blog-post.templ` - Full post view with cover image, content, related posts

### Navigation
- `base.templ` - Includes "Админ" link in nav (visible only for ADMIN role)

---

## Phase 11: TinyMCE Integration ✅

In `admin/post-form.templ`:
- TinyMCE loaded from CDN
- Image upload handler POSTs to `/admin/upload`
- Dark theme to match admin UI

---

## Additional Implementations (Beyond Original Plan)

### Centralized Configuration ✅
**File:** `app/internal/config/config.go`

Single source of truth for all environment variables:
- Server: `Port()`, `Environment()`, `IsDevelopment()`, `IsProduction()`
- Database: `DBHost()`, `DBPort()`, `DBUser()`, `DBPassword()`, `DBName()`, `DBSSLMode()`, `DBMaxConns()`
- JWT: `JWTAccessKey()`, `JWTRefreshKey()`
- Security: `XSRFKey()`, `CORSOrigins()`, `AllowRegistration()`
- SMTP: `SMTPHost()`, `SMTPPort()`, `SMTPUsername()`, `SMTPPassword()`, `SMTPFrom()`, `SMTPConfigured()`
- Cloudinary: `CloudinaryCloudName()`, `CloudinaryAPIKey()`, `CloudinaryAPISecret()`, `CloudinaryFolder()`, `CloudinaryConfigured()`
- App: `BaseURL()`

### Registration Control ✅
Registration routes conditionally enabled via `config.AllowRegistration()` (disabled by default)

### Testing ✅
- Unit tests: securityutil, PostService, AuthService, middleware, httputils
- Integration tests: auth API, blog API, admin API
- Test infrastructure with testcontainers

### Docker ✅
- Multi-stage Dockerfile (Tailwind → Go build → Alpine runtime)
- `.dockerignore` for optimized builds
- Health endpoint at `/health`

---

## File Summary

### Implemented Files
```
# Domain
app/internal/domain/posts/post.go
app/internal/domain/posts/postRepository.go
app/internal/domain/category/category.go
app/internal/domain/category/categoryRepository.go

# Application
app/internal/application/posts/postService.go
app/internal/application/posts/postService_test.go
app/internal/application/categories/categoryService.go

# Infrastructure
app/internal/infrastructure/cloudinary/cloudinaryService.go
app/internal/infrastructure/email/emailService.go
app/internal/config/config.go

# HTTP
app/internal/http/middleware/admin.go
app/internal/http/middleware/admin_test.go
app/internal/http/handlers/adminHandler.go
app/internal/http/handlers/blogHandler.go
app/internal/http/handlers/models/post.go
app/internal/http/handlers/models/category.go
app/internal/http/routes/adminRoutes.go
app/internal/http/routes/blogRoutes.go

# Templates
app/web/templates/admin/dashboard.templ
app/web/templates/admin/posts-list.templ
app/web/templates/admin/post-form.templ
app/web/templates/blog-list.templ
app/web/templates/blog-post.templ

# Tests
app/tests/integration/blog_api_test.go
app/tests/integration/auth_api_test.go
```

---

## Verification

### Admin Flow
1. Login as admin user
2. Navigate to `/admin`
3. Create new post with TinyMCE, upload image
4. Save as draft, then publish
5. Verify post appears on `/blog`

### Public Flow
1. Visit `/blog` as anonymous user
2. Browse posts, click category filter
3. View single post at `/blog/{slug}`
4. Check SEO meta tags in page source

### Security
1. `/admin/*` returns 403 for non-admin users
2. Draft posts don't appear on public `/blog`
3. Image upload requires admin auth
4. Registration disabled by default (`ALLOW_REGISTRATION=false`)
