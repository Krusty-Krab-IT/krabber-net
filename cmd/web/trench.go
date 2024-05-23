package web

import (
	"fmt"
	"net/http"
)

func (app *Application) crabTrench(w http.ResponseWriter, r *http.Request) {
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}

	trench, err := app.Trench.Get(id)
	if err != nil {
		fmt.Printf("ERROR getting trench %v", trench)
		app.serverError(w, r, err)
		return
	}

	c, err := app.Crabs.ByID(id)
	if err != nil {
		fmt.Printf("ERROR getting crab %v", err)
		app.serverError(w, r, err)
		return
	}

	molts := app.Molts.GetTrenchMolts(trench)
	data := app.NewTemplateData(r)
	if molts != nil {
		data.Molts = molts
	}
	data.Crab = c
	if len(molts) > 0 {
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
	}

	app.Render(w, r, http.StatusOK, "trench.html", data)
}
