package web

import "net/http"

func (app *Application) crabTrench(w http.ResponseWriter, r *http.Request) {
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	trench, err := app.Trench.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	molts := app.Molts.GetTrenchMolts(trench)
	data := app.NewTemplateData(r)
	data.Molts = molts
	// access by index instead of copy which is default behaviour
	for i := range data.Molts {
		comments, err := app.Comments.On(data.Molts[i].ID) // add get string here
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		for _, c := range comments {
			data.Molts[i].Comments = append(data.Molts[i].Comments, c)
		}
	}

	app.Render(w, r, http.StatusOK, "trench.html", data)
}
