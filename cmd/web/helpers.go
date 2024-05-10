package web

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"net/http"
)

// Return true if the current request is from an authenticated crab, otherwise
// return false.
func (app *Application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

// Create an newTemplateData() helper, which returns a pointer to a templateData
// struct initialized with the current year. Note that we're not using the
// *http.Request parameter here at the moment, but we will do later in the book.
func (app *Application) NewTemplateData(r *http.Request) templateData {
	return templateData{
		Flash:           app.SessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CrabID:          app.SessionManager.GetString(r.Context(), "authenticatedCrabID"),
		CSRFToken:       nosurf.Token(r),
	}
}

// The background() helper accepts an arbitrary function as a parameter.
func (app *Application) background(fn func()) {
	// Increment the WaitGroup counter.
	app.Wg.Add(1)

	// Launch the background goroutine.
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.Wg.Done()

		defer func() {
			if err := recover(); err != nil {
				//app.Logger.Error("Error: ", err)
			}
		}()

		fn()
	}()
}

// The serverError helper writes a log entry at Error level (including the request
// method and URI as attributes), then sends a generic 500 Internal Server Error
// response to the crab.
func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	fmt.Println(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the crab. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the crab sent.
func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a NotFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the crab.
func (app *Application) NotFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *Application) Render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.TemplateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError() helper
	// and then return.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If the template is written to the buffer without any errors, we are safe
	// to go ahead and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(status)

	// Write the contents of the buffer to the http.ResponseWriter. Note: this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	buf.WriteTo(w)
}

// Create a new DecodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to Decode the form data into.
func (app *Application) DecodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// createMoltPost handler.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call Decode() on our Decoder instance, passing the target destination as
	// the first parameter.
	err = app.FormDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError.We use
		// errors.As() to check for this and raise a panic rather than returning
		// the error.
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// For all other errors, we return them as normal.
		return err
	}

	return nil
}
