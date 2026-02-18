package models

import (
	"database/sql"
	"errors"
	"time"
)

type BookModelInterface interface {
	Insert(title, author, isbn string, totalCopies int) (int, error)
	Get(id int) (*Book, error)
	Delete(id int) error
	Search(query string) ([]*Book, error)
	List() ([]*Book, error)
	DecrementAvailable(id int) error
	IncrementAvailable(id int) error
}

type Book struct {
	ID              int
	Title           string
	Author          string
	ISBN            string
	TotalCopies     int
	AvailableCopies int
	Created         time.Time
}

type BookModel struct {
	DB *sql.DB
}

func (m *BookModel) Insert(title, author, isbn string, totalCopies int) (int, error) {
	stmt := `INSERT INTO books (title, author, isbn, total_copies, available_copies, created)
             VALUES (?, ?, ?, ?, ?, NOW())`
	result, err := m.DB.Exec(stmt, title, author, isbn, totalCopies, totalCopies)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (m *BookModel) Get(id int) (*Book, error) {
	b := &Book{}
	stmt := `SELECT id, title, author, isbn, total_copies, available_copies, created FROM books WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.TotalCopies, &b.AvailableCopies, &b.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return b, nil
}

func (m *BookModel) Delete(id int) error {
	// remove issues first to satisfy FK constraint
	_, err := m.DB.Exec("DELETE FROM issues WHERE book_id = ?", id)
	if err != nil {
		return err
	}
	_, err = m.DB.Exec("DELETE FROM books WHERE id = ?", id)
	return err
}

func (m *BookModel) Search(query string) ([]*Book, error) {
	like := "%" + query + "%"
	stmt := `SELECT id, title, author, isbn, total_copies, available_copies, created
             FROM books WHERE title LIKE ? OR author LIKE ? OR isbn LIKE ?
             ORDER BY title`
	rows, err := m.DB.Query(stmt, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBooks(rows)
}

func (m *BookModel) List() ([]*Book, error) {
	rows, err := m.DB.Query(`SELECT id, title, author, isbn, total_copies, available_copies, created FROM books ORDER BY title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBooks(rows)
}

func (m *BookModel) DecrementAvailable(id int) error {
	_, err := m.DB.Exec("UPDATE books SET available_copies = available_copies - 1 WHERE id = ? AND available_copies > 0", id)
	return err
}

func (m *BookModel) IncrementAvailable(id int) error {
	_, err := m.DB.Exec("UPDATE books SET available_copies = available_copies + 1 WHERE id = ?", id)
	return err
}

func scanBooks(rows *sql.Rows) ([]*Book, error) {
	var books []*Book
	for rows.Next() {
		b := &Book{}
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.TotalCopies, &b.AvailableCopies, &b.Created)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}
