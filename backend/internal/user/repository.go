package user

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SQLRepository struct {
	db *sqlx.DB
}

func NewSQLRepository(db *sqlx.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, u User) (User, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, email, display_name, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?, ?)`,
		u.Username, nil, u.DisplayName, u.PasswordHash, u.Role, u.Status)
	if isDuplicate(err) {
		return User{}, ErrDuplicateUsername
	}
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	u.ID = id
	return u, nil
}

func (r *SQLRepository) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `
		SELECT id, username, password_hash, role, COALESCE(display_name, '') AS display_name, status
		FROM users WHERE id = ?`, id)
	return u, err
}

func (r *SQLRepository) FindByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `
		SELECT id, username, password_hash, role, COALESCE(display_name, '') AS display_name, status
		FROM users WHERE username = ?`, username)
	return u, err
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
