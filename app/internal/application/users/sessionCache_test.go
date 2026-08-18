package users

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"
)

type stubCutoffSource struct {
	mu     sync.Mutex
	cutoff sql.NullTime
	err    error
	calls  int
}

func (s *stubCutoffSource) TokensValidAfter(context.Context, string) (sql.NullTime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	return s.cutoff, s.err
}

func (s *stubCutoffSource) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

func (s *stubCutoffSource) setCutoff(cutoff sql.NullTime) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cutoff = cutoff
}

func TestCachedSessionValidator_ReusesTheCachedCutoff(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, time.Minute)

	for i := 0; i < 10; i++ {
		if !validator.IsSessionValid(context.Background(), "user-1", time.Now()) {
			t.Fatal("session should be valid when no cutoff is set")
		}
	}

	if source.callCount() != 1 {
		t.Errorf("source called %d times, want 1", source.callCount())
	}
}

func TestCachedSessionValidator_CachesPerUser(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, time.Minute)

	validator.IsSessionValid(context.Background(), "user-1", time.Now())
	validator.IsSessionValid(context.Background(), "user-2", time.Now())
	validator.IsSessionValid(context.Background(), "user-1", time.Now())

	if source.callCount() != 2 {
		t.Errorf("source called %d times, want 2 (one per user)", source.callCount())
	}
}

// The verdict depends on each token's issue time, so the cached cutoff must be
// re-evaluated per token rather than the answer being cached.
func TestCachedSessionValidator_VerdictFollowsTheToken(t *testing.T) {
	revokedAt := time.Now()
	source := &stubCutoffSource{cutoff: sql.NullTime{Time: revokedAt, Valid: true}}
	validator := NewCachedSessionValidator(source, time.Minute)

	ctx := context.Background()

	if validator.IsSessionValid(ctx, "user-1", revokedAt.Add(-time.Minute)) {
		t.Error("a token issued before the cutoff should be rejected")
	}

	if !validator.IsSessionValid(ctx, "user-1", revokedAt.Add(time.Minute)) {
		t.Error("a token issued after the cutoff should be accepted, from the same cached entry")
	}

	if source.callCount() != 1 {
		t.Errorf("source called %d times, want 1", source.callCount())
	}
}

func TestCachedSessionValidator_RefreshesAfterTTL(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, 20*time.Millisecond)

	ctx := context.Background()
	issuedAt := time.Now()

	if !validator.IsSessionValid(ctx, "user-1", issuedAt) {
		t.Fatal("session should start out valid")
	}

	// Revoke behind the cache's back, as another instance would.
	source.setCutoff(sql.NullTime{Time: time.Now().Add(time.Second), Valid: true})

	if !validator.IsSessionValid(ctx, "user-1", issuedAt) {
		t.Error("the cached entry should still be used before it expires")
	}

	time.Sleep(30 * time.Millisecond)

	if validator.IsSessionValid(ctx, "user-1", issuedAt) {
		t.Error("the revocation should be picked up once the entry expires")
	}
}

func TestCachedSessionValidator_ForgetForcesAReRead(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, time.Minute)

	ctx := context.Background()
	validator.IsSessionValid(ctx, "user-1", time.Now())
	validator.Forget("user-1")
	validator.IsSessionValid(ctx, "user-1", time.Now())

	if source.callCount() != 2 {
		t.Errorf("source called %d times, want 2 after Forget", source.callCount())
	}
}

// A failed lookup must deny access, and must not be cached as an answer.
func TestCachedSessionValidator_FailsClosedAndDoesNotCacheErrors(t *testing.T) {
	source := &stubCutoffSource{err: errors.New("database is down")}
	validator := NewCachedSessionValidator(source, time.Minute)

	ctx := context.Background()

	if validator.IsSessionValid(ctx, "user-1", time.Now()) {
		t.Error("a failed lookup should deny the session")
	}

	if validator.IsSessionValid(ctx, "user-1", time.Now()) {
		t.Error("a failed lookup should deny the session on retry too")
	}

	if source.callCount() != 2 {
		t.Errorf("source called %d times, want 2 - errors must not be cached", source.callCount())
	}
}

func TestCachedSessionValidator_BoundsCacheSize(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, time.Minute)

	for i := 0; i < maxCachedCutoffs+500; i++ {
		validator.IsSessionValid(context.Background(), string(rune(i%1000))+"-"+time.Now().Format("150405.000000000"), time.Now())
	}

	validator.mu.Lock()
	size := len(validator.entries)
	validator.mu.Unlock()

	if size > maxCachedCutoffs {
		t.Errorf("cache holds %d entries, want <= %d", size, maxCachedCutoffs)
	}
}

func TestCachedSessionValidator_ConcurrentUse(t *testing.T) {
	source := &stubCutoffSource{}
	validator := NewCachedSessionValidator(source, time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				validator.IsSessionValid(context.Background(), "user-1", time.Now())
			}
		}()
	}

	wg.Wait()
}
