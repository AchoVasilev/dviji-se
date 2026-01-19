# Security TODO

## Critical Priority

### Rate Limiting
- [x] Implement rate limiting middleware for authentication endpoints
  - Login: 5 attempts per 15 minutes per IP
  - Register: 3 accounts per hour per IP
  - Implemented in `internal/http/middleware/ratelimit.go`

### Authorization
- [x] Create authorization middleware to check user roles/permissions
- [x] Protect admin endpoints (RequireAuth + RequireAdmin middleware)
- [x] Add role-based access control (RBAC) checks in handlers
- [ ] Protect category endpoints (create, update, delete)
- [ ] Protect user-specific endpoints

## High Priority

### Password Security
- [x] Increase bcrypt cost from 10 to 12 in `util/securityutil/password.go`
- [ ] Add password complexity validation (uppercase, lowercase, numbers, special chars)
- [ ] Increase minimum password length to 12 characters

### Security Headers
- [x] Re-enable Content-Security-Policy middleware in `server.go`
- [x] Add X-Frame-Options: DENY
- [x] Add X-Content-Type-Options: nosniff
- [ ] Add Strict-Transport-Security header for production (requires HTTPS)

## Medium Priority

### Token Management
- [ ] Implement RefreshToken endpoint in `authHandler.go`
- [ ] Add token revocation/blacklist system (Redis or database)
- [ ] Invalidate tokens on password change

### Bug Fixes
- [x] Fix user context type assertion in `util/ctxutils/ctxutils.go:66-73`
  - Fixed: now correctly asserts `*LoggedInUser` pointer type

### CSRF Improvements
- [x] Use actual request method instead of hardcoded POST for CSRF tokens
  - Fixed: now uses empty string for action ID (method-agnostic)

## Low Priority

### Logging & Monitoring
- [ ] Add audit logging for authentication events
- [ ] Log authorization failures
- [ ] Add alerting for suspicious activity

### Password Features
- [ ] Implement password change endpoint
- [x] Implement forgot password flow
  - Password reset token generation and validation
  - Email sending (logs email in dev mode when SMTP not configured)

### Email
- [x] Set up email service infrastructure (`internal/infrastructure/email`)
- [ ] Configure production SMTP (SendGrid/Mailgun/AWS SES)
- [x] Create email templates (password reset)
- [x] Implement password reset token generation and validation

## Testing

- [x] Unit tests: securityutil (password hashing, JWT)
- [x] Unit tests: PostService (slug generation, reading time)
- [x] Unit tests: AuthService and PasswordResetService
- [x] Unit tests: middleware (rate limiting, admin, auth)
- [x] Unit tests: httputils (validation, cookies)
- [x] Integration tests: testcontainers infrastructure
- [x] Integration tests: auth API (register, login, password reset)
- [x] Integration tests: blog API (list, view, category filter)
- [x] Integration tests: admin API (create post, auth/role checks)

---

# Blog Features TODO

## High Priority

### Search
- [ ] Add full-text search for posts
  - Search by title, content, excerpt
  - `GET /blog/search?q={query}`
  - Search results page template
  - Consider PostgreSQL full-text search or add search index

### RSS Feed
- [ ] Implement RSS feed endpoint
  - `GET /feed.xml` or `GET /rss`
  - Include title, description, link, pubDate for each post
  - Only published posts, ordered by date

## Medium Priority

### Post Scheduling
- [ ] Allow scheduling posts for future publication
  - Add `scheduled_at` field to posts table
  - Background job to publish scheduled posts
  - Show scheduled status in admin

### Tags
- [ ] Implement tagging system
  - Create `tags` and `posts_tags` tables
  - Add tag input to post form
  - Filter posts by tag: `GET /blog/tag/{slug}`
  - Display tags on post cards and single post view

### View Counter
- [ ] Track post views
  - Add `view_count` column to posts
  - Increment on each unique view (consider IP/session dedup)
  - Display view count on admin dashboard

### Related Posts
- [ ] Show truly related posts instead of just recent
  - Match by category and/or tags
  - Exclude current post from results

## Low Priority

### SEO & Discovery
- [ ] Add XML sitemap (`/sitemap.xml`)
- [ ] Add Open Graph meta tags for social sharing
- [ ] Add Schema.org Article markup

### Admin Enhancements
- [ ] Post duplication (clone existing post)
- [ ] Bulk actions (publish/archive multiple posts)
- [ ] Advanced filters (date range, author)
- [ ] Post revision history

### User Experience
- [ ] Post preview before publishing
- [ ] Reading list / bookmarks for users
- [ ] Infinite scroll option for blog list

---

# Feature Roadmap

> Detailed implementation plans are in `IMPLEMENTATION_PLAN.md`

## Milestone 1: Search & Discovery
- [ ] Full-text search for posts (`GET /blog/search?q=`)
- [ ] RSS feed (`GET /feed.xml`)
- [ ] XML sitemap (`GET /sitemap.xml`)

## Milestone 2: Content Organization
- [ ] Tags system (many-to-many with posts)
- [ ] Related posts (by tags/category)

## Milestone 3: Admin Enhancements
- [ ] Post scheduling (scheduled_at + background worker)
- [ ] View counter (with IP deduplication)
- [ ] Bulk actions (publish/archive/delete multiple)

## Milestone 4: Security Hardening
- [ ] Token revocation (blacklist on logout/password change)
- [ ] Password complexity validation (12+ chars, mixed case, numbers, symbols)
- [ ] Audit logging (login, logout, password changes)

## Milestone 5: SEO & Social
- [ ] Open Graph meta tags
- [ ] Schema.org Article markup (JSON-LD)

## Milestone 6: Monetization & Ads
- [ ] Ad consent system (popup → minimized widget, ads only if consented)
- [ ] Third-party ad networks (Google AdSense)
- [ ] Affiliate links tracking (`/go/{slug}`)
- [ ] Sponsored posts (sponsor badge, sponsor fields)
- [ ] Self-hosted ads (full ad management system)

## Milestone 7: GDPR & Privacy Compliance
- [ ] Cookie consent banner (necessary/analytics/advertising categories)
- [ ] Privacy policy page (`/privacy`)
- [ ] User data export (`/account/export` - right to access)
- [ ] Account deletion (`/account/delete` - right to be forgotten)
- [ ] Consent logging (track all user consents)
- [ ] Integration: cookie consent → ad consent flow

## Milestone 8: Social Login (OAuth2)
- [ ] OAuth2 infrastructure (provider interface, manager)
- [ ] Google login
- [ ] Facebook login
- [ ] Apple login (optional)
- [ ] GitHub login (optional)
- [ ] Account linking/unlinking from settings

---

## Notes

- `local.env` is for development/testing only - production uses secure secrets management
- JWT secrets in `local.env` are intentionally simple for testing
- Run tests: `go test ./...` (all) or `go test -short ./...` (skip integration)
