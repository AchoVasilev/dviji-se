# Security TODO

## Critical Priority

### Rate Limiting
- [x] Implement rate limiting middleware for authentication endpoints
  - Login: 5 attempts per 15 minutes per IP
  - Register: 3 accounts per hour per IP
  - Implemented in `internal/http/middleware/ratelimit.go`
- [x] Only trust X-Forwarded-For / X-Real-IP from configured proxies
  - `TRUSTED_PROXIES` (CIDRs or IPs); empty means trust none
  - Without this any client could rotate the header and bypass the limit
- [x] Bound the rate limiter map
  - Capped at 50k tracked clients; expired entries are reclaimed first, then
    the oldest is evicted

### Authorization
- [x] Create authorization middleware to check user roles/permissions
- [x] Protect admin endpoints (RequireAuth + RequireAdmin middleware)
- [x] Add role-based access control (RBAC) checks in handlers
- [x] Protect category endpoints (`POST /categories` was fully unauthenticated)
- [x] Protect user-specific endpoints - nothing to protect yet
  - Audited every route: all are public content, auth entry points, or already
    behind RequireAuth+RequireAdmin. There are no user-scoped endpoints
  - Revisit when the account pages arrive (bookmarks, data export, deletion)

### CSRF
- [x] Compare the CSRF cookie against the submitted header
  - Each was validated on its own, so any server-issued token matched any other
- [x] Bind the CSRF token to the session
  - Tokens now carry the user id when logged in, and a per browser id when not
  - CSRFCookie re-mints whenever the identity changes, so login and logout
    rotate the token without handler changes
  - CheckAuth moved ahead of the CSRF middlewares so the identity is known

## High Priority

### Password Security
- [x] Increase bcrypt cost from 10 to 12 in `util/securityutil/password.go`
- [x] Add password complexity validation (uppercase, lowercase, numbers, special chars)
- [x] Increase minimum password length to 12 characters
  - `securityutil.IsPasswordStrong` backs both the `strongpassword` validator
    tag and the reset service, so one policy governs every path
  - Login deliberately enforces no policy: existing accounts hold weaker
    passwords and must still be able to sign in

### Security Headers
- [x] Re-enable Content-Security-Policy middleware in `server.go`
- [x] Add X-Frame-Options: DENY
- [x] Add X-Content-Type-Options: nosniff
- [x] Add Strict-Transport-Security header for production (production only, so
      dev does not pin localhost to HTTPS)

## Medium Priority

### Token Management
- [x] Implement the refresh token flow
  - Login stores the refresh token in a cookie scoped to `/refresh-token`
  - `POST /refresh-token` reloads the user from the database rather than
    trusting the token, so role changes and deletions take effect on refresh
  - Registered outside the `ALLOW_REGISTRATION` block so admin sessions refresh
- [x] Add token revocation
  - `users.tokens_valid_after`: any access token issued at or before it is
    refused. No token store needed, and it covers "revoke all sessions"
  - Checked in `CheckAuth` and again when refreshing; failures fail closed
- [x] Invalidate tokens on password change
  - `UpdatePassword` sets the cutoff in the same statement
- [ ] Rotate the refresh token on use
  - Refresh currently re-issues only the access token, so a stolen refresh
    token stays usable until it expires
- [ ] Cache the revocation lookup
  - It costs one query per authenticated request

### Bug Fixes
- [x] Fix user context type assertion in `util/ctxutils/ctxutils.go:66-73`
  - Fixed: now correctly asserts `*LoggedInUser` pointer type
- [x] Stop mutating the global logger per request
  - `slog.SetDefault` in a middleware raced and attributed ids to the wrong
    request; the default handler now reads the id from the record's context
- [x] Fix printf-style `slog` calls that emitted a literal `%v`
  - Panic stack traces in `recovery.go` were unreadable
- [x] Set response headers before writing the body
  - `csrfError` and `recovery.go` set them afterwards, so they were dropped
- [x] Fix nil dereference in `CheckAuth`
  - An invalid bearer header with no auth cookie panicked on `authCookie.Value`
- [x] Fix inverted `ValidationResult.Success` in `util/httputils/validation.go`
- [x] Limit request body size
  - `LimitRequestBody` middleware: 1 MiB default, 12 MiB on the upload routes;
    oversized bodies return 413
- [x] Guard the empty-permissions case in `UserRepository.Create`
- [x] Fix the blog card crash on authors with no name
  - `[]rune(post.AuthorFirstName)[0]` panicked, and registration never collects
    names, so any post by a self-registered author 500s the blog and search pages

### Logging
- [x] Migrate request-scoped `slog` calls to the `*Context` variants
  - Handlers, middleware, services and the email/JSON helpers now carry the
    request id; startup logs (main, database, config) intentionally do not
- [x] Thread a context through the `httputils.Send*` response helpers

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
- [x] Search posts by title, content, excerpt
  - `GET /blog/search?q={query}` plus `GET /blog/search/suggestions`
  - Search results page template
- [x] Replace the `ILIKE '%q%'` scan with PostgreSQL full-text search
  - Generated `tsvector` column + GIN index, weighted title > excerpt > content
  - Measured on 205k rows: 163ms parallel seq scan -> 1.7ms index scan
  - Note: matching is now by word (with a prefix match on the last term for
    type-ahead) rather than arbitrary substring

### RSS Feed
- [x] Implement RSS feed endpoint
  - `GET /feed.xml`, published posts ordered by date

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
- [x] Add XML sitemap (`/sitemap.xml`) and `robots.txt`
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
- [x] Search for posts (`GET /blog/search?q=`) - full-text search, GIN indexed
- [x] RSS feed (`GET /feed.xml`)
- [x] XML sitemap (`GET /sitemap.xml`)
- [x] robots.txt (`GET /robots.txt`)

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
