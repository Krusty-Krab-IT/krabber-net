package web

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"krabber.net/internal/models"
	"krabber.net/internal/models/ksuid"
	"krabber.net/internal/models/validator"
	"net/http"
	"time"
)

type moltCreateForm struct {
	Content             string `form:"content"`
	validator.Validator `form:"-"`
}

func (app *Application) moltCommonActionsPost(w http.ResponseWriter, r *http.Request) {
	var form moltCreateForm

	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}
	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	KSUID := ksuid.GenerateKSUID()
	id := uuid.New().String()

	now := time.Now()
	y, mnth, d := now.Date()

	molt := &models.Molt{
		ID:      id,
		PK:      fmt.Sprintf("M#%s", crabID),
		SK:      fmt.Sprintf("M#%s#%s", crabID, KSUID),
		GSI3PK:  "M#" + fmt.Sprintf("%d-%d-%d", y, int(mnth), d), //fmt.Sprintf("M#%s", time.Now().Format(time.RFC3339))
		GSI3SK:  fmt.Sprintf("M#%s", id),
		GSI5PK:  fmt.Sprintf("M#%s", id),
		GSI5SK:  fmt.Sprintf("M#%s", id),
		Author:  crabID,
		Deleted: false,
		Content: form.Content,
	}

	res := app.Molts.Insert(molt)
	if res != nil {
		fmt.Println("error after inserting molt..?")
		app.serverError(w, r, err)
		return
	}
	// if err == nil then get Followers
	// then for each follower do writing to their trench
	Followers := app.Follows.Followers(crabID)
	if Followers != nil {
		err := app.Trench.Insert(Followers, molt)
		fmt.Println("Alerting followers of my molt...")
		if err != nil {
			app.serverError(w, r, err)
		}
	}
	app.SessionManager.Put(r.Context(), "flash", "Molt successfully created!")
	tmpl := template.Must(template.ParseFiles("public/html/pages/profile.html")) // TODO remove this long af thing
	tmpl.ExecuteTemplate(w, "molt-list-element", molt)
}

func (app *Application) moltLikePost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	fmt.Println("liking", id)
	if id == "" {
		app.NotFound(w)
		return
	}
	molt, err := app.Molts.ByID(id)
	if err != nil {
		app.NotFound(w)
		return
	}

	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")

	res := app.Likes.Insert(crabID, molt)
	if res != nil {
		fmt.Println("error creating a like on the molt")
		app.serverError(w, r, err)
		return
	}
	app.SessionManager.Put(r.Context(), "flash", "Like successfully created!")

}

func (app *Application) moltRemoltPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	fmt.Println("remolting", id)
	if id == "" {
		app.NotFound(w)
		return
	}
	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	crab, err := app.Crabs.ByID(crabID)
	if err != nil {
		app.NotFound(w)
		return
	}
	oldMolt, err := app.Molts.ByID(id)
	if err != nil {
		app.NotFound(w)
		return
	}
	KSUID := ksuid.GenerateKSUID()
	mid := uuid.New().String()
	newMolt := &models.Molt{
		ID:      mid,
		PK:      fmt.Sprintf("M#%s", crabID), // it is a new molt for the crab doing it
		SK:      fmt.Sprintf("M#%s#%s", crabID, KSUID),
		GSI3PK:  fmt.Sprintf("M#%s", time.Now().Format(time.RFC3339)),
		GSI3SK:  fmt.Sprintf("M#%s", mid),
		GSI5PK:  fmt.Sprintf("M#%s", mid),
		GSI5SK:  fmt.Sprintf("M#%s", mid),
		Author:  oldMolt.Author,  // original author
		Content: oldMolt.Content, // original content
	}

	res := app.Molts.ReMolt(crab, oldMolt, newMolt)
	if res != nil {
		fmt.Println("error creating a remolt on the molt")
		app.serverError(w, r, err)
		return
	}
	app.SessionManager.Put(r.Context(), "flash", "Remolt successfully created!")

}

func (app *Application) moltView(w http.ResponseWriter, r *http.Request) {
	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	c, err := app.Crabs.ByID(crabID)

	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	if id == "" {
		app.NotFound(w)
		return
	}

	molt, err := app.Molts.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.NotFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.NewTemplateData(r)
	data.Molt = *molt
	data.Crab = c
	comments, err := app.Comments.On(data.Molt.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	for _, c := range comments {
		data.Molt.Comments = append(data.Molt.Comments, c)
	}
	app.Render(w, r, http.StatusOK, "view.html", data)
}

func (app *Application) moltLikesView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	if id == "" {
		app.NotFound(w)
		return
	}
	data := app.NewTemplateData(r)
	likes, err := app.Likes.On(id) // add get string here
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	for _, l := range likes {
		data.Likes = append(data.Likes, l)
	}

	fmt.Printf("data%+v: ", data)
	app.Render(w, r, http.StatusOK, "likes.html", data)
}

func (app *Application) moltCreatePost(w http.ResponseWriter, r *http.Request) {
	var form moltCreateForm

	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "create.html", data) // redirect oops page?
		return
	}
	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	author := app.SessionManager.GetString(r.Context(), "authenticatedCrabUserName")
	KSUID := ksuid.GenerateKSUID()
	id := uuid.New().String()
	c, err := app.Crabs.ByID(crabID)
	fmt.Printf("HERE's CRAB BY ID : %v", c)
	if err != nil {
		fmt.Printf("ERROR retrieving crab for molt %v", err)
	}
	now := time.Now()
	y, mnth, d := now.Date()

	molt := &models.Molt{
		ID:            id,
		PK:            fmt.Sprintf("M#%s", crabID),
		SK:            fmt.Sprintf("M#%s#%s", crabID, KSUID),
		GSI3PK:        "M#" + fmt.Sprintf("%d-%d-%d", y, int(mnth), d), //fmt.Sprintf("M#%s", time.Now().Format(time.RFC3339))
		GSI3SK:        fmt.Sprintf("M#%s", id),
		GSI5PK:        fmt.Sprintf("M#%s", id),
		GSI5SK:        fmt.Sprintf("M#%s", id),
		Author:        author,
		CreatorAvatar: c.Avatar,
		Deleted:       false,
		Content:       form.Content,
	}

	res := app.Molts.Insert(molt)
	if res != nil {
		fmt.Println("error after inserting molt..?")
		app.serverError(w, r, err)
		return
	}
	Followers := app.Follows.Followers(crabID)
	if Followers != nil {
		err := app.Trench.Insert(Followers, molt)
		fmt.Println("Alerting followers of my molt...")
		if err != nil {
			app.serverError(w, r, err)
		}
	}
	app.SessionManager.Put(r.Context(), "flash", "Molt successfully created!")
	const p = "profile.html"
	file := app.TemplateCache[p]
	if err != nil {
		fmt.Printf("Error moltin time: %v", err)
	}
	err = file.ExecuteTemplate(w, "molt-list-element", molt)
	if err != nil {
		fmt.Printf("ERROR molt-list-element %v", err)
	}
}

func (app *Application) moltModalCreatePost(w http.ResponseWriter, r *http.Request) {
	var form moltCreateForm

	err := app.DecodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		app.Render(w, r, http.StatusUnprocessableEntity, "create.html", data) // redirect oops page?
		return
	}
	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	author := app.SessionManager.GetString(r.Context(), "authenticatedCrabUserName")
	KSUID := ksuid.GenerateKSUID()
	id := uuid.New().String()
	c, err := app.Crabs.ByID(crabID)
	fmt.Printf("HERE's CRAB BY ID : %v", c)
	if err != nil {
		fmt.Printf("ERROR retrieving crab for molt %v", err)
	}

	now := time.Now()
	y, mnth, d := now.Date()

	molt := &models.Molt{
		ID:            id,
		PK:            fmt.Sprintf("M#%s", crabID),
		SK:            fmt.Sprintf("M#%s#%s", crabID, KSUID),
		GSI3PK:        "M#" + fmt.Sprintf("%d-%d-%d", y, int(mnth), d), //fmt.Sprintf("M#%s", time.Now().Format(time.RFC3339))
		GSI3SK:        fmt.Sprintf("M#%s", id),
		GSI5PK:        fmt.Sprintf("M#%s", id),
		GSI5SK:        fmt.Sprintf("M#%s", id),
		Author:        author,
		CreatorAvatar: c.Avatar,
		Deleted:       false,
		Content:       form.Content,
	}

	res := app.Molts.Insert(molt)
	if res != nil {
		fmt.Println("error after inserting molt..?")
		app.serverError(w, r, err)
		return
	}
	Followers := app.Follows.Followers(crabID)
	if Followers != nil {
		err := app.Trench.Insert(Followers, molt)
		fmt.Println("Alerting followers of my molt...")
		if err != nil {
			app.serverError(w, r, err)
		}
	}
	app.SessionManager.Put(r.Context(), "flash", "Molt successfully created!")
	const moltinTime = "nav.html"
	file := app.TemplateCache[moltinTime]
	if err != nil {
		fmt.Printf("Error moltin time: %v", err)
	}
	err = file.ExecuteTemplate(w, "modal-nav", "")
	if err != nil {
		fmt.Printf("ERROR exampleModal %v", err)
	}
}
