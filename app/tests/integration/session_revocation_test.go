package integration

import (
	"context"
	"testing"
	"time"

	"server/internal/application/users"
	"server/internal/domain/user"
	"server/tests/integration/testdb"
	"server/util/securityutil"

	"github.com/google/uuid"
)

// seedRevocationUser inserts a user directly so the test controls the id.
func seedRevocationUser(t *testing.T, tdb *testdb.TestDB, email string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := tdb.DB.Exec(
		`INSERT INTO users (id, email, password, status) VALUES ($1, $2, $3, 'Active')`,
		id, email, "irrelevant-hash")
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	return id
}

func TestSessionRevocation_PasswordChangeInvalidatesExistingTokens(t *testing.T) {
	tdb := testdb.SetupTestDB(t)
	defer tdb.CleanupTables(t)

	ctx := context.Background()
	repo := user.NewUserRepository(tdb.DB)
	service := users.NewUserService(repo)

	userId := seedRevocationUser(t, tdb, "revocation@example.com")

	// A token issued before any revocation is accepted.
	issuedAt := time.Now().UTC()
	if !service.IsSessionValid(ctx, userId.String(), issuedAt) {
		t.Fatal("a freshly issued token should be valid")
	}

	// Changing the password revokes everything issued up to that point.
	if err := repo.UpdatePassword(ctx, userId.String(), "new-hash"); err != nil {
		t.Fatalf("failed to update the password: %v", err)
	}

	if service.IsSessionValid(ctx, userId.String(), issuedAt) {
		t.Error("a token issued before the password change should be rejected")
	}

	// A token minted after the change works again, so the user can sign back in.
	if !service.IsSessionValid(ctx, userId.String(), time.Now().UTC().Add(time.Second)) {
		t.Error("a token issued after the password change should be accepted")
	}
}

func TestSessionRevocation_RevokeSessions(t *testing.T) {
	tdb := testdb.SetupTestDB(t)
	defer tdb.CleanupTables(t)

	ctx := context.Background()
	service := users.NewUserService(user.NewUserRepository(tdb.DB))

	userId := seedRevocationUser(t, tdb, "revoke-all@example.com")

	issuedAt := time.Now().UTC()
	if err := service.RevokeSessions(ctx, userId.String()); err != nil {
		t.Fatalf("failed to revoke sessions: %v", err)
	}

	if service.IsSessionValid(ctx, userId.String(), issuedAt) {
		t.Error("tokens issued before the revocation should be rejected")
	}
}

// An unknown user must not authenticate: the lookup fails, and the check is
// expected to fail closed rather than default to valid.
func TestSessionRevocation_UnknownUserFailsClosed(t *testing.T) {
	tdb := testdb.SetupTestDB(t)
	defer tdb.CleanupTables(t)

	service := users.NewUserService(user.NewUserRepository(tdb.DB))

	if service.IsSessionValid(context.Background(), uuid.New().String(), time.Now().UTC()) {
		t.Error("a token for a non-existent user should be rejected")
	}
}

// Refreshing rebuilds the access token from the database, so the roles it
// carries must survive the round trip - otherwise an admin silently loses
// their role on refresh.
func TestSessionRevocation_RefreshedTokenKeepsRoles(t *testing.T) {
	tdb := testdb.SetupTestDB(t)
	defer tdb.CleanupTables(t)

	ctx := context.Background()
	repo := user.NewUserRepository(tdb.DB)

	userId := seedRevocationUser(t, tdb, "roles@example.com")
	_, err := tdb.DB.Exec(
		`INSERT INTO users_roles (user_id, role_id) VALUES ($1, '22222222-2222-2222-2222-222222222222')`,
		userId)
	if err != nil {
		t.Fatalf("failed to grant the admin role: %v", err)
	}

	reloaded, err := repo.FindById(ctx, userId.String())
	if err != nil {
		t.Fatalf("failed to reload the user: %v", err)
	}

	token, _ := securityutil.GenerateAccessToken(reloaded, false)

	rebuilt, err := securityutil.UserFromToken(token)
	if err != nil {
		t.Fatalf("failed to parse the rebuilt token: %v", err)
	}

	var isAdmin bool
	for _, role := range rebuilt.Roles {
		if role.Name == "ADMIN" {
			isAdmin = true
		}
	}

	if !isAdmin {
		t.Errorf("the refreshed token lost the ADMIN role, got roles %+v", rebuilt.Roles)
	}

	if rebuilt.IssuedAt.IsZero() {
		t.Error("the token should carry an issued-at claim for revocation checks")
	}
}
