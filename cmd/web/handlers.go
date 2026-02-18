package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/kayden-vs/library/internal/models"
	"github.com/kayden-vs/library/internal/validator"
	"github.com/kayden-vs/library/ui/html/pages"
)

// --- auth forms ---

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	props := pages.SignupFormParams{}
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		props.CSRFToken = csrfToken
		return pages.SignupPage(props, isAuthenticated)
	})
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	props := pages.SignupFormParams{
		Name:        form.Name,
		Email:       form.Email,
		FieldErrors: form.FieldErrors,
	}

	if !form.Valid() {
		app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
			props.CSRFToken = csrfToken
			return pages.SignupPage(props, isAuthenticated)
		})
		return
	}

	id, err := app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			props.FieldErrors = form.FieldErrors
			app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
				props.CSRFToken = csrfToken
				return pages.SignupPage(props, isAuthenticated)
			})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Account created successfully.")
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	props := pages.LoginFormParams{}
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		props.CSRFToken = csrfToken
		return pages.LoginPage(props, flash, isAuthenticated)
	})
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	props := pages.LoginFormParams{
		Email:          form.Email,
		FieldErrors:    form.FieldErrors,
		NonFieldErrors: form.NonFieldErrors,
	}
	if !form.Valid() {
		app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
			props.CSRFToken = csrfToken
			return pages.LoginPage(props, flash, isAuthenticated)
		})
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			props.NonFieldErrors = form.NonFieldErrors
			app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
				props.CSRFToken = csrfToken
				return pages.LoginPage(props, flash, isAuthenticated)
			})
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- home ---

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.HomePage(flash, isAuthenticated)
	})
}

// --- books ---

func (app *application) bookList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var books []*models.Book
	var err error

	if query != "" {
		books, err = app.books.Search(query)
	} else {
		books, err = app.books.List()
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	isLibrarian := app.isLibrarian(r)
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.BookListPage(books, query, flash, isAuthenticated, isLibrarian, csrfToken)
	})
}

type bookForm struct {
	Title               string `form:"title"`
	Author              string `form:"author"`
	ISBN                string `form:"isbn"`
	Copies              string `form:"copies"`
	validator.Validator `form:"-"`
}

func (app *application) bookCreateForm(w http.ResponseWriter, r *http.Request) {
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.BookFormPage(pages.BookFormParams{CSRFToken: csrfToken}, flash, isAuthenticated)
	})
}

func (app *application) bookCreatePost(w http.ResponseWriter, r *http.Request) {
	var form bookForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "Title cannot be blank")
	form.CheckField(validator.NotBlank(form.Author), "author", "Author cannot be blank")
	form.CheckField(validator.NotBlank(form.ISBN), "isbn", "ISBN cannot be blank")

	copies, err := strconv.Atoi(form.Copies)
	if err != nil || copies < 1 {
		form.AddFieldError("copies", "Must be a number >= 1")
	}

	props := pages.BookFormParams{
		Title:       form.Title,
		Author:      form.Author,
		ISBN:        form.ISBN,
		Copies:      form.Copies,
		FieldErrors: form.FieldErrors,
	}

	if !form.Valid() {
		app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
			props.CSRFToken = csrfToken
			return pages.BookFormPage(props, flash, isAuthenticated)
		})
		return
	}

	_, err = app.books.Insert(form.Title, form.Author, form.ISBN, copies)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book added successfully.")
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

func (app *application) bookDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.notFound(w)
		return
	}
	err = app.books.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Book removed.")
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

// --- issues ---

func (app *application) issueBookPost(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.notFound(w)
		return
	}

	book, err := app.books.Get(bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if book.AvailableCopies < 1 {
		app.sessionManager.Put(r.Context(), "flash", "No copies available right now.")
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	userID := app.getUserID(r)
	_, err = app.issues.GetActiveIssue(bookID, userID)
	if err == nil {
		// already has active issue
		app.sessionManager.Put(r.Context(), "flash", "You already have this book issued.")
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	dueDate := time.Now().AddDate(0, 0, 14) // 2 week due
	_, err = app.issues.Issue(bookID, userID, dueDate)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.books.DecrementAvailable(bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book issued! Due: "+dueDate.Format("02 Jan 2006"))
	http.Redirect(w, r, "/my-books", http.StatusSeeOther)
}

func (app *application) returnBookPost(w http.ResponseWriter, r *http.Request) {
	issueID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.notFound(w)
		return
	}

	userID := app.getUserID(r)
	// get the issue first so we can increment available copies
	activeIssues, err := app.issues.GetActiveByUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var bookID int
	for _, iss := range activeIssues {
		if iss.ID == issueID {
			bookID = iss.BookID
			break
		}
	}

	if bookID == 0 {
		app.clientError(w, http.StatusForbidden)
		return
	}

	err = app.issues.Return(issueID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.books.IncrementAvailable(bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book returned successfully.")
	http.Redirect(w, r, "/my-books", http.StatusSeeOther)
}

func (app *application) myBooks(w http.ResponseWriter, r *http.Request) {
	userID := app.getUserID(r)
	issues, err := app.issues.GetActiveByUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.MyBooksPage(issues, flash, isAuthenticated, csrfToken)
	})
}

// --- admin ---

func (app *application) adminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.ListUsers()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.AdminUsersPage(users, flash, isAuthenticated, csrfToken)
	})
}

func (app *application) adminPromotePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.notFound(w)
		return
	}
	err = app.users.PromoteToLibrarian(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "User promoted to librarian.")
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// -- librarian issue management --

func (app *application) allIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := app.issues.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
		return pages.AllIssuesPage(issues, flash, isAuthenticated)
	})
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
