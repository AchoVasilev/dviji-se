# Security TODO

## Critical Priority

### Deployment blockers
- [x] Bootstrap the first administrator
  - Registration only grants USER, so a fresh database had no way into the
    admin panel; `ADMIN_EMAIL`/`ADMIN_PASSWORD` seed one on first start
  - Ignored once an administrator exists, so it cannot reset an account
- [x] Build the about page (`/about`) and advertise it by default
- [x] Hide nav links to pages that do not exist (`/workouts`, `/nutrition`)
  - Kept in the markup behind `ENABLED_WORKOUTS`, `ENABLED_NUTRITION` and
    `ENABLED_ABOUT`, all defaulting to off
  - The flags control visibility only; each still needs a route before it is
    turned on
- [x] Stop fragment endpoints serving partial pages
  - `/blog/recent`, `/blog/search/suggestions` and `/categories` returned
    unstyled markup to anyone opening the URL directly
  - `middleware.FragmentOnly` redirects non-HTMX requests to the hosting page
    and marks the responses noindex
- [x] Point the mobile Тренировки / Хранене entries at the category pages
  - They fall back to the category page and switch to the dedicated section
    when `ENABLED_WORKOUTS` / `ENABLED_NUTRITION` are turned on
- [x] Hide the public login entries when registration is off
  - Two mobile links (menu and bottom bar) were ungated and pointed at a route
    that is not registered, so they 404'd
- [x] Default `ALLOW_REGISTRATION` to off
- [x] Redirect the public auth paths home when registration is off
  - They stay registered so a stale bookmark lands on the site rather than a
    404; `/admin/login` is deliberately excluded

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
- [x] Refuse non administrators at `/admin/login`
  - It stays registered when public registration is off, so it was the one
    entry point that would hand a session to any account with a valid password
  - Checked after the password and answered with the same message, so the
    response cannot be used to find out which addresses are administrators
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

### Denial of service
- [x] Set the HTTP server timeouts
  - `http.Server` had none, so a client could hold a connection open by
    trickling a request one byte at a time
  - Header 10s, body 60s (a 12 MiB upload over a slow link), write 60s,
    idle 120s

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
- [x] Cache the revocation lookup
  - Static assets skip auth entirely: the auth cookie rides along on every one,
    so a page view cost one lookup per asset (measured 4 for a page + 3 assets)
  - Remaining page requests read a 30s TTL cache of the cutoff, so revocation is
    eventually consistent within that window

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
- [x] Add Open Graph meta tags for social sharing
- [x] Add Schema.org Article markup
  - `templates.SEO` drives title, canonical, Open Graph, Twitter card and
    JSON-LD; posts become og:type=article with author, section and dates
  - Search results deliberately emit no canonical so they are not indexed
  - Requires `APP_BASE_URL` to be correct in production: canonical and og:url
    are absolute and built from it

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
- [x] Token revocation (cutoff per user, applied on password change)
- [x] Password complexity validation (12+ chars, mixed case, numbers, symbols)
- [ ] Audit logging (login, logout, password changes)

## Milestone 5: SEO & Social
- [x] Open Graph meta tags
- [x] Schema.org Article markup (JSON-LD)

## Milestone 6: Monetization & Ads
- [x] Ad consent system (popup → floating button, ads only if consented)
  - `templates.Consent` renders both dialogs and the button; `consent.js`
    stores the answer in a `consent` cookie (`necessary` or `ads`)
  - `templates.AdSlot(id, variant)` is one component for every placement:
    `AdRail` beside the content from 1024px up, `AdBanner` full width and one
    viewport tall below it, `AdInline` anywhere in a page
  - Both layout slots live in `LayoutSEO`, so every public page carries them
  - Slots hold no third party markup and are display:none until consent.js
    marks the document `ads-on` - visibility is a stylesheet concern because
    which placement applies depends on the viewport, not on the answer
  - The button stays after either answer, so consent can be withdrawn as
    easily as it was given
- [ ] Third-party ad networks (Google AdSense)
  - Register `window.dvijiSe.renderAds`; `consent.js` calls it only once
    advertising is consented to
  - Needs the public CSP widened for the network's script and frame hosts,
    and `AdSlot` placed on the pages that should carry ads
- [ ] Affiliate links tracking (`/go/{slug}`)
- [ ] Sponsored posts (sponsor badge, sponsor fields)
- [ ] Self-hosted ads (full ad management system)

## Milestone 7: GDPR & Privacy Compliance
- [x] Cookie consent banner (necessary + advertising)
  - Asked once on the first visit and unanswerable by dismissal, so no
    non-essential cookie is set without a choice
  - Analytics is deliberately not offered: nothing measures visitors yet, and
    a category that does nothing is a consent nobody can act on
- [x] Privacy policy page (`/privacy`)
  - Describes what the app actually does: the four cookies by name, the IP
    kept in memory by the rate limiter, hashed reset tokens, and the fact
    that no ad network is wired in yet
  - Linked from the footer and from the cookie dialog, and listed in the
    sitemap along with `/about`, which was missing from it
  - `PRIVACY_CONTACT_EMAIL` names the mailbox; it falls back to
    privacy@<host of APP_BASE_URL>
  - Still owner supplied: the legal identity of whoever runs the site, if it
    is operated as a business rather than personally
- [ ] User data export (`/account/export` - right to access)
- [ ] Account deletion (`/account/delete` - right to be forgotten)
- [ ] Consent logging (track all user consents)
- [x] Integration: cookie consent → ad consent flow
  - Accepting everything in the cookie dialog answers the advertising
    question too; the floating button reopens it later

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
