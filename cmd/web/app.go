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
type Application struct {
	Comments       *models.CommentModel
	Crabs          *models.CrabModel
	Follows        *models.FollowModel
	FormDecoder    *form.Decoder
	Molts          *models.MoltModel
	Mailer         mailer.Mailer
	SessionManager *scs.SessionManager
	TemplateCache  map[string]*template.Template
	Likes          *models.LikesModel
	Tokens         *models.TokenModel
	Trench         *models.TrenchModel
	Wg             sync.WaitGroup
	Notifications  *models.NotificationModel
}
