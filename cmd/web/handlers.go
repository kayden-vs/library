package main

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/kayden-vs/library/internal/models"
	"github.com/kayden-vs/library/internal/validator"
	"github.com/kayden-vs/library/ui/html/pages"
)

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

	// Validate the form contents using our helper functions.
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

	// If there are any errors, redisplay the signup form along with a 422 status code.
	if !form.Valid() {
		app.RenderPage(w, r, func(flash string, isAuthenticated bool, csrfToken string) templ.Component {
			props.CSRFToken = csrfToken
			return pages.SignupPage(props, isAuthenticated)
		})
		return
	}

	var id int

	id, err = app.users.Insert(form.Name, form.Email, form.Password)
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

	app.sessionManager.Put(r.Context(), "flash", "Account created Succesfully.")

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
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

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out Succesfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
