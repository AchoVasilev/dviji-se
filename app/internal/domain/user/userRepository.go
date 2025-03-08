package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	query := `INSERT INTO users (id, email, password, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := repo.db.Exec(query, user.Id, user.Email, user.Password, user.Status, user.CreatedAt)
	if err != nil {
		return err
	}

	roleQuery := `SELECT * FROM roles r
	JOIN roles_permissions rp ON r.id = rp.role_id
	JOIN permissions p ON p.id = rp.permission_id
	WHERE name = 'USER'`
	roles, err := repo.db.Query(roleQuery)
	if err != nil {
		return err
	}

	defer roles.Close()
	role := new(Role)
	for roles.Next() {
		err := roles.Scan(
			&role.Id,
			&role.Name,
			&role.Permissions,
		)

		if err != nil {
			return err
		}
	}

	rolesQuery := `INSERT INTO users_roles VALUES ($1, $2)`
	_, err = repo.db.Exec(rolesQuery, user.Id, role.Id)
	if err != nil {
		return err
	}

	var permissionsQuery strings.Builder
	permissionsQuery.WriteString("INSERT INTO users_roles VALUES ")

	for i, perm := range role.Permissions {
		if i == len(role.Permissions) {
			permissionsQuery.WriteString(fmt.Sprintf("(%v, %v);", user.Id, perm.Id))
			break
		}

		permissionsQuery.WriteString(fmt.Sprintf("(%v, %v),", user.Id, perm.Id))
	}

	_, err = repo.db.Exec(permissionsQuery.String())

	return err
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT * FROM users u 
		JOIN users_roles ur ON u.id = ur.user_id
		JOIN roles r ON r.id = ur.role_id
		JOIN users_permissions up ON u.id = up.user_id
		JOIN permissions p ON p.id = up.permission_id
		WHERE u.email = ? AND u.is_deleted = FALSE`, email)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	user := new(User)
	for rows.Next() {
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Status,
			&user.IsDeleted,
			&user.Roles,
			&user.Permissions,
		)

		if err != nil {
			return nil, err
		}
	}

	return user, nil
}
