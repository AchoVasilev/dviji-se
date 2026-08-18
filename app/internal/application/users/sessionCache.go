package users

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

// DefaultRevocationCacheTTL bounds how long a revocation can go unnoticed.
// Shorter means revocations bite sooner at the cost of more lookups.
const DefaultRevocationCacheTTL = 30 * time.Second

// maxCachedCutoffs bounds the cache so a stream of distinct user ids cannot
// grow it without limit.
const maxCachedCutoffs = 10_000

// cutoffSource reads a user's revocation cutoff. Satisfied by UserRepository.
type cutoffSource interface {
	TokensValidAfter(ctx context.Context, userId string) (sql.NullTime, error)
}

type cachedCutoff struct {
	cutoff    sql.NullTime
	expiresAt time.Time
}

// CachedSessionValidator answers revocation checks from a short lived cache.
// Every authenticated request asks the same question, so without this a single
// page view costs one lookup per request.
//
// The cutoff is cached rather than the verdict, because whether a token is
// still valid depends on when that particular token was issued.
//
// Revocation is eventually consistent: a change made elsewhere - by another
// instance, or by the password reset flow writing straight to the database -
// takes effect once the entry expires, within the TTL.
type CachedSessionValidator struct {
	source cutoffSource
	ttl    time.Duration

	mu      sync.Mutex
	entries map[string]cachedCutoff
}

func NewCachedSessionValidator(source cutoffSource, ttl time.Duration) *CachedSessionValidator {
	if ttl <= 0 {
		ttl = DefaultRevocationCacheTTL
	}

	return &CachedSessionValidator{
		source:  source,
		ttl:     ttl,
		entries: make(map[string]cachedCutoff),
	}
}

// IsSessionValid reports whether a token minted at issuedAt still authenticates.
// Lookup failures fail closed.
func (c *CachedSessionValidator) IsSessionValid(ctx context.Context, userId string, issuedAt time.Time) bool {
	now := time.Now()

	if cutoff, ok := c.cached(userId, now); ok {
		return sessionValidAt(cutoff, issuedAt)
	}

	cutoff, err := c.source.TokensValidAfter(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Could not read the token revocation cutoff", "error", err, "userId", userId)
		return false
	}

	c.store(userId, cutoff, now)

	return sessionValidAt(cutoff, issuedAt)
}

// Forget drops a user's cached cutoff so the next check reads through. Used
// after revoking in this process; the TTL covers every other case.
func (c *CachedSessionValidator) Forget(userId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, userId)
}

func (c *CachedSessionValidator) cached(userId string, now time.Time) (sql.NullTime, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[userId]
	if !ok || now.After(entry.expiresAt) {
		return sql.NullTime{}, false
	}

	return entry.cutoff, true
}

func (c *CachedSessionValidator) store(userId string, cutoff sql.NullTime, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= maxCachedCutoffs {
		c.evictLocked(now)
	}

	c.entries[userId] = cachedCutoff{cutoff: cutoff, expiresAt: now.Add(c.ttl)}
}

// evictLocked reclaims expired entries, falling back to clearing the cache if
// they were all still live. Callers must hold c.mu.
func (c *CachedSessionValidator) evictLocked(now time.Time) {
	for id, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, id)
		}
	}

	if len(c.entries) >= maxCachedCutoffs {
		clear(c.entries)
	}
}
