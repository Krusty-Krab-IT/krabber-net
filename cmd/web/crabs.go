package web

import (
	"errors"
	"fmt"
	"krabber.net/internal/models"
	"krabber.net/internal/models/validator"
	"net/http"
)

type crabActivateForm struct {
	Token               string `form:"token"`
	validator.Validator `form:"-"`
}

type crabLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type crabSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *Application) crabSignup(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = crabSignupForm{}
	app.Render(w, r, http.StatusOK, "signup.html", data)
}

func (app *Application) crabSignupPost(w http.ResponseWriter, r *http.Request) {
	// Declare zero-valued instance of our crabSignupForm struct.
	var form crabSignupForm

	// Parse the form data into the crabSignupForm struct.
	err := app.DecodePostForm(r, &form)
	if err != nil {
		fmt.Println("ERROR: ", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	// If there are any errors, redisplay the signup form along with a 422
	// status code.
	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}
	c := &models.Crab{
		Activated: false,
		Email:     form.Email,
		UserName:  form.Name,
	}
	err = c.Password.Set(form.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	crab, err := app.Crabs.Insert(c)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.NewTemplateData(r)
			data.Form = form
			app.Render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}
	token, err := app.Tokens.New(crab, models.ScopeActivation)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"crabID":          crab.ID,
		}

		// Send the welcome email, passing in the map above as dynamic models.
		err = app.Mailer.Send(crab.Email, "crab_welcome.html", data)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	})
	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked.
	app.SessionManager.Put(r.Context(), "flash", "Your signup was successful. Please check your e-mail.")

	// And redirect the crab to the login page.
	http.Redirect(w, r, "/crab/activate", http.StatusSeeOther)
}

func (app *Application) crabActivate(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = crabActivateForm{}
	app.Render(w, r, http.StatusOK, "activate.html", data)
}

func (app *Application) crabActivatePost(w http.ResponseWriter, r *http.Request) {
	var form crabActivateForm
	// Parse the form data into the crabSignupForm struct.
	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Token), "token", "This field cannot be blank")
	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "activate.html", data)
		return
	}
	t, err := app.Tokens.Get(models.ScopeActivation, form.Token)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.serverError(w, r, err)
		default:
			app.serverError(w, r, err)
		}
		return
	}
	crab, err := app.Crabs.ByToken(t)
	if err != nil {
		fmt.Printf("ERROR fetching crab by token %v", err)
	}
	crab.Activated = true
	err = app.Crabs.Activate(crab)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked.
	app.SessionManager.Put(r.Context(), "flash", "You've been activated. Please login to use the Krabber.net")

	// And redirect the crab to the login page.
	http.Redirect(w, r, "/crab/login", http.StatusSeeOther)

}

func (app *Application) crabLogin(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = crabLoginForm{}
	app.Render(w, r, http.StatusOK, "login.html", data)

}

func (app *Application) crabLoginPost(w http.ResponseWriter, r *http.Request) {
	// Decode the form data into the crabLoginForm struct.
	var form crabLoginForm

	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Do some validation checks on the form. We check that both email and
	// password are provided, and also check the format of the email address as
	// a UX-nicety (in case the crab makes a typo).
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// non-field error message and re-display the login page.
	crab, err := app.Crabs.ByEmail(form.Email)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.NewTemplateData(r)
			data.Form = form
			app.Render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	// Check if the provided password matches the actual password for the crab.
	match, err := models.Equal(form.Password, crab.PasswordHash)
	if !match {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.NewTemplateData(r)
			data.Form = form
			app.Render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	// add new token to db with 24 hr sessions
	_, err = app.Tokens.New(crab, models.ScopeAuthentication)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the crab (e.g. login
	// and logout operations).
	err = app.SessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	// Add the ID of the current crab to the session, so that they are now
	// 'logged in'.
	app.SessionManager.Put(r.Context(), "authenticatedCrabID", crab.ID)
	app.SessionManager.Put(r.Context(), "authenticatedCrabUserName", crab.UserName)
	app.SessionManager.Put(r.Context(), "authenticatedCrabEmail", crab.Email)
	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	// Redirect the crab to the create molt page.
	http.Redirect(w, r, "/moltinTime", http.StatusSeeOther)
}

func (app *Application) crabLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.SessionManager.RenewToken(r.Context())
	if err != nil {
		fmt.Printf("ERR %v", err)
		//app.serverError(w, r, err)
		//return
	}

	// Remove the authenticatedCrabID from the session data so that the crab is
	// 'logged out'.
	app.SessionManager.Remove(r.Context(), "authenticatedCrabID")

	// Add a flash message to the session to confirm to the crab that they've been
	// logged out.
	app.SessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	// Redirect the crab to the Application home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
