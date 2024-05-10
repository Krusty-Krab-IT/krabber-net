package web

import "net/http"

func (app *Application) moltinTime(w http.ResponseWriter, r *http.Request) {
	// for now show this logged in crabs molts
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	molts, err := app.Molts.Show(id) // add get string here
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.NewTemplateData(r)
	data.Molts = molts
	app.Render(w, r, http.StatusOK, "moltinTime.html", data)
}
