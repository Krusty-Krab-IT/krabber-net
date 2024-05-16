package web

import "net/http"

func (app *Application) notifications(w http.ResponseWriter, r *http.Request) {
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	notifications, err := app.Notifications.Show(id) // get the crabs notifications
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.NewTemplateData(r)
	data.Notifications = notifications // copy stored
	// render the view
	// but before that update each notification to 'viewed' == true
	app.Notifications.MarkAsViewed(notifications) // permanent update so only show once... otherwise deleted in 7 days
	app.Render(w, r, http.StatusOK, "notifications.html", data)
}
