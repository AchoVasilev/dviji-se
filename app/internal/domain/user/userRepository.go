package user

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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

func (repo *UserRepository) Create(user User) error {
	ctx := context.Background()
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
		WHERE r.name = 'USER' AND r.is_deleted = FALSE`
	roles, err := tx.QueryContext(ctx, roleQuery)
	if err != nil {
		return err
	}

	defer roles.Close()

	var role Role
	role.Permissions = []Permission{}
	for roles.Next() {
		var perm Permission
		err := roles.Scan(
			&role.Id, &role.Name, &role.CreatedAt, &role.UpdatedAt, &role.UpdatedBy, &role.IsDeleted,
			&perm.Id, &perm.Name, &perm.CreatedAt, &perm.UpdatedAt, &perm.UpdatedBy, &perm.IsDeleted,
		)
		if err != nil {
			return err
		}

		role.Permissions = append(role.Permissions, perm)
	}

	rolesQuery := `INSERT INTO users_roles VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, rolesQuery, user.Id, role.Id)
	if err != nil {
		return err
	}

	var permissionsQuery strings.Builder
	permissionsQuery.WriteString("INSERT INTO users_permissions VALUES ")

	args := []any{}
	argPos := 1

	for i, perm := range role.Permissions {
		permissionsQuery.WriteString(fmt.Sprintf("($%d, $%d)", argPos, argPos+1))
		if i < len(role.Permissions)-1 {
			permissionsQuery.WriteString(", ")
		}
		args = append(args, user.Id, perm.Id)
		argPos += 2
	}

	permQueryString := permissionsQuery.String()
	_, err = tx.ExecContext(ctx, permQueryString, args...)
	if err != nil {
		slog.Info(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	rows, err := repo.db.QueryContext(ctx, `
	SELECT
		u.id, u.email, u.first_name, u.last_name, u.password, u.status, u.created_at, u.updated_at, u.is_deleted,
		r.id, r.name, r.created_at, r.updated_at, r.updated_by, r.is_deleted,
		p.id, p.name, p.created_at, p.updated_at, p.updated_by, p.is_deleted
	FROM users u
	JOIN users_roles ur ON u.id = ur.user_id
	JOIN roles r ON r.id = ur.role_id
	JOIN users_permissions up ON u.id = up.user_id
	JOIN permissions p ON p.id = up.permission_id
	WHERE u.email = $1 AND u.is_deleted = FALSE`, email)
	if err != nil {
		return User{}, err
	}

	defer rows.Close()

	var user User
	user.Roles = []Role{}
	user.Permissions = []Permission{}
	seenRoles := map[uuid.UUID]bool{}
	seenPerms := map[uuid.UUID]bool{}

	for rows.Next() {
		var (
			uid        uuid.UUID
			email      string
			firstName  sql.NullString
			lastName   sql.NullString
			password   string
			status     UserStatus
			createdAt  time.Time
			updatedAt  sql.NullTime
			isDeleted  bool
			role       Role
			permission Permission
		)

		err := rows.Scan(
			&uid, &email, &firstName, &lastName, &password, &status, &createdAt, &updatedAt, &isDeleted,
			&role.Id, &role.Name, &role.CreatedAt, &role.UpdatedAt, &role.UpdatedBy, &role.IsDeleted,
			&permission.Id, &permission.Name, &permission.CreatedAt, &permission.UpdatedAt, &permission.UpdatedBy, &permission.IsDeleted,
		)
		if err != nil {
			return User{}, err
		}

		if user.Id == uuid.Nil {
			user = User{
				Id:        uid,
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
				Password:  password,
				Status:    status,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				IsDeleted: isDeleted,
			}
		}

		if !seenRoles[role.Id] {
			user.Roles = append(user.Roles, role)
			seenRoles[role.Id] = true
		}

		if !seenPerms[permission.Id] {
			user.Permissions = append(user.Permissions, permission)
			seenPerms[permission.Id] = true
		}
	}

	if user.Id == uuid.Nil {
		return User{}, sql.ErrNoRows
	}

	return user, nil
}

func (repo *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_deleted = FALSE)`
	err := repo.db.QueryRowContext(ctx, query, email).Scan(&exists)

	return exists, err
}
