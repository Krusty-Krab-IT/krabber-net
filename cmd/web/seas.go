package web

import "net/http"

func (app *Application) sea(w http.ResponseWriter, r *http.Request) {
	molts, err := app.Molts.Sea()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

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

	app.Render(w, r, http.StatusOK, "sea.html", data)
}
