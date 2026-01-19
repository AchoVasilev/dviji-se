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

## Milestone 6: Monetization & Ads

### 6.1 Ad Consent System

**Goal:** User-friendly, consent-based ad display with persistent widget.

**Behavior:**
1. On first visit, a popup asks: "Искате ли да гледате реклами?" (Do you want to watch ads?)
2. User clicks "Да" (Yes) or "Не" (No)
3. Regardless of choice, the popup minimizes to a small widget in the bottom-right corner
4. Ads ONLY display if user consented
5. User can change their preference anytime by clicking the minimized widget

**Files to create/modify:**
```
app/web/templates/components/ad-consent.templ (NEW - popup + widget)
app/web/templates/components/ad-unit.templ (NEW - conditional ad display)
app/web/templates/base.templ (include consent component)
app/web/static/js/ad-consent.js (NEW - consent logic)
```

**Consent Popup Component:**
```go
templ AdConsentPopup() {
    <!-- Full popup (shown on first visit) -->
    <div id="ad-consent-popup" class="hidden fixed inset-0 bg-black/50 z-50 flex items-center justify-center">
        <div class="bg-white rounded-lg shadow-xl p-6 max-w-sm mx-4">
            <h3 class="text-lg font-semibold mb-3">Подкрепете ни</h3>
            <p class="text-gray-600 mb-4">
                Искате ли да гледате реклами? Това ни помага да поддържаме сайта безплатен.
            </p>
            <div class="flex gap-3">
                <button onclick="setAdConsent(true)"
                        class="flex-1 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                    Да, съгласен съм
                </button>
                <button onclick="setAdConsent(false)"
                        class="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300">
                    Не, благодаря
                </button>
            </div>
        </div>
    </div>

    <!-- Minimized widget (always visible after first interaction) -->
    <div id="ad-consent-widget" class="hidden fixed bottom-4 right-4 z-40">
        <button onclick="toggleAdConsentPopup()"
                class="bg-white shadow-lg rounded-full p-3 hover:shadow-xl transition-shadow"
                title="Промени настройките за реклами">
            <span id="ad-widget-icon" class="text-xl">📺</span>
        </button>
        <span id="ad-widget-status" class="absolute -top-1 -right-1 w-3 h-3 rounded-full"></span>
    </div>

    <!-- Mini popup for changing preference -->
    <div id="ad-consent-mini" class="hidden fixed bottom-20 right-4 z-40 bg-white rounded-lg shadow-xl p-4 w-64">
        <p class="text-sm text-gray-600 mb-3" id="ad-consent-mini-text"></p>
        <button id="ad-consent-toggle" onclick="toggleAdConsent()"
                class="w-full px-4 py-2 rounded text-sm font-medium">
        </button>
    </div>
}
```

**JavaScript Logic (`ad-consent.js`):**
```javascript
const AD_CONSENT_KEY = 'ad_consent';
const AD_CONSENT_SHOWN_KEY = 'ad_consent_shown';

function initAdConsent() {
    const hasSeenPopup = localStorage.getItem(AD_CONSENT_SHOWN_KEY);

    if (!hasSeenPopup) {
        // First visit - show full popup
        document.getElementById('ad-consent-popup').classList.remove('hidden');
    } else {
        // Returning visitor - show minimized widget
        showWidget();
        if (hasAdConsent()) {
            loadAds();
        }
    }
}

function setAdConsent(consented) {
    localStorage.setItem(AD_CONSENT_KEY, consented ? 'true' : 'false');
    localStorage.setItem(AD_CONSENT_SHOWN_KEY, 'true');

    // Hide popup, show widget
    document.getElementById('ad-consent-popup').classList.add('hidden');
    showWidget();

    if (consented) {
        loadAds();
    }
}

function hasAdConsent() {
    return localStorage.getItem(AD_CONSENT_KEY) === 'true';
}

function showWidget() {
    const widget = document.getElementById('ad-consent-widget');
    const status = document.getElementById('ad-widget-status');

    widget.classList.remove('hidden');

    if (hasAdConsent()) {
        status.classList.add('bg-green-500');
        status.classList.remove('bg-gray-400');
    } else {
        status.classList.add('bg-gray-400');
        status.classList.remove('bg-green-500');
    }
}

function toggleAdConsentPopup() {
    const mini = document.getElementById('ad-consent-mini');
    const text = document.getElementById('ad-consent-mini-text');
    const btn = document.getElementById('ad-consent-toggle');

    if (mini.classList.contains('hidden')) {
        if (hasAdConsent()) {
            text.textContent = 'В момента гледате реклами.';
            btn.textContent = 'Изключи рекламите';
            btn.className = 'w-full px-4 py-2 rounded text-sm font-medium bg-gray-200 text-gray-700 hover:bg-gray-300';
        } else {
            text.textContent = 'Рекламите са изключени.';
            btn.textContent = 'Включи рекламите';
            btn.className = 'w-full px-4 py-2 rounded text-sm font-medium bg-green-600 text-white hover:bg-green-700';
        }
        mini.classList.remove('hidden');
    } else {
        mini.classList.add('hidden');
    }
}

function toggleAdConsent() {
    const newConsent = !hasAdConsent();
    localStorage.setItem(AD_CONSENT_KEY, newConsent ? 'true' : 'false');

    document.getElementById('ad-consent-mini').classList.add('hidden');
    showWidget();

    if (newConsent) {
        loadAds();
    } else {
        hideAds();
    }
}

function loadAds() {
    // Show all ad containers
    document.querySelectorAll('.ad-container').forEach(el => {
        el.classList.remove('hidden');
    });

    // Initialize AdSense if using third-party
    if (typeof adsbygoogle !== 'undefined') {
        document.querySelectorAll('.adsbygoogle').forEach(() => {
            (adsbygoogle = window.adsbygoogle || []).push({});
        });
    }
}

function hideAds() {
    document.querySelectorAll('.ad-container').forEach(el => {
        el.classList.add('hidden');
    });
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initAdConsent);
```

**Conditional Ad Unit Component:**
```go
// Ad containers are hidden by default, shown only if consent given
templ AdUnit(slot string, placement string) {
    <div class="ad-container hidden" data-placement={ placement }>
        <!-- Third-party ad (Google AdSense) -->
        <ins class="adsbygoogle"
             style="display:block"
             data-ad-client={ config.GoogleAdsClientId() }
             data-ad-slot={ slot }
             data-ad-format="auto"
             data-full-width-responsive="true"></ins>
    </div>
}

// For self-hosted ads
templ SelfHostedAdUnit(ad *ads.Ad) {
    if ad != nil {
        <div class="ad-container hidden" data-placement={ ad.Placement }>
            <a href={ templ.URL("/ads/click/" + ad.Id.String()) }
               target="_blank" rel="sponsored noopener">
                <img src={ ad.ImageUrl } alt={ ad.Name } class="w-full rounded" />
            </a>
            <span class="text-xs text-gray-500">Реклама</span>
        </div>
    }
}
```

**Configuration:**
```go
func GoogleAdsClientId() string { return os.Getenv("GOOGLE_ADS_CLIENT_ID") }
func AdsEnabled() bool { return os.Getenv("ADS_ENABLED") == "true" }
```

**Placements:**
- Sidebar (blog list)
- In-content (after 3rd paragraph)
- Footer (blog post)

---

### 6.2 Self-Hosted Ads

**Goal:** Full control over ad inventory, direct sponsorships, no revenue share.

**Files to create:**
```
app/cmd/db/migrations/00009_ads.up.sql
app/cmd/db/migrations/00009_ads.down.sql
app/internal/domain/ads/ad.go
app/internal/domain/ads/adRepository.go
app/internal/application/ads/adService.go
app/internal/http/handlers/adminAdsHandler.go
app/internal/http/routes/adminRoutes.go (add ad routes)
app/web/templates/admin/ads-list.templ
app/web/templates/admin/ad-form.templ
app/web/templates/components/ad-unit.templ
```

**Database Migration:**
```sql
CREATE TABLE ads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  image_url VARCHAR(500) NOT NULL,
  target_url VARCHAR(500) NOT NULL,
  placement VARCHAR(50) NOT NULL, -- 'sidebar', 'in-content', 'header', 'footer'
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  start_date TIMESTAMPTZ,
  end_date TIMESTAMPTZ,
  impressions INTEGER NOT NULL DEFAULT 0,
  clicks INTEGER NOT NULL DEFAULT 0,
  priority INTEGER NOT NULL DEFAULT 0, -- higher = shown first
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ
);

CREATE INDEX idx_ads_active ON ads(placement, is_active, priority DESC)
    WHERE is_active = true
    AND (start_date IS NULL OR start_date <= now())
    AND (end_date IS NULL OR end_date >= now());
```

**Ad Model:**
```go
type Ad struct {
    Id          uuid.UUID  `db:"id"`
    Name        string     `db:"name"`
    ImageUrl    string     `db:"image_url"`
    TargetUrl   string     `db:"target_url"`
    Placement   string     `db:"placement"` // sidebar, in-content, header, footer
    IsActive    bool       `db:"is_active"`
    StartDate   *time.Time `db:"start_date"`
    EndDate     *time.Time `db:"end_date"`
    Impressions int        `db:"impressions"`
    Clicks      int        `db:"clicks"`
    Priority    int        `db:"priority"`
    CreatedAt   time.Time  `db:"created_at"`
    UpdatedAt   *time.Time `db:"updated_at"`
}

type AdPlacement string

const (
    PlacementSidebar   AdPlacement = "sidebar"
    PlacementInContent AdPlacement = "in-content"
    PlacementHeader    AdPlacement = "header"
    PlacementFooter    AdPlacement = "footer"
)
```

**Repository Methods:**
```go
func (r *AdRepository) FindActiveByPlacement(ctx context.Context, placement AdPlacement) (*Ad, error) {
    sql := `
        SELECT * FROM ads
        WHERE placement = $1 AND is_active = true
          AND (start_date IS NULL OR start_date <= now())
          AND (end_date IS NULL OR end_date >= now())
        ORDER BY priority DESC, RANDOM()
        LIMIT 1`
    // Returns one ad, with highest priority or random among equal priority
}

func (r *AdRepository) RecordImpression(ctx context.Context, id uuid.UUID) error {
    _, err := r.db.ExecContext(ctx,
        "UPDATE ads SET impressions = impressions + 1 WHERE id = $1", id)
    return err
}

func (r *AdRepository) RecordClick(ctx context.Context, id uuid.UUID) error {
    _, err := r.db.ExecContext(ctx,
        "UPDATE ads SET clicks = clicks + 1 WHERE id = $1", id)
    return err
}
```

**Click Tracking Handler:**
```go
// GET /ads/click/{id}
func (h *AdHandler) TrackClick(w http.ResponseWriter, r *http.Request) {
    id := uuid.MustParse(r.PathValue("id"))
    ad, err := h.adRepo.FindById(r.Context(), id)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    h.adRepo.RecordClick(r.Context(), id)
    http.Redirect(w, r, ad.TargetUrl, http.StatusFound)
}
```

**Ad Component:**
```go
templ AdUnit(ad *ads.Ad) {
    if ad != nil {
        <div class="ad-unit" data-ad-id={ ad.Id.String() }>
            <a href={ templ.URL("/ads/click/" + ad.Id.String()) } target="_blank" rel="sponsored noopener">
                <img src={ ad.ImageUrl } alt={ ad.Name } class="w-full rounded" />
            </a>
            <span class="text-xs text-gray-500">Реклама</span>
        </div>
    }
}
```

**Admin Routes:**
- `GET /admin/ads` - List all ads with stats
- `GET /admin/ads/new` - Create ad form
- `GET /admin/ads/{id}` - Edit ad form
- `POST /admin/ads` - Create ad
- `PUT /admin/ads/{id}` - Update ad
- `DELETE /admin/ads/{id}` - Delete ad

---

### 6.3 Affiliate Links

**Goal:** Track affiliate product links for commission reporting.

**Files to create:**
```
app/cmd/db/migrations/00010_affiliates.up.sql
app/internal/domain/affiliates/affiliate.go
app/internal/domain/affiliates/affiliateRepository.go
app/internal/http/handlers/affiliateHandler.go
```

**Database Migration:**
```sql
CREATE TABLE affiliate_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(50) NOT NULL UNIQUE, -- short link: /go/protein-powder
  target_url VARCHAR(500) NOT NULL, -- actual affiliate URL
  partner VARCHAR(50), -- 'amazon', 'iherb', etc.
  clicks INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_affiliate_slug ON affiliate_links(slug);
```

**Redirect Handler:**
```go
// GET /go/{slug}
func (h *AffiliateHandler) Redirect(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("slug")
    link, err := h.repo.FindBySlug(r.Context(), slug)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    h.repo.RecordClick(r.Context(), link.Id)
    http.Redirect(w, r, link.TargetUrl, http.StatusFound)
}
```

**Usage in posts:** Link to `/go/protein-powder` instead of raw affiliate URLs.

---

### 6.4 Sponsored Posts

**Goal:** Mark posts as sponsored and display sponsor information.

**Files to modify:**
```
app/cmd/db/migrations/00011_sponsored_posts.up.sql
app/internal/domain/posts/post.go (add sponsor fields)
app/web/templates/admin/post-form.templ (add sponsor inputs)
app/web/templates/blog-post.templ (display sponsor badge)
app/web/templates/blog-list.templ (sponsored indicator)
```

**Database Migration:**
```sql
ALTER TABLE posts ADD COLUMN is_sponsored BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE posts ADD COLUMN sponsor_name VARCHAR(100);
ALTER TABLE posts ADD COLUMN sponsor_url VARCHAR(500);
ALTER TABLE posts ADD COLUMN sponsor_logo_url VARCHAR(500);
```

**Post Model Update:**
```go
type Post struct {
    // ... existing fields ...
    IsSponsored    bool    `db:"is_sponsored"`
    SponsorName    *string `db:"sponsor_name"`
    SponsorUrl     *string `db:"sponsor_url"`
    SponsorLogoUrl *string `db:"sponsor_logo_url"`
}
```

**Sponsored Badge Component:**
```go
templ SponsoredBadge(post *posts.Post) {
    if post.IsSponsored && post.SponsorName != nil {
        <div class="sponsored-badge bg-yellow-100 text-yellow-800 px-3 py-1 rounded-full text-sm">
            <span>Спонсорирано от </span>
            if post.SponsorUrl != nil {
                <a href={ templ.URL(*post.SponsorUrl) } target="_blank" rel="sponsored noopener" class="font-medium hover:underline">
                    { *post.SponsorName }
                </a>
            } else {
                <span class="font-medium">{ *post.SponsorName }</span>
            }
        </div>
    }
}
```

---

## Milestone 7: GDPR & Privacy Compliance

### 7.1 Cookie Consent Banner

**Goal:** Compliant cookie consent with granular control.

**Files to create/modify:**
```
app/web/templates/components/cookie-consent.templ (NEW)
app/web/static/js/cookie-consent.js (NEW)
app/web/templates/base.templ (include banner)
```

**Cookie Categories:**
- **Necessary** - Always enabled (session, CSRF, auth tokens)
- **Analytics** - Google Analytics, view tracking
- **Advertising** - Ad networks, affiliate tracking

**Cookie Banner Component:**
```go
templ CookieConsentBanner() {
    <div id="cookie-banner" class="hidden fixed bottom-0 inset-x-0 z-50 bg-white border-t shadow-lg">
        <div class="max-w-7xl mx-auto p-4">
            <div class="flex flex-col md:flex-row items-start md:items-center gap-4">
                <div class="flex-1">
                    <h3 class="font-semibold">Използваме бисквитки 🍪</h3>
                    <p class="text-sm text-gray-600">
                        Използваме бисквитки за подобряване на вашето изживяване.
                        <a href="/privacy" class="text-blue-600 hover:underline">Научете повече</a>
                    </p>
                </div>
                <div class="flex gap-2">
                    <button onclick="acceptAllCookies()"
                            class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                        Приемам всички
                    </button>
                    <button onclick="openCookieSettings()"
                            class="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300">
                        Настройки
                    </button>
                    <button onclick="acceptNecessaryCookies()"
                            class="text-gray-500 hover:text-gray-700 px-2">
                        Само необходими
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Cookie Settings Modal -->
    <div id="cookie-settings-modal" class="hidden fixed inset-0 z-50 bg-black/50 flex items-center justify-center">
        <div class="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-y-auto">
            <div class="p-6">
                <h2 class="text-xl font-semibold mb-4">Настройки за бисквитки</h2>

                <!-- Necessary -->
                <div class="border-b pb-4 mb-4">
                    <div class="flex items-center justify-between">
                        <span class="font-medium">Необходими</span>
                        <span class="text-sm text-gray-500">Винаги активни</span>
                    </div>
                    <p class="text-sm text-gray-600 mt-1">
                        Необходими за работата на сайта (сесия, сигурност).
                    </p>
                </div>

                <!-- Analytics -->
                <div class="border-b pb-4 mb-4">
                    <div class="flex items-center justify-between">
                        <span class="font-medium">Аналитични</span>
                        <label class="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" id="cookie-analytics" class="sr-only peer">
                            <div class="w-11 h-6 bg-gray-200 peer-checked:bg-green-600 rounded-full
                                        peer-focus:ring-2 peer-focus:ring-green-300
                                        after:content-[''] after:absolute after:top-[2px] after:left-[2px]
                                        after:bg-white after:rounded-full after:h-5 after:w-5
                                        after:transition-all peer-checked:after:translate-x-full"></div>
                        </label>
                    </div>
                    <p class="text-sm text-gray-600 mt-1">
                        Помагат ни да разберем как използвате сайта.
                    </p>
                </div>

                <!-- Advertising -->
                <div class="pb-4 mb-4">
                    <div class="flex items-center justify-between">
                        <span class="font-medium">Рекламни</span>
                        <label class="relative inline-flex items-center cursor-pointer">
                            <input type="checkbox" id="cookie-advertising" class="sr-only peer">
                            <div class="w-11 h-6 bg-gray-200 peer-checked:bg-green-600 rounded-full
                                        peer-focus:ring-2 peer-focus:ring-green-300
                                        after:content-[''] after:absolute after:top-[2px] after:left-[2px]
                                        after:bg-white after:rounded-full after:h-5 after:w-5
                                        after:transition-all peer-checked:after:translate-x-full"></div>
                        </label>
                    </div>
                    <p class="text-sm text-gray-600 mt-1">
                        Използват се за показване на подходящи реклами.
                    </p>
                </div>

                <div class="flex gap-2">
                    <button onclick="saveCookieSettings()"
                            class="flex-1 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                        Запази настройките
                    </button>
                    <button onclick="closeCookieSettings()"
                            class="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300">
                        Отказ
                    </button>
                </div>
            </div>
        </div>
    </div>
}
```

**JavaScript (`cookie-consent.js`):**
```javascript
const COOKIE_CONSENT_KEY = 'cookie_consent';

function initCookieConsent() {
    const consent = getCookieConsent();
    if (!consent) {
        document.getElementById('cookie-banner').classList.remove('hidden');
    } else {
        applyCookieSettings(consent);
    }
}

function getCookieConsent() {
    const stored = localStorage.getItem(COOKIE_CONSENT_KEY);
    return stored ? JSON.parse(stored) : null;
}

function setCookieConsent(consent) {
    consent.timestamp = new Date().toISOString();
    localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify(consent));
    document.getElementById('cookie-banner').classList.add('hidden');
    applyCookieSettings(consent);
}

function acceptAllCookies() {
    setCookieConsent({ necessary: true, analytics: true, advertising: true });
}

function acceptNecessaryCookies() {
    setCookieConsent({ necessary: true, analytics: false, advertising: false });
}

function openCookieSettings() {
    const consent = getCookieConsent() || { analytics: false, advertising: false };
    document.getElementById('cookie-analytics').checked = consent.analytics;
    document.getElementById('cookie-advertising').checked = consent.advertising;
    document.getElementById('cookie-settings-modal').classList.remove('hidden');
}

function closeCookieSettings() {
    document.getElementById('cookie-settings-modal').classList.add('hidden');
}

function saveCookieSettings() {
    setCookieConsent({
        necessary: true,
        analytics: document.getElementById('cookie-analytics').checked,
        advertising: document.getElementById('cookie-advertising').checked
    });
    closeCookieSettings();
}

function applyCookieSettings(consent) {
    if (consent.analytics) {
        // Enable Google Analytics
        enableAnalytics();
    }
    if (consent.advertising) {
        // Enable ad-related cookies
        enableAdvertising();
    }
}

function enableAnalytics() {
    // Load GA script dynamically
    if (typeof gtag === 'undefined') {
        const script = document.createElement('script');
        script.src = 'https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID';
        script.async = true;
        document.head.appendChild(script);
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', 'GA_MEASUREMENT_ID');
    }
}

function enableAdvertising() {
    // Trigger ad consent system
    if (typeof initAdConsent === 'function') {
        initAdConsent();
    }
}

document.addEventListener('DOMContentLoaded', initCookieConsent);
```

---

### 7.2 Privacy Policy Page

**Goal:** GDPR-compliant privacy policy.

**Files to create:**
```
app/web/templates/privacy.templ
app/internal/http/handlers/staticHandler.go (add privacy route)
app/internal/http/routes/staticRoutes.go
```

**Route:** `GET /privacy`

**Content sections:**
1. Кой сме ние (Who we are)
2. Какви данни събираме (Data we collect)
3. Как използваме данните (How we use data)
4. Бисквитки (Cookies)
5. Споделяне с трети страни (Third-party sharing)
6. Вашите права (Your rights)
7. Контакт (Contact)

---

### 7.3 User Data Export (Right to Access)

**Goal:** Allow users to download all their personal data.

**Files to create/modify:**
```
app/cmd/db/migrations/00012_gdpr.up.sql
app/internal/domain/user/userRepository.go (add export methods)
app/internal/application/user/userService.go (add ExportUserData)
app/internal/http/handlers/accountHandler.go (add export endpoint)
app/web/templates/account/settings.templ (add export button)
```

**Handler:**
```go
// GET /account/export
func (h *AccountHandler) ExportData(w http.ResponseWriter, r *http.Request) {
    user := ctxutils.GetUser(r.Context())

    data := h.userService.ExportUserData(r.Context(), user.Id)

    // Generate JSON file
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition",
        fmt.Sprintf("attachment; filename=my-data-%s.json", time.Now().Format("2006-01-02")))

    json.NewEncoder(w).Encode(data)
}
```

**Export Data Structure:**
```go
type UserDataExport struct {
    ExportedAt time.Time `json:"exported_at"`
    User       struct {
        Email     string    `json:"email"`
        FirstName string    `json:"first_name"`
        LastName  string    `json:"last_name"`
        CreatedAt time.Time `json:"created_at"`
    } `json:"user"`
    Posts []struct {
        Title       string    `json:"title"`
        CreatedAt   time.Time `json:"created_at"`
        PublishedAt *time.Time `json:"published_at"`
    } `json:"posts"`
    ActivityLog []struct {
        Action    string    `json:"action"`
        Timestamp time.Time `json:"timestamp"`
        IpAddress string    `json:"ip_address"`
    } `json:"activity_log"`
}
```

---

### 7.4 Account Deletion (Right to be Forgotten)

**Goal:** Allow users to permanently delete their account and data.

**Files to modify:**
```
app/internal/domain/user/userRepository.go (add delete methods)
app/internal/application/user/userService.go (add DeleteAccount)
app/internal/http/handlers/accountHandler.go (add delete endpoint)
app/web/templates/account/settings.templ (add delete section)
app/web/templates/account/delete-confirm.templ (confirmation page)
```

**Deletion Flow:**
1. User clicks "Изтрий акаунта ми" (Delete my account)
2. Confirmation page explains consequences
3. User enters password to confirm
4. Account soft-deleted, anonymized after 30 days grace period
5. Confirmation email sent

**Handler:**
```go
// POST /account/delete
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
    user := ctxutils.GetUser(r.Context())
    password := r.FormValue("password")

    // Verify password
    if !h.userService.VerifyPassword(r.Context(), user.Id, password) {
        // Return error
        return
    }

    // Schedule deletion (30-day grace period)
    h.userService.ScheduleAccountDeletion(r.Context(), user.Id)

    // Clear session
    httputils.ClearAuthCookies(w)

    // Redirect to confirmation
    http.Redirect(w, r, "/account/deleted", http.StatusSeeOther)
}
```

**Database Migration:**
```sql
ALTER TABLE users ADD COLUMN deletion_scheduled_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;

-- Background job will permanently delete after 30 days
-- and anonymize related data (posts become "Изтрит потребител")
```

**Anonymization (after 30 days):**
```go
func (s *UserService) PermanentlyDeleteUser(ctx context.Context, userId uuid.UUID) error {
    // Anonymize posts (keep content, remove author link)
    s.postRepo.AnonymizeByUser(ctx, userId, "Изтрит потребител")

    // Delete personal data
    s.userRepo.PermanentlyDelete(ctx, userId)

    // Delete from audit logs older than legal requirement
    s.auditRepo.DeleteByUser(ctx, userId)

    return nil
}
```

---

### 7.5 Consent Management

**Goal:** Track and manage user consents.

**Files to create:**
```
app/cmd/db/migrations/00013_consent_log.up.sql
app/internal/domain/consent/consent.go
app/internal/domain/consent/consentRepository.go
```

**Database Migration:**
```sql
CREATE TABLE consent_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  session_id VARCHAR(64), -- for anonymous users
  consent_type VARCHAR(50) NOT NULL, -- 'cookies', 'ads', 'newsletter'
  consented BOOLEAN NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_consent_logs_user ON consent_logs(user_id, consent_type);
CREATE INDEX idx_consent_logs_session ON consent_logs(session_id, consent_type);
```

**Consent Types:**
- `cookies_necessary` - Always true
- `cookies_analytics` - Analytics tracking
- `cookies_advertising` - Ad tracking
- `ads` - Ad consent popup
- `newsletter` - Email marketing

**Log Consent:**
```go
func (r *ConsentRepository) LogConsent(ctx context.Context, log ConsentLog) error {
    sql := `
        INSERT INTO consent_logs (user_id, session_id, consent_type, consented, ip_address, user_agent)
        VALUES ($1, $2, $3, $4, $5, $6)`
    _, err := r.db.ExecContext(ctx, sql,
        log.UserId, log.SessionId, log.ConsentType, log.Consented, log.IpAddress, log.UserAgent)
    return err
}
```

---

### 7.6 Integration with Ad Consent

The existing ad consent system (6.1) should integrate with GDPR cookie consent:

**Updated Flow:**
1. Cookie banner appears first (GDPR requirement)
2. If user accepts advertising cookies → Ad consent popup can appear
3. If user declines advertising cookies → No ad popup, ads disabled
4. All consents logged to `consent_logs` table

**Update `ad-consent.js`:**
```javascript
function initAdConsent() {
    const cookieConsent = getCookieConsent();

    // Only show ad consent if advertising cookies are accepted
    if (!cookieConsent || !cookieConsent.advertising) {
        return; // Don't show ad consent
    }

    // ... existing ad consent logic
}
```

---

## Milestone 8: Social Login (OAuth2)

### 8.1 OAuth2 Infrastructure

**Goal:** Allow users to login/register with social accounts.

**Supported Providers:**
- Google
- Facebook
- Apple (optional)
- GitHub (optional, for developers)

**Files to create:**
```
app/cmd/db/migrations/00014_social_auth.up.sql
app/cmd/db/migrations/00014_social_auth.down.sql
app/internal/infrastructure/oauth/oauth.go
app/internal/infrastructure/oauth/google.go
app/internal/infrastructure/oauth/facebook.go
app/internal/infrastructure/oauth/apple.go
app/internal/application/auth/socialAuthService.go
app/internal/http/handlers/oauthHandler.go
app/internal/http/routes/oauthRoutes.go
app/web/templates/auth/login.templ (update with social buttons)
app/web/templates/auth/register.templ (update with social buttons)
```

**Database Migration:**
```sql
-- Link social accounts to users
CREATE TABLE user_social_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(20) NOT NULL, -- 'google', 'facebook', 'apple', 'github'
  provider_user_id VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  name VARCHAR(255),
  avatar_url VARCHAR(500),
  access_token TEXT,
  refresh_token TEXT,
  token_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ,

  CONSTRAINT uq_social_account UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_social_accounts_user ON user_social_accounts(user_id);
CREATE INDEX idx_social_accounts_provider ON user_social_accounts(provider, provider_user_id);

-- Allow users without password (social-only accounts)
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Add avatar support
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
```

**Configuration (add to `config.go`):**
```go
// Google OAuth
func GoogleClientId() string     { return os.Getenv("GOOGLE_CLIENT_ID") }
func GoogleClientSecret() string { return os.Getenv("GOOGLE_CLIENT_SECRET") }

// Facebook OAuth
func FacebookClientId() string     { return os.Getenv("FACEBOOK_CLIENT_ID") }
func FacebookClientSecret() string { return os.Getenv("FACEBOOK_CLIENT_SECRET") }

// Apple OAuth
func AppleClientId() string     { return os.Getenv("APPLE_CLIENT_ID") }
func AppleTeamId() string       { return os.Getenv("APPLE_TEAM_ID") }
func AppleKeyId() string        { return os.Getenv("APPLE_KEY_ID") }
func ApplePrivateKey() string   { return os.Getenv("APPLE_PRIVATE_KEY") }

// GitHub OAuth
func GitHubClientId() string     { return os.Getenv("GITHUB_CLIENT_ID") }
func GitHubClientSecret() string { return os.Getenv("GITHUB_CLIENT_SECRET") }

func OAuthCallbackURL() string { return config.BaseURL() + "/auth/callback" }
```

---

### 8.2 OAuth Provider Interface

**File:** `app/internal/infrastructure/oauth/oauth.go`

```go
package oauth

import (
    "context"
    "golang.org/x/oauth2"
)

type Provider string

const (
    ProviderGoogle   Provider = "google"
    ProviderFacebook Provider = "facebook"
    ProviderApple    Provider = "apple"
    ProviderGitHub   Provider = "github"
)

type UserInfo struct {
    Provider       Provider
    ProviderUserId string
    Email          string
    Name           string
    FirstName      string
    LastName       string
    AvatarUrl      string
}

type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
    GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}

type OAuthManager struct {
    providers map[Provider]OAuthProvider
}

func NewOAuthManager() *OAuthManager {
    m := &OAuthManager{
        providers: make(map[Provider]OAuthProvider),
    }

    // Register enabled providers
    if config.GoogleClientId() != "" {
        m.providers[ProviderGoogle] = NewGoogleProvider()
    }
    if config.FacebookClientId() != "" {
        m.providers[ProviderFacebook] = NewFacebookProvider()
    }
    if config.AppleClientId() != "" {
        m.providers[ProviderApple] = NewAppleProvider()
    }
    if config.GitHubClientId() != "" {
        m.providers[ProviderGitHub] = NewGitHubProvider()
    }

    return m
}

func (m *OAuthManager) GetProvider(provider Provider) (OAuthProvider, bool) {
    p, ok := m.providers[provider]
    return p, ok
}

func (m *OAuthManager) EnabledProviders() []Provider {
    var providers []Provider
    for p := range m.providers {
        providers = append(providers, p)
    }
    return providers
}
```

---

### 8.3 Google Provider

**File:** `app/internal/infrastructure/oauth/google.go`

```go
package oauth

import (
    "context"
    "encoding/json"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type GoogleProvider struct {
    config *oauth2.Config
}

func NewGoogleProvider() *GoogleProvider {
    return &GoogleProvider{
        config: &oauth2.Config{
            ClientID:     config.GoogleClientId(),
            ClientSecret: config.GoogleClientSecret(),
            RedirectURL:  config.OAuthCallbackURL() + "/google",
            Scopes: []string{
                "https://www.googleapis.com/auth/userinfo.email",
                "https://www.googleapis.com/auth/userinfo.profile",
            },
            Endpoint: google.Endpoint,
        },
    }
}

func (p *GoogleProvider) GetAuthURL(state string) string {
    return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return p.config.Exchange(ctx, code)
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    client := p.config.Client(ctx, token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var data struct {
        Id            string `json:"id"`
        Email         string `json:"email"`
        Name          string `json:"name"`
        GivenName     string `json:"given_name"`
        FamilyName    string `json:"family_name"`
        Picture       string `json:"picture"`
        VerifiedEmail bool   `json:"verified_email"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, err
    }

    return &UserInfo{
        Provider:       ProviderGoogle,
        ProviderUserId: data.Id,
        Email:          data.Email,
        Name:           data.Name,
        FirstName:      data.GivenName,
        LastName:       data.FamilyName,
        AvatarUrl:      data.Picture,
    }, nil
}
```

---

### 8.4 Facebook Provider

**File:** `app/internal/infrastructure/oauth/facebook.go`

```go
package oauth

import (
    "context"
    "encoding/json"
    "fmt"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/facebook"
)

type FacebookProvider struct {
    config *oauth2.Config
}

func NewFacebookProvider() *FacebookProvider {
    return &FacebookProvider{
        config: &oauth2.Config{
            ClientID:     config.FacebookClientId(),
            ClientSecret: config.FacebookClientSecret(),
            RedirectURL:  config.OAuthCallbackURL() + "/facebook",
            Scopes:       []string{"email", "public_profile"},
            Endpoint:     facebook.Endpoint,
        },
    }
}

func (p *FacebookProvider) GetAuthURL(state string) string {
    return p.config.AuthCodeURL(state)
}

func (p *FacebookProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return p.config.Exchange(ctx, code)
}

func (p *FacebookProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    client := p.config.Client(ctx, token)
    resp, err := client.Get("https://graph.facebook.com/me?fields=id,email,name,first_name,last_name,picture.type(large)")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var data struct {
        Id        string `json:"id"`
        Email     string `json:"email"`
        Name      string `json:"name"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Picture   struct {
            Data struct {
                Url string `json:"url"`
            } `json:"data"`
        } `json:"picture"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, err
    }

    return &UserInfo{
        Provider:       ProviderFacebook,
        ProviderUserId: data.Id,
        Email:          data.Email,
        Name:           data.Name,
        FirstName:      data.FirstName,
        LastName:       data.LastName,
        AvatarUrl:      data.Picture.Data.Url,
    }, nil
}
```

---

### 8.5 OAuth Handler

**File:** `app/internal/http/handlers/oauthHandler.go`

```go
package handlers

type OAuthHandler struct {
    oauthManager     *oauth.OAuthManager
    socialAuthService *auth.SocialAuthService
}

// GET /auth/{provider} - Redirect to provider
func (h *OAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
    providerName := r.PathValue("provider")
    provider, ok := h.oauthManager.GetProvider(oauth.Provider(providerName))
    if !ok {
        http.Error(w, "Unsupported provider", http.StatusBadRequest)
        return
    }

    // Generate state token for CSRF protection
    state := generateSecureToken()
    http.SetCookie(w, &http.Cookie{
        Name:     "oauth_state",
        Value:    state,
        Path:     "/",
        MaxAge:   300, // 5 minutes
        HttpOnly: true,
        Secure:   config.IsProduction(),
        SameSite: http.SameSiteLaxMode,
    })

    authURL := provider.GetAuthURL(state)
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GET /auth/callback/{provider} - Handle OAuth callback
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    providerName := r.PathValue("provider")
    provider, ok := h.oauthManager.GetProvider(oauth.Provider(providerName))
    if !ok {
        http.Error(w, "Unsupported provider", http.StatusBadRequest)
        return
    }

    // Verify state
    stateCookie, err := r.Cookie("oauth_state")
    if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // Clear state cookie
    http.SetCookie(w, &http.Cookie{
        Name:   "oauth_state",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    })

    // Check for error from provider
    if errMsg := r.URL.Query().Get("error"); errMsg != "" {
        slog.Warn("OAuth error", "provider", providerName, "error", errMsg)
        http.Redirect(w, r, "/login?error=oauth_denied", http.StatusSeeOther)
        return
    }

    // Exchange code for token
    code := r.URL.Query().Get("code")
    token, err := provider.ExchangeCode(r.Context(), code)
    if err != nil {
        slog.Error("Failed to exchange code", "error", err)
        http.Redirect(w, r, "/login?error=oauth_failed", http.StatusSeeOther)
        return
    }

    // Get user info from provider
    userInfo, err := provider.GetUserInfo(r.Context(), token)
    if err != nil {
        slog.Error("Failed to get user info", "error", err)
        http.Redirect(w, r, "/login?error=oauth_failed", http.StatusSeeOther)
        return
    }

    // Login or register user
    user, isNew, err := h.socialAuthService.LoginOrRegister(r.Context(), userInfo, token)
    if err != nil {
        slog.Error("Failed to login/register", "error", err)
        http.Redirect(w, r, "/login?error=oauth_failed", http.StatusSeeOther)
        return
    }

    // Generate JWT tokens
    accessToken, accessExp := securityutil.GenerateAccessToken(*user, false)
    refreshToken, refreshExp := securityutil.GenerateRefreshToken(*user)

    // Set cookies
    httputils.SetAuthCookies(w, accessToken, accessExp, refreshToken, refreshExp)

    // Redirect
    if isNew {
        http.Redirect(w, r, "/welcome", http.StatusSeeOther)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}
```

---

### 8.6 Social Auth Service

**File:** `app/internal/application/auth/socialAuthService.go`

```go
package auth

type SocialAuthService struct {
    userRepo          *user.UserRepository
    socialAccountRepo *user.SocialAccountRepository
}

func (s *SocialAuthService) LoginOrRegister(
    ctx context.Context,
    info *oauth.UserInfo,
    token *oauth2.Token,
) (*user.User, bool, error) {

    // Check if social account already exists
    socialAccount, err := s.socialAccountRepo.FindByProvider(
        ctx, string(info.Provider), info.ProviderUserId)

    if err == nil && socialAccount != nil {
        // Existing user - update tokens and return
        s.socialAccountRepo.UpdateTokens(ctx, socialAccount.Id, token)
        user, _ := s.userRepo.FindById(ctx, socialAccount.UserId)
        return user, false, nil
    }

    // Check if email already registered
    existingUser, _ := s.userRepo.FindByEmail(ctx, info.Email)
    if existingUser != nil {
        // Link social account to existing user
        s.createSocialAccount(ctx, existingUser.Id, info, token)
        return existingUser, false, nil
    }

    // Create new user
    newUser := &user.User{
        Id:        uuid.New(),
        Email:     info.Email,
        FirstName: info.FirstName,
        LastName:  info.LastName,
        AvatarUrl: &info.AvatarUrl,
        Role:      "USER",
        CreatedAt: time.Now().UTC(),
    }

    if err := s.userRepo.Create(ctx, newUser); err != nil {
        return nil, false, err
    }

    // Create social account link
    s.createSocialAccount(ctx, newUser.Id, info, token)

    return newUser, true, nil
}

func (s *SocialAuthService) createSocialAccount(
    ctx context.Context,
    userId uuid.UUID,
    info *oauth.UserInfo,
    token *oauth2.Token,
) error {
    account := &user.SocialAccount{
        Id:             uuid.New(),
        UserId:         userId,
        Provider:       string(info.Provider),
        ProviderUserId: info.ProviderUserId,
        Email:          info.Email,
        Name:           info.Name,
        AvatarUrl:      info.AvatarUrl,
        AccessToken:    token.AccessToken,
        RefreshToken:   token.RefreshToken,
        TokenExpiresAt: token.Expiry,
        CreatedAt:      time.Now().UTC(),
    }
    return s.socialAccountRepo.Create(ctx, account)
}
```

---

### 8.7 Login/Register Templates

**Update:** `app/web/templates/auth/login.templ`

```go
templ LoginPage(providers []oauth.Provider, errorMsg string) {
    <div class="max-w-md mx-auto mt-10">
        <h1 class="text-2xl font-bold mb-6">Вход</h1>

        if errorMsg != "" {
            <div class="bg-red-100 text-red-700 p-3 rounded mb-4">{ errorMsg }</div>
        }

        <!-- Social Login Buttons -->
        if len(providers) > 0 {
            <div class="space-y-3 mb-6">
                for _, provider := range providers {
                    @SocialLoginButton(provider)
                }
            </div>

            <div class="relative mb-6">
                <div class="absolute inset-0 flex items-center">
                    <div class="w-full border-t border-gray-300"></div>
                </div>
                <div class="relative flex justify-center text-sm">
                    <span class="px-2 bg-white text-gray-500">или</span>
                </div>
            </div>
        }

        <!-- Email/Password Form -->
        <form method="POST" action="/login" class="space-y-4">
            <!-- CSRF token -->
            <input type="hidden" name="_csrf" value={ ctx.Value("csrf_token").(string) } />

            <div>
                <label class="block text-sm font-medium mb-1">Имейл</label>
                <input type="email" name="email" required
                       class="w-full px-3 py-2 border rounded focus:ring-2 focus:ring-blue-500" />
            </div>

            <div>
                <label class="block text-sm font-medium mb-1">Парола</label>
                <input type="password" name="password" required
                       class="w-full px-3 py-2 border rounded focus:ring-2 focus:ring-blue-500" />
            </div>

            <button type="submit"
                    class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700">
                Вход
            </button>
        </form>

        <p class="mt-4 text-center text-sm text-gray-600">
            Нямате акаунт?
            <a href="/register" class="text-blue-600 hover:underline">Регистрация</a>
        </p>
    </div>
}

templ SocialLoginButton(provider oauth.Provider) {
    switch provider {
        case oauth.ProviderGoogle:
            <a href="/auth/google"
               class="flex items-center justify-center gap-3 w-full px-4 py-2 border rounded hover:bg-gray-50">
                <svg class="w-5 h-5" viewBox="0 0 24 24">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                </svg>
                <span>Продължи с Google</span>
            </a>
        case oauth.ProviderFacebook:
            <a href="/auth/facebook"
               class="flex items-center justify-center gap-3 w-full px-4 py-2 bg-[#1877F2] text-white rounded hover:bg-[#166FE5]">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
                </svg>
                <span>Продължи с Facebook</span>
            </a>
        case oauth.ProviderApple:
            <a href="/auth/apple"
               class="flex items-center justify-center gap-3 w-full px-4 py-2 bg-black text-white rounded hover:bg-gray-800">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
                </svg>
                <span>Продължи с Apple</span>
            </a>
        case oauth.ProviderGitHub:
            <a href="/auth/github"
               class="flex items-center justify-center gap-3 w-full px-4 py-2 bg-[#24292F] text-white rounded hover:bg-[#1C2128]">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                <span>Продължи с GitHub</span>
            </a>
    }
}
```

---

### 8.8 Routes

**File:** `app/internal/http/routes/oauthRoutes.go`

```go
func OAuthRoutes(mux *http.ServeMux, oauthHandler *handlers.OAuthHandler) {
    // Initiate OAuth flow
    mux.HandleFunc("GET /auth/{provider}", oauthHandler.InitiateOAuth)

    // OAuth callbacks
    mux.HandleFunc("GET /auth/callback/{provider}", oauthHandler.HandleCallback)
}
```

---

### 8.9 Account Linking (Optional)

Allow users to link/unlink social accounts from settings:

**Routes:**
- `GET /account/connections` - View linked accounts
- `POST /account/connections/{provider}` - Link new provider
- `DELETE /account/connections/{provider}` - Unlink provider

**Rules:**
- Can't unlink last auth method (must have password OR at least one social)
- Linking requires OAuth flow
- Show which accounts are linked in settings

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
| 11 | Cookie Consent (GDPR) | Medium | High |
| 12 | Privacy Policy | Low | High |
| 13 | Social Login (Google) | Medium | High |
| 14 | Social Login (Facebook) | Low | Medium |
| 15 | Third-Party Ads | Low | High |
| 16 | Affiliate Links | Low | Medium |
| 17 | Sponsored Posts | Low | Medium |
| 18 | Data Export (GDPR) | Medium | Medium |
| 19 | Account Deletion (GDPR) | Medium | Medium |
| 20 | Self-Hosted Ads | High | High |
