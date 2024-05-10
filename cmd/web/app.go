package web

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html/template"
	"krabber.net/internal/models"
	"krabber.net/internal/models/mailer"
	"sync"
)

// Define an application struct to hold the application-wide dependencies for the
// web application.
// Add a molts field to the application struct. This will allow us to
// make the MoltModel object available to our handlers.
type Application struct {
	Comments    *models.CommentModel
	Crabs       *models.CrabModel
	Follows     *models.FollowModel
	FormDecoder *form.Decoder
	//Logger         *slog.Logger
	Molts          *models.MoltModel
	Mailer         mailer.Mailer
	SessionManager *scs.SessionManager
	TemplateCache  map[string]*template.Template
	Likes          *models.LikesModel
	Tokens         *models.TokenModel
	Trench         *models.TrenchModel
	Wg             sync.WaitGroup
}
