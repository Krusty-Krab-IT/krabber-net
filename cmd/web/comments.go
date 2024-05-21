package web

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"krabber.net/internal/models"
	"krabber.net/internal/models/validator"
	"net/http"
	"time"
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
	c := &models.Comment{
		PK:      fmt.Sprintf("MC#%s", cu),                                           //  -> getCommentsForMolt()
		SK:      fmt.Sprintf("MC#%s", fmt.Sprintf(time.Now().Format(time.RFC3339))), // latest order
		GSI4PK:  fmt.Sprintf("MC#%s", m.ID),                                         // get comments on a molt id
		GSI4SK:  fmt.Sprintf("MC#%s", fmt.Sprintf(time.Now().Format(time.RFC3339))),
		Content: form.Comment,
	}

	err = app.Comments.Insert(c, m, cu)
	if err != nil {
		fmt.Println("error trying to add comment")
		app.serverError(w, r, err)
		return
	}
	app.SessionManager.Put(r.Context(), "flash", "Comment successfully created!")
	const view = "view.html"
	file := app.TemplateCache[view]
	if err != nil {
		fmt.Printf("Error commenting time: %v", err)
	}
	err = file.ExecuteTemplate(w, "comment-list-element", c)
	if err != nil {
		fmt.Printf("ERROR comment-list-element %v", err)
	}
}
