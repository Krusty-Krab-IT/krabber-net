package web

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *Application) followCreatePost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	fmt.Print("Params: ", params)
	id := params.ByName("id")

	fmt.Println("following", id)
	if id == "" {
		app.NotFound(w)
		return
	}

	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	follower, err := app.Crabs.ByID(crabID)
	//fmt.Println("FOLLOWER: ", follower)
	if err != nil {
		fmt.Print("Error finding crab by ID for Follower %s", err)
	}

	followee, err := app.Crabs.ByID(id)
	fmt.Println("FOLLOWEE: ", followee)
	if err != nil {
		fmt.Print("Error finding crab by ID for Followee %s", err)
	}

	res := app.Follows.Insert(follower, followee)
	if res != nil {
		fmt.Println("error creating follower relationship ")
		app.serverError(w, r, err)
		return
	}
	app.SessionManager.Put(r.Context(), "flash", "Follow successfully created!")

}

func (app *Application) followDeletePost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")
	fmt.Println("unfollowing", id)
	if id == "" {
		app.NotFound(w)
		return
	}

	crabID := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	follower, err := app.Crabs.ByID(crabID)
	if err != nil {
		fmt.Print("ERR %s", err)
	}
	followee, err := app.Crabs.ByID(id)
	if err != nil {
		fmt.Print("ERR %s", err)
	}
	res := app.Follows.Delete(follower, followee)
	if res != nil {
		fmt.Println("error deleting follower relationship ")
		app.serverError(w, r, err)
		return
	}
	app.SessionManager.Put(r.Context(), "flash", "Unfollow successfully created!")
}
