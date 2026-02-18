package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) (int, error)
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	GetUserInfo(id int) (*User, error)
	ComparePassword(id int, password string) error
	UpdatePassword(id int, password string) error
	GetRole(id int) (string, error)
	PromoteToLibrarian(id int) error
	ListUsers() ([]*User, error)
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Role           string
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, NOW())`

	result, err := m.DB.Exec(stmt, name, email, hashedPassword)
	if err != nil {
		// Checking if the error is for duplicate email
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users_uc_email") {
				return 0, ErrDuplicateEmail
			}
		}
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	// Check if the email exists in db
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) GetUserInfo(id int) (*User, error) {
	user := &User{}

	stmt := "SELECT id, name, email, created, role FROM users WHERE id = ?"

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return user, nil
}

func (m *UserModel) ComparePassword(id int, password string) error {
	stmt := `SELECT hashed_password FROM users WHERE id = ?`

	var hashedPassword []byte

	err := m.DB.QueryRow(stmt, id).Scan(&hashedPassword)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}
	return nil
}

func (m *UserModel) UpdatePassword(id int, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET hashed_password = ? WHERE id = ?`

	_, err = m.DB.Exec(stmt, hashedPassword, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) GetRole(id int) (string, error) {
	var role string
	stmt := "SELECT role FROM users WHERE id = ?"
	err := m.DB.QueryRow(stmt, id).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (m *UserModel) PromoteToLibrarian(id int) error {
	_, err := m.DB.Exec("UPDATE users SET role = 'librarian' WHERE id = ?", id)
	return err
}

func (m *UserModel) ListUsers() ([]*User, error) {
	rows, err := m.DB.Query("SELECT id, name, email, created, role FROM users ORDER BY created DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
