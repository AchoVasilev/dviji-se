package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create adds a user with the default USER role.
func (repo *UserRepository) Create(ctx context.Context, user User) error {
	return repo.createWithRole(ctx, user, "USER")
}

// CreateAdmin adds a user with the ADMIN role. Used to bootstrap the first
// administrator, since registration only ever grants USER.
func (repo *UserRepository) CreateAdmin(ctx context.Context, user User) error {
	return repo.createWithRole(ctx, user, "ADMIN")
}

func (repo *UserRepository) createWithRole(ctx context.Context, user User, roleName string) error {
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `INSERT INTO users (id, email, password, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, query, user.Id, user.Email, user.Password, user.Status, user.CreatedAt)
	if err != nil {
		return err
	}

	roleQuery := `
		SELECT
			r.id, r.name, r.created_at, r.updated_at, r.updated_by, r.is_deleted,
			p.id, p.name, p.created_at, p.updated_at, p.updated_by, p.is_deleted
		FROM roles r
		JOIN roles_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE r.name = $1 AND r.is_deleted = FALSE`
	roles, err := tx.QueryContext(ctx, roleQuery, roleName)
	if err != nil {
		return err
	}

	defer roles.Close()

	var role Role
	role.Permissions = []Permission{}
	for roles.Next() {
		var perm Permission
		// Assign to the outer err so the deferred rollback sees the failure.
		err = roles.Scan(
			&role.Id, &role.Name, &role.CreatedAt, &role.UpdatedAt, &role.UpdatedBy, &role.IsDeleted,
			&perm.Id, &perm.Name, &perm.CreatedAt, &perm.UpdatedAt, &perm.UpdatedBy, &perm.IsDeleted,
		)
		if err != nil {
			return err
		}

		role.Permissions = append(role.Permissions, perm)
	}

	if err = roles.Err(); err != nil {
		return err
	}

	if role.Id == uuid.Nil {
		err = fmt.Errorf("role %q not found, cannot create user", roleName)
		return err
	}

	rolesQuery := `INSERT INTO users_roles (user_id, role_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, rolesQuery, user.Id, role.Id)
	if err != nil {
		return err
	}

	// A role with no permissions is legitimate; skip the insert rather than
	// building "INSERT INTO users_permissions VALUES " with no rows.
	if len(role.Permissions) > 0 {
		var permissionsQuery strings.Builder
		permissionsQuery.WriteString("INSERT INTO users_permissions (user_id, permission_id) VALUES ")

		args := make([]any, 0, len(role.Permissions)*2)
		argPos := 1

		for i, perm := range role.Permissions {
			if i > 0 {
				permissionsQuery.WriteString(", ")
			}

			fmt.Fprintf(&permissionsQuery, "($%d, $%d)", argPos, argPos+1)
			args = append(args, user.Id, perm.Id)
			argPos += 2
		}

		_, err = tx.ExecContext(ctx, permissionsQuery.String(), args...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	// 1. Get user
	var user User
	var firstName, lastName sql.NullString
	var updatedAt sql.NullTime

	err := repo.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, password, status, created_at, updated_at, is_deleted, tokens_valid_after
		FROM users
		WHERE email = $1 AND is_deleted = FALSE`, email).Scan(
		&user.Id, &user.Email, &firstName, &lastName, &user.Password,
		&user.Status, &user.CreatedAt, &updatedAt, &user.IsDeleted, &user.TokensValidAfter,
	)
	if err != nil {
		return User{}, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = updatedAt

	if err := repo.loadRolesAndPermissions(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

// loadRolesAndPermissions fills in the user's roles and permissions. Shared so
// that every path which builds an access token sees the same authority.
func (repo *UserRepository) loadRolesAndPermissions(ctx context.Context, user *User) error {
	roleRows, err := repo.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.created_at, r.updated_at, r.updated_by, r.is_deleted
		FROM roles r
		JOIN users_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_deleted = FALSE`, user.Id)
	if err != nil {
		return err
	}
	defer roleRows.Close()

	user.Roles = []Role{}
	for roleRows.Next() {
		var role Role
		if scanErr := roleRows.Scan(&role.Id, &role.Name, &role.CreatedAt, &role.UpdatedAt, &role.UpdatedBy, &role.IsDeleted); scanErr != nil {
			return scanErr
		}
		user.Roles = append(user.Roles, role)
	}

	if rowsErr := roleRows.Err(); rowsErr != nil {
		return rowsErr
	}

	permRows, err := repo.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.created_at, p.updated_at, p.updated_by, p.is_deleted
		FROM permissions p
		JOIN users_permissions up ON p.id = up.permission_id
		WHERE up.user_id = $1 AND p.is_deleted = FALSE`, user.Id)
	if err != nil {
		return err
	}
	defer permRows.Close()

	user.Permissions = []Permission{}
	for permRows.Next() {
		var perm Permission
		if scanErr := permRows.Scan(&perm.Id, &perm.Name, &perm.CreatedAt, &perm.UpdatedAt, &perm.UpdatedBy, &perm.IsDeleted); scanErr != nil {
			return scanErr
		}
		user.Permissions = append(user.Permissions, perm)
	}

	return permRows.Err()
}

func (repo *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_deleted = FALSE)`
	err := repo.db.QueryRowContext(ctx, query, email).Scan(&exists)

	return exists, err
}

// UpdatePassword changes the password and revokes tokens issued before now, so
// a password change ends sessions that were already open.
func (repo *UserRepository) UpdatePassword(ctx context.Context, userId string, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = NOW(), tokens_valid_after = NOW()
		WHERE id = $2 AND is_deleted = FALSE`
	_, err := repo.db.ExecContext(ctx, query, hashedPassword, userId)

	return err
}

// FindById loads the user together with roles and permissions. Refreshing a
// session rebuilds the access token from this, so omitting roles here would
// silently strip an admin of their role on the next refresh.
func (repo *UserRepository) FindById(ctx context.Context, userId string) (User, error) {
	var user User
	var firstName, lastName sql.NullString
	var updatedAt sql.NullTime

	err := repo.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, password, status, created_at, updated_at, is_deleted, tokens_valid_after
		FROM users
		WHERE id = $1 AND is_deleted = FALSE`, userId).Scan(
		&user.Id, &user.Email, &firstName, &lastName, &user.Password,
		&user.Status, &user.CreatedAt, &updatedAt, &user.IsDeleted, &user.TokensValidAfter,
	)
	if err != nil {
		return User{}, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = updatedAt

	if err := repo.loadRolesAndPermissions(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

// AdminExists reports whether any administrator account is present.
func (repo *UserRepository) AdminExists(ctx context.Context) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users u
			JOIN users_roles ur ON u.id = ur.user_id
			JOIN roles r ON r.id = ur.role_id
			WHERE r.name = 'ADMIN' AND u.is_deleted = FALSE AND r.is_deleted = FALSE
		)`).Scan(&exists)

	return exists, err
}

// RevokeTokensIssuedBefore stops every access token minted at or before the
// given instant from authenticating.
func (repo *UserRepository) RevokeTokensIssuedBefore(ctx context.Context, userId string, cutoff time.Time) error {
	query := `UPDATE users SET tokens_valid_after = $1, updated_at = NOW() WHERE id = $2`
	_, err := repo.db.ExecContext(ctx, query, cutoff.UTC(), userId)

	return err
}

// TokensValidAfter returns the revocation cutoff for a user, if any.
func (repo *UserRepository) TokensValidAfter(ctx context.Context, userId string) (sql.NullTime, error) {
	var validAfter sql.NullTime
	err := repo.db.QueryRowContext(ctx,
		`SELECT tokens_valid_after FROM users WHERE id = $1 AND is_deleted = FALSE`, userId).Scan(&validAfter)

	return validAfter, err
}
