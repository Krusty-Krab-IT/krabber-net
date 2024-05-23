package web

import "net/http"

func (app *Application) profile(w http.ResponseWriter, r *http.Request) {
	// for now show this logged in crabs molts
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	// get crab by ID
	c, err := app.Crabs.ByID(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	molts, err := app.Molts.Show(id) // add get string here
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.NewTemplateData(r)
	data.Molts = molts
	data.Crab = c
	app.Render(w, r, http.StatusOK, "profile.html", data)
}
