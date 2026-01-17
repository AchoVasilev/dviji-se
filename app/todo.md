# Security TODO

## Critical Priority

### Rate Limiting
- [ ] Implement rate limiting middleware for authentication endpoints
  - Login: 5 attempts per 15 minutes per IP
  - Register: 3 accounts per hour per IP
  - Consider using `golang.org/x/time/rate` or `github.com/ulule/limiter`

### Authorization
- [ ] Create authorization middleware to check user roles/permissions
- [ ] Protect category endpoints (create, update, delete)
- [ ] Protect user-specific endpoints
- [ ] Add role-based access control (RBAC) checks in handlers

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
- [ ] Implement forgot password flow

## Notes

- `local.env` is for development/testing only - production uses secure secrets management
- JWT secrets in `local.env` are intentionally simple for testing
