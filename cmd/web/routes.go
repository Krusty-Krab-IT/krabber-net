package web

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"krabber.net/public"
	"net/http"
)

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux.
func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.NotFound(w)
	})

	// Leave the static files route unchanged.
	fileServer := http.FileServer(http.FS(public.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	dynamic := alice.New(app.SessionManager.LoadAndSave, noSurf, app.authenticate)

	// SEA
	router.Handler(http.MethodGet, "/sea", dynamic.ThenFunc(app.sea))

	// TRENCH
	router.Handler(http.MethodGet, "/trench", dynamic.ThenFunc(app.crabTrench))

	// ROOT
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.root))
	router.Handler(http.MethodPost, "/", dynamic.ThenFunc(app.moltCommonActionsPost))

	// FOLLOW
	router.Handler(http.MethodPost, "/follow/:id", dynamic.ThenFunc(app.followCreatePost))
	router.Handler(http.MethodPost, "/unfollow/:id", dynamic.ThenFunc(app.followDeletePost))

	// PROFILE
	router.Handler(http.MethodGet, "/moltinTime", dynamic.ThenFunc(app.moltinTime))
	router.Handler(http.MethodGet, "/notifications", dynamic.ThenFunc(app.notifications))
	router.Handler(http.MethodGet, "/settings", dynamic.ThenFunc(app.settings))

	// COMMENT
	router.Handler(http.MethodPost, "/comment/:id", dynamic.ThenFunc(app.commentCreatePost))

	// MOLTS
	router.Handler(http.MethodGet, "/molt/view/:id", dynamic.ThenFunc(app.moltView))
	router.Handler(http.MethodPost, "/molt/like/:id", dynamic.ThenFunc(app.moltLikePost))
	router.Handler(http.MethodPost, "/remolt/:id", dynamic.ThenFunc(app.moltRemoltPost))

	// LIKES
	router.Handler(http.MethodGet, "/molt/likes/view/:id", dynamic.ThenFunc(app.moltLikesView))

	// CRABMIN
	router.Handler(http.MethodGet, "/crabmin", dynamic.ThenFunc(app.crabmin))
	router.Handler(http.MethodPost, "/crabmin/sea", dynamic.ThenFunc(app.crabminCreateSea))

	// CRAB
	router.Handler(http.MethodGet, "/crab/signup", dynamic.ThenFunc(app.crabSignup))
	router.Handler(http.MethodGet, "/crabs", dynamic.ThenFunc(app.allCrabs))
	router.Handler(http.MethodPost, "/crab/signup", dynamic.ThenFunc(app.crabSignupPost))
	router.Handler(http.MethodGet, "/crab/login", dynamic.ThenFunc(app.crabLogin))
	router.Handler(http.MethodPost, "/crab/login", dynamic.ThenFunc(app.crabLoginPost))
	router.Handler(http.MethodGet, "/crab/activate", dynamic.ThenFunc(app.crabActivate))
	router.Handler(http.MethodPost, "/crab/activate", dynamic.ThenFunc(app.crabActivatePost))

	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodPost, "/molt/create", protected.ThenFunc(app.moltCreatePost))
	router.Handler(http.MethodPost, "/crab/logout", protected.ThenFunc(app.crabLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
