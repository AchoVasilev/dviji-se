# Dviji Se

A Bulgarian fitness blog platform built with Go, featuring an admin panel for content management and public blog pages for readers.

## Features

- **Blog System**: Posts with categories, slugs, excerpts, cover images, and SEO metadata
- **Admin Panel**: Dashboard, post management with TinyMCE editor, image uploads via Cloudinary
- **Authentication**: JWT-based auth with access/refresh tokens, remember me, password reset
- **Security**: CSRF protection, rate limiting, role-based access control (RBAC), security headers
- **Templates**: Server-side rendering with Templ and HTMX for interactivity

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL
- **Templates**: [Templ](https://templ.guide/)
- **Styling**: Tailwind CSS
- **Editor**: TinyMCE
- **Image Storage**: Cloudinary
- **Migrations**: golang-migrate

## Project Structure

```
app/
├── cmd/
│   ├── main.go                 # Application entry point
│   └── db/
│       ├── database/           # Database connection
│       └── migrations/         # SQL migrations
├── internal/
│   ├── config/                 # Centralized configuration
│   ├── domain/                 # Domain models and repositories
│   │   ├── posts/
│   │   ├── category/
│   │   └── user/
│   ├── application/            # Business logic services
│   │   ├── posts/
│   │   ├── categories/
│   │   ├── auth/
│   │   └── users/
│   ├── infrastructure/         # External services
│   │   ├── cloudinary/
│   │   ├── email/
│   │   └── environment/
│   ├── http/
│   │   ├── handlers/           # HTTP handlers
│   │   ├── middleware/         # Middleware (auth, CORS, CSRF, rate limiting)
│   │   └── routes/             # Route definitions
│   └── server/                 # HTTP server setup
├── web/
│   ├── static/                 # Static assets (CSS, JS, images)
│   └── templates/              # Templ templates
│       ├── admin/              # Admin panel templates
│       └── *.templ             # Public templates
├── util/                       # Utility packages
└── tests/                      # Integration tests
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- Node.js (for Tailwind CSS)

### Environment Variables

Create a `local.env` file:

```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DBHOST=localhost
DBPORT=5432
DBUSER=postgres
DBPASSWORD=your_password
DBNAME=dviji_se
DBSSLMODE=disable
DBCONNECTIONS=10

# JWT (use strong secrets in production)
JWT_KEY=your-jwt-secret-key
JWT_REFRESH_KEY=your-jwt-refresh-secret-key

# Security
XSRF=your-xsrf-secret-key
CORS_ORIGINS=http://localhost:3000
ALLOW_REGISTRATION=false

# SMTP (optional - emails logged in dev mode if not configured)
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@example.com

# Cloudinary (optional - required for image uploads)
CLOUDINARY_CLOUD_NAME=
CLOUDINARY_API_KEY=
CLOUDINARY_API_SECRET=
CLOUDINARY_FOLDER=dviji-se/blog

# App
APP_BASE_URL=http://localhost:8080
```

### Running Locally

1. **Install dependencies**
   ```bash
   cd app
   go mod download
   npm install
   ```

2. **Generate templates**
   ```bash
   templ generate
   ```

3. **Build Tailwind CSS**
   ```bash
   npx @tailwindcss/cli -i ./web/static/css/input.css -o ./web/static/css/styles.css
   ```

4. **Run the server**
   ```bash
   go run cmd/main.go
   ```

### Running with Docker

```bash
docker build -t dviji-se .
docker run -p 8080:8080 --env-file local.env dviji-se
```

## Testing

```bash
# Run all tests
go test ./...

# Run unit tests only (skip integration)
go test -short ./...

# Run with verbose output
go test -v ./...
```

Integration tests use [testcontainers](https://testcontainers.com/) to spin up a PostgreSQL container.

## API Routes

### Public
- `GET /` - Home page
- `GET /blog` - Blog listing
- `GET /blog/{slug}` - Single post
- `GET /blog/category/{slug}` - Posts by category
- `GET /health` - Health check

### Authentication
- `GET /login` - Login page
- `POST /login` - Authenticate
- `GET /register` - Register page (if enabled)
- `POST /register` - Create account (if enabled)
- `POST /refresh-token` - Refresh JWT
- `GET /forgot-password` - Forgot password page
- `POST /forgot-password` - Request password reset
- `GET /reset-password` - Reset password page
- `POST /reset-password` - Set new password

### Admin (requires ADMIN role)
- `GET /admin` - Dashboard
- `GET /admin/posts` - Posts list
- `GET /admin/posts/new` - Create post form
- `GET /admin/posts/{id}` - Edit post form
- `POST /admin/posts` - Create post
- `PUT /admin/posts/{id}` - Update post
- `DELETE /admin/posts/{id}` - Delete post
- `POST /admin/upload` - Upload image

## Configuration

All configuration is centralized in `internal/config/config.go`. Access values using:

```go
import "server/internal/config"

port := config.Port()
dbHost := config.DBHost()
jwtKey := config.JWTAccessKey()
```

## Security Features

- **CSRF Protection**: Double-submit cookie pattern with XSRF tokens
- **Rate Limiting**: Per-IP limits on auth endpoints (5 login attempts/15min)
- **Security Headers**: X-Frame-Options, X-Content-Type-Options, CSP, Referrer-Policy
- **Password Hashing**: bcrypt with cost factor 12
- **JWT**: Short-lived access tokens (15min), longer refresh tokens (24h)

## License

MIT
