package web

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"krabber.net/internal/models/validator"
	"net/http"
)

type commentCreateForm struct {
	Comment             string `form:"comment"`
	validator.Validator `form:"-"`
}

func (app *Application) commentCreatePost(w http.ResponseWriter, r *http.Request) {
	var form commentCreateForm

	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Comment), "content", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	cu := app.SessionManager.GetString(r.Context(), "authenticatedCrabUserName")
	params := httprouter.ParamsFromContext(r.Context())
	mid := params.ByName("id")

	if mid == "" {
		app.NotFound(w)
		return
	}

	m, err := app.Molts.ByID(mid)
	if err != nil {
		app.NotFound(w)
		return
	}

	err = app.Comments.Insert(cu, form.Comment, m)
	if err != nil {
		fmt.Println("error trying to add comment")
		app.serverError(w, r, err)
		return
	}
}
