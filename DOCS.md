# Library Management System — Prototype Docs

## What This Is

A basic library management system built with Go, chi, templ, MySQL, and SCS (session management). This is a prototype — it covers the core flows and nothing more.

---

## Roles

| Role | Who | What they can do |
|---|---|---|
| `student` | Default for every new signup | Browse & search books, issue books, return books |
| `librarian` | Promoted by admin | Everything a student can + add books, delete books, view all issue records |
| `admin` | Must be set manually in DB | Everything a librarian can + promote users to librarian |

**There is no "create admin" UI.** Admins are set directly in the database:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

---

## Database Setup

Run `schema.sql` against your MySQL instance to create all tables.

```bash
mysql -u library_user -p library < schema.sql
```

If the `users` table already exists (from before this prototype), add the role column manually:

```sql
ALTER TABLE users ADD COLUMN role ENUM('student', 'librarian', 'admin') NOT NULL DEFAULT 'student';
```

Tables created:
- `users` — stores user accounts with role
- `sessions` — SCS session store
- `books` — book catalogue with copy tracking
- `issues` — tracks which user has which book, with issue/due/return dates

---

## Running

```bash
go run ./cmd/web
```

Default port is `:4000`. Override with:

```bash
go run ./cmd/web -addr :8080
```

Default DSN: `library_user:eren@tcp(localhost:3306)/library?parseTime=true`

Override:

```bash
go run ./cmd/web -dsn "user:pass@tcp(host:port)/dbname?parseTime=true"
```

---

## URL Routes

| Method | Path | Access | Description |
|---|---|---|---|
| GET | `/` | All | Home page |
| GET | `/books` | All | Book catalogue + search |
| GET | `/user/signup` | Guest | Signup form |
| POST | `/user/signup` | Guest | Create account |
| GET | `/user/login` | Guest | Login form |
| POST | `/user/login` | Guest | Authenticate |
| POST | `/user/logout` | Authenticated | Logout |
| GET | `/my-books` | Authenticated | Books currently issued to me |
| POST | `/books/{id}/issue` | Authenticated | Issue a book |
| POST | `/issues/{id}/return` | Authenticated | Return a book |
| GET | `/books/new` | Librarian/Admin | Add book form |
| POST | `/books/new` | Librarian/Admin | Submit new book |
| POST | `/books/{id}/delete` | Librarian/Admin | Delete a book |
| GET | `/issues` | Librarian/Admin | All issue records |
| GET | `/admin/users` | Admin | User list + promote |
| POST | `/admin/users/{id}/promote` | Admin | Promote user to librarian |

---

## Project Structure

```
cmd/web/
  main.go        — app setup, DB connection, server start
  handlers.go    — all HTTP handlers
  routes.go      — chi router setup
  middleware.go  — auth, role, csrf, security headers
  helpers.go     — render, decode, isAuthenticated, getUserRole, etc.
  context.go     — context keys

internal/models/
  users.go       — user CRUD + role methods
  books.go       — book CRUD + availability tracking
  issues.go      — issue/return tracking
  errors.go      — sentinel errors

ui/
  efs.go         — embedded FS
  static/
    styles.css   — all CSS (book/library theme)
  html/
    base.templ   — shared layout with nav
    pages/
      home.templ
      login.templ
      signup.templ
      books.templ         — book catalogue + search
      book_form.templ     — add book form (librarian)
      mybooks.templ       — user's issued books
      issues.templ        — all issues (librarian view)
      admin_users.templ   — user management (admin)
```

---

## Book Issue Logic

- A book can only be issued if `available_copies > 0`
- A user cannot issue the same book twice (checked against active issues)
- Due date is set to **14 days** from issue date
- Returning a book increments `available_copies` back

---

## What's Missing (intentional, it's a prototype)

- No overdue notifications or fine system
- No pagination on book/issue lists
- No book editing (only add/delete)
- No profile/account management page (the existing password change methods are in the model but not wired to a UI — TODO)
- No admin demotion (promote only)
- No email verification

---

## Dependencies

All already in `go.mod`:

- `github.com/go-chi/chi/v5` — router
- `github.com/a-h/templ` — HTML templating
- `github.com/alexedwards/scs/v2` + `mysqlstore` — sessions
- `github.com/go-playground/form` — form decoding
- `github.com/go-sql-driver/mysql` — MySQL driver
- `github.com/justinas/nosurf` — CSRF protection
- `golang.org/x/crypto` — bcrypt
