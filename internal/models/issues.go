package models

import (
	"database/sql"
	"errors"
	"time"
)

type IssueModelInterface interface {
	Issue(bookID, userID int, dueDate time.Time) (int, error)
	Return(issueID, userID int) error
	GetActiveByUser(userID int) ([]*Issue, error)
	GetActiveByBook(bookID int) ([]*Issue, error)
	GetAll() ([]*Issue, error)
	GetActiveIssue(bookID, userID int) (*Issue, error)
}

type Issue struct {
	ID         int
	BookID     int
	UserID     int
	BookTitle  string
	UserName   string
	IssuedAt   time.Time
	DueDate    time.Time
	ReturnedAt *time.Time
}

type IssueModel struct {
	DB *sql.DB
}

func (m *IssueModel) Issue(bookID, userID int, dueDate time.Time) (int, error) {
	stmt := `INSERT INTO issues (book_id, user_id, issued_at, due_date) VALUES (?, ?, NOW(), ?)`
	result, err := m.DB.Exec(stmt, bookID, userID, dueDate)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (m *IssueModel) Return(issueID, userID int) error {
	result, err := m.DB.Exec(
		`UPDATE issues SET returned_at = NOW() WHERE id = ? AND user_id = ? AND returned_at IS NULL`,
		issueID, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("issue not found or already returned")
	}
	return nil
}

func (m *IssueModel) GetActiveByUser(userID int) ([]*Issue, error) {
	stmt := `SELECT i.id, i.book_id, i.user_id, b.title, u.name, i.issued_at, i.due_date, i.returned_at
             FROM issues i
             JOIN books b ON b.id = i.book_id
             JOIN users u ON u.id = i.user_id
             WHERE i.user_id = ? AND i.returned_at IS NULL
             ORDER BY i.issued_at DESC`
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows)
}

func (m *IssueModel) GetActiveByBook(bookID int) ([]*Issue, error) {
	stmt := `SELECT i.id, i.book_id, i.user_id, b.title, u.name, i.issued_at, i.due_date, i.returned_at
             FROM issues i
             JOIN books b ON b.id = i.book_id
             JOIN users u ON u.id = i.user_id
             WHERE i.book_id = ? AND i.returned_at IS NULL`
	rows, err := m.DB.Query(stmt, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows)
}

func (m *IssueModel) GetAll() ([]*Issue, error) {
	stmt := `SELECT i.id, i.book_id, i.user_id, b.title, u.name, i.issued_at, i.due_date, i.returned_at
             FROM issues i
             JOIN books b ON b.id = i.book_id
             JOIN users u ON u.id = i.user_id
             ORDER BY i.issued_at DESC`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows)
}

func (m *IssueModel) GetActiveIssue(bookID, userID int) (*Issue, error) {
	issue := &Issue{}
	stmt := `SELECT i.id, i.book_id, i.user_id, b.title, u.name, i.issued_at, i.due_date, i.returned_at
             FROM issues i
             JOIN books b ON b.id = i.book_id
             JOIN users u ON u.id = i.user_id
             WHERE i.book_id = ? AND i.user_id = ? AND i.returned_at IS NULL`
	err := m.DB.QueryRow(stmt, bookID, userID).Scan(
		&issue.ID, &issue.BookID, &issue.UserID, &issue.BookTitle, &issue.UserName,
		&issue.IssuedAt, &issue.DueDate, &issue.ReturnedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return issue, nil
}

func scanIssues(rows *sql.Rows) ([]*Issue, error) {
	var issues []*Issue
	for rows.Next() {
		i := &Issue{}
		err := rows.Scan(&i.ID, &i.BookID, &i.UserID, &i.BookTitle, &i.UserName, &i.IssuedAt, &i.DueDate, &i.ReturnedAt)
		if err != nil {
			return nil, err
		}
		issues = append(issues, i)
	}
	return issues, rows.Err()
}
