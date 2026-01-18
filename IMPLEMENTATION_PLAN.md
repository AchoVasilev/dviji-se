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

---

# Roadmap

## Milestone 1: Search & Discovery

### 1.1 Full-Text Search

**Goal:** Allow users to search posts by title, content, and excerpt.

**Files to create/modify:**
```
app/cmd/db/migrations/00003_search_index.up.sql
app/cmd/db/migrations/00003_search_index.down.sql
app/internal/domain/posts/postRepository.go (add Search method)
app/internal/application/posts/postService.go (add Search method)
app/internal/http/handlers/blogHandler.go (add SearchPosts handler)
app/internal/http/routes/blogRoutes.go (add route)
app/web/templates/blog-search.templ
```

**Database Migration:**
```sql
-- Add tsvector column for search
ALTER TABLE posts ADD COLUMN search_vector tsvector;

-- Function to update search vector (supports Bulgarian via 'simple' config)
CREATE OR REPLACE FUNCTION posts_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(NEW.excerpt, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for automatic updates
CREATE TRIGGER posts_search_vector_trigger
  BEFORE INSERT OR UPDATE ON posts
  FOR EACH ROW EXECUTE FUNCTION posts_search_vector_update();

-- GIN index for fast search
CREATE INDEX idx_posts_search ON posts USING GIN(search_vector);

-- Backfill existing posts
UPDATE posts SET title = title WHERE true;
```

**Repository Method:**
```go
func (r *PostRepository) Search(ctx context.Context, query string, limit, offset int) ([]PostWithAuthor, int, error) {
    countSQL := `
        SELECT COUNT(*) FROM posts
        WHERE status = 'published' AND is_deleted = false
          AND search_vector @@ plainto_tsquery('simple', $1)`

    dataSQL := `
        SELECT p.*, u.email, u.first_name, u.last_name,
               c.name as category_name, c.slug as category_slug,
               ts_rank(search_vector, plainto_tsquery('simple', $1)) as rank
        FROM posts p
        JOIN users u ON p.creator_user_id = u.id
        JOIN categories c ON p.category_id = c.id
        WHERE p.status = 'published' AND p.is_deleted = false
          AND search_vector @@ plainto_tsquery('simple', $1)
        ORDER BY rank DESC, published_at DESC
        LIMIT $2 OFFSET $3`

    // Execute queries...
}
```

**Route:** `GET /blog/search?q={query}&page=1`

**Template:** Search input in header, results page with pagination, "no results" state.

---

### 1.2 RSS Feed

**Goal:** Provide RSS feed for feed readers and syndication.

**Files to create:**
```
app/internal/http/handlers/feedHandler.go
app/internal/http/routes/feedRoutes.go
```

**Handler:**
```go
type FeedHandler struct {
    postService *posts.PostService
}

func (h *FeedHandler) GetRSSFeed(w http.ResponseWriter, r *http.Request) {
    posts, _, _ := h.postService.GetPublished(r.Context(), 1, 20)

    w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")

    feed := &RSS{
        Version: "2.0",
        Channel: Channel{
            Title:       "Движи се - Фитнес блог",
            Link:        config.BaseURL(),
            Description: "Фитнес съвети, рецепти и тренировки",
            Language:    "bg",
            Items:       make([]Item, len(posts)),
        },
    }

    for i, post := range posts {
        feed.Channel.Items[i] = Item{
            Title:       post.Title,
            Link:        fmt.Sprintf("%s/blog/%s", config.BaseURL(), post.Slug),
            Description: post.Excerpt,
            PubDate:     post.PublishedAt.Format(time.RFC1123Z),
            GUID:        fmt.Sprintf("%s/blog/%s", config.BaseURL(), post.Slug),
        }
    }

    xml.NewEncoder(w).Encode(feed)
}
```

**Route:** `GET /feed.xml`

**Add to base.templ:** `<link rel="alternate" type="application/rss+xml" title="RSS" href="/feed.xml">`

---

### 1.3 XML Sitemap

**Goal:** Help search engines discover and index all pages.

**Files to create:**
```
app/internal/http/handlers/sitemapHandler.go
app/internal/http/routes/sitemapRoutes.go
```

**Handler:** Generate XML sitemap with all published posts, categories, and static pages.

**Route:** `GET /sitemap.xml`

**Add robots.txt:** `Sitemap: https://yourdomain.com/sitemap.xml`

---

## Milestone 2: Content Organization

### 2.1 Tags System

**Goal:** Allow posts to have multiple tags for better organization.

**Files to create/modify:**
```
app/cmd/db/migrations/00004_tags.up.sql
app/cmd/db/migrations/00004_tags.down.sql
app/internal/domain/tags/tag.go
app/internal/domain/tags/tagRepository.go
app/internal/application/tags/tagService.go
app/internal/http/handlers/blogHandler.go (add GetBlogByTag)
app/internal/http/routes/blogRoutes.go (add route)
app/internal/http/handlers/adminHandler.go (update post form handling)
app/web/templates/admin/post-form.templ (add tag input)
app/web/templates/blog-post.templ (display tags)
app/web/templates/blog-tag.templ (new template)
```

**Database Migration:**
```sql
CREATE TABLE tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) NOT NULL,
  slug VARCHAR(50) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE posts_tags (
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (post_id, tag_id)
);

CREATE INDEX idx_posts_tags_tag ON posts_tags(tag_id);
```

**Tag Model:**
```go
type Tag struct {
    Id        uuid.UUID `db:"id"`
    Name      string    `db:"name"`
    Slug      string    `db:"slug"`
    CreatedAt time.Time `db:"created_at"`
}
```

**Repository Methods:**
- `FindAll(ctx) ([]Tag, error)`
- `FindBySlug(ctx, slug) (*Tag, error)`
- `FindByPostId(ctx, postId) ([]Tag, error)`
- `FindOrCreateByNames(ctx, names []string) ([]Tag, error)`
- `SyncPostTags(ctx, postId, tagIds []uuid.UUID) error`

**Route:** `GET /blog/tag/{slug}`

---

### 2.2 Related Posts

**Goal:** Show relevant posts instead of just recent ones.

**Files to modify:**
```
app/internal/domain/posts/postRepository.go (add FindRelated)
app/internal/application/posts/postService.go (add GetRelated)
app/web/templates/blog-post.templ (update related section)
```

**Repository Method:**
```go
func (r *PostRepository) FindRelated(ctx context.Context, post *Post, limit int) ([]PostWithAuthor, error) {
    // If tags exist, find posts with matching tags
    // Otherwise, find posts in same category
    // Exclude current post, order by relevance + date
    sql := `
        WITH post_tag_ids AS (
            SELECT tag_id FROM posts_tags WHERE post_id = $1
        ),
        scored AS (
            SELECT p.id,
                   COUNT(pt.tag_id) as tag_score,
                   CASE WHEN p.category_id = $2 THEN 2 ELSE 0 END as cat_score
            FROM posts p
            LEFT JOIN posts_tags pt ON p.id = pt.post_id
                AND pt.tag_id IN (SELECT tag_id FROM post_tag_ids)
            WHERE p.id != $1 AND p.status = 'published' AND p.is_deleted = false
            GROUP BY p.id, p.category_id
            HAVING COUNT(pt.tag_id) > 0 OR p.category_id = $2
        )
        SELECT p.*, ... FROM posts p
        JOIN scored s ON p.id = s.id
        ORDER BY (s.tag_score + s.cat_score) DESC, p.published_at DESC
        LIMIT $3`
}
```

---

## Milestone 3: Admin Enhancements

### 3.1 Post Scheduling

**Goal:** Allow posts to be scheduled for future publication.

**Files to create/modify:**
```
app/cmd/db/migrations/00005_scheduling.up.sql
app/internal/domain/posts/post.go (add ScheduledAt field)
app/internal/domain/posts/postRepository.go (add scheduling methods)
app/internal/application/posts/postService.go (add scheduling logic)
app/internal/http/handlers/adminHandler.go (handle scheduled_at)
app/web/templates/admin/post-form.templ (add datetime picker)
app/cmd/scheduler/main.go (background worker)
```

**Database Migration:**
```sql
ALTER TABLE posts ADD COLUMN scheduled_at TIMESTAMPTZ;

-- Update status check constraint
ALTER TABLE posts DROP CONSTRAINT chk_posts_status;
ALTER TABLE posts ADD CONSTRAINT chk_posts_status
    CHECK (status IN ('created', 'draft', 'scheduled', 'published', 'archived'));

CREATE INDEX idx_posts_scheduled ON posts(scheduled_at)
    WHERE status = 'scheduled' AND is_deleted = false;
```

**Background Worker:**
```go
// cmd/scheduler/main.go or integrate into main.go
func startScheduler(postRepo *posts.PostRepository) {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for range ticker.C {
            ctx := context.Background()
            posts, _ := postRepo.FindDueForPublishing(ctx)
            for _, p := range posts {
                postRepo.Publish(ctx, p.Id)
                slog.Info("Published scheduled post", "id", p.Id)
            }
        }
    }()
}
```

---

### 3.2 View Counter

**Goal:** Track post views for analytics.

**Files to create/modify:**
```
app/cmd/db/migrations/00006_view_counter.up.sql
app/internal/domain/posts/postRepository.go (add view methods)
app/internal/http/handlers/blogHandler.go (record views)
app/web/templates/admin/dashboard.templ (show stats)
```

**Database Migration:**
```sql
ALTER TABLE posts ADD COLUMN view_count INTEGER NOT NULL DEFAULT 0;

-- Optional: detailed tracking table
CREATE TABLE post_views (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES posts(id),
  viewer_hash VARCHAR(64), -- hashed IP for deduplication
  viewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_post_views_dedup ON post_views(post_id, viewer_hash, viewed_at);
```

**Deduplication Logic:** One view per IP per post per hour.

---

### 3.3 Bulk Actions

**Goal:** Perform actions on multiple posts at once.

**Files to modify:**
```
app/internal/http/handlers/adminHandler.go (add BulkAction)
app/internal/http/routes/adminRoutes.go (add route)
app/web/templates/admin/posts-list.templ (add checkboxes, action dropdown)
```

**Handler:**
```go
func (h *AdminHandler) BulkAction(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    action := r.FormValue("action")
    ids := r.Form["ids[]"]

    userId := ctxutils.GetUser(r.Context()).Id
    for _, id := range ids {
        postId := uuid.MustParse(id)
        switch action {
        case "publish":
            h.postService.Publish(r.Context(), postId)
        case "archive":
            h.postService.Archive(r.Context(), postId)
        case "delete":
            h.postService.Delete(r.Context(), postId, userId)
        }
    }
    http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}
```

**Route:** `POST /admin/posts/bulk`

---

## Milestone 4: Security Hardening

### 4.1 Token Revocation

**Goal:** Allow invalidating tokens on logout/password change.

**Files to create/modify:**
```
app/cmd/db/migrations/00007_token_blacklist.up.sql
app/internal/domain/token/blacklist.go
app/internal/domain/token/blacklistRepository.go
app/util/securityutil/token.go (check blacklist)
app/internal/http/handlers/authHandler.go (add logout)
```

**Database Migration:**
```sql
CREATE TABLE token_blacklist (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  token_hash VARCHAR(64) NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  expires_at TIMESTAMPTZ NOT NULL,
  reason VARCHAR(20) NOT NULL, -- 'logout', 'password_change'
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_token_blacklist_hash ON token_blacklist(token_hash)
    WHERE expires_at > now();
```

**Token Validation Update:**
```go
func validateToken(tokenStr string) (*jwt.Token, error) {
    token, err := parseToken(tokenStr, config.JWTAccessKey())
    if err != nil {
        return nil, err
    }

    // Check blacklist
    hash := sha256.Sum256([]byte(tokenStr))
    if blacklistRepo.IsRevoked(ctx, hex.EncodeToString(hash[:])) {
        return nil, errors.New("token has been revoked")
    }

    return token, nil
}
```

---

### 4.2 Password Complexity

**Goal:** Enforce strong passwords.

**Files to modify:**
```
app/util/httputils/validation.go (add password validator)
app/internal/http/handlers/models/user.go (use validator)
```

**Validator:**
```go
func ValidatePasswordComplexity(password string) error {
    if len(password) < 12 {
        return errors.New("Паролата трябва да е поне 12 символа")
    }
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        return errors.New("Паролата трябва да съдържа главна буква")
    }
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        return errors.New("Паролата трябва да съдържа малка буква")
    }
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        return errors.New("Паролата трябва да съдържа цифра")
    }
    if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) {
        return errors.New("Паролата трябва да съдържа специален символ")
    }
    return nil
}
```

---

### 4.3 Audit Logging

**Goal:** Track security-relevant events.

**Files to create:**
```
app/cmd/db/migrations/00008_audit_log.up.sql
app/internal/domain/audit/audit.go
app/internal/domain/audit/auditRepository.go
app/internal/application/audit/auditService.go
```

**Database Migration:**
```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type VARCHAR(50) NOT NULL,
  user_id UUID REFERENCES users(id),
  ip_address INET,
  user_agent TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_type ON audit_logs(event_type, created_at DESC);
```

**Event Types:** `login_success`, `login_failure`, `logout`, `password_change`, `password_reset`, `admin_action`

---

## Milestone 5: SEO & Social

### 5.1 Open Graph Meta Tags

**Goal:** Rich previews when sharing on social media.

**Files to modify:**
```
app/web/templates/base.templ (add OG meta tags)
app/web/templates/blog-post.templ (pass OG data)
```

**Template Update:**
```go
templ Base(props PageProps) {
    <head>
        <meta property="og:title" content={ props.Title } />
        <meta property="og:description" content={ props.Description } />
        <meta property="og:image" content={ props.Image } />
        <meta property="og:type" content={ props.Type } />
        <meta property="og:url" content={ props.URL } />
        <meta name="twitter:card" content="summary_large_image" />
    </head>
}
```

---

### 5.2 Schema.org Markup

**Goal:** Structured data for search engines.

**Files to modify:**
```
app/web/templates/blog-post.templ (add JSON-LD)
```

**JSON-LD Template:**
```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "{{ .Title }}",
  "image": "{{ .CoverImageUrl }}",
  "datePublished": "{{ .PublishedAt }}",
  "author": { "@type": "Person", "name": "{{ .AuthorName }}" }
}
</script>
```

---

## Implementation Priority

| Priority | Milestone | Effort | Impact |
|----------|-----------|--------|--------|
| 1 | Search | Medium | High |
| 2 | RSS Feed | Low | Medium |
| 3 | Tags | Medium | High |
| 4 | View Counter | Low | Medium |
| 5 | Post Scheduling | Medium | Medium |
| 6 | Sitemap | Low | Medium |
| 7 | OG Tags | Low | Medium |
| 8 | Token Revocation | Medium | High |
| 9 | Bulk Actions | Low | Low |
| 10 | Audit Logging | Medium | Medium |
