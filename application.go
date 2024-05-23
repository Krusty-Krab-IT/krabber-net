package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/form/v4"
	"github.com/joho/godotenv"
	w "krabber.net/cmd/web"
	"krabber.net/internal/models"
	"krabber.net/internal/models/mailer"
	_ "krabber.net/internal/models/validator"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type conf struct {
	port    int
	env     string
	crabmin string
	db      struct {
		tableName string
		region    string
		url       string
		akid      string
		sac       string
		st        string
		source    string
	}
	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

func main() {
	var cfg conf
	var err error
	prod := false

	if !prod {
		cfg.db.tableName = goDotEnvVariable("TABLE_NAME")
		cfg.db.region = goDotEnvVariable("REGION")
		cfg.db.sac = goDotEnvVariable("DB_SAC")
		cfg.db.akid = goDotEnvVariable("DB_AKID")
		cfg.smtp.host = goDotEnvVariable("SMTP_HOST")
		p, err := strconv.Atoi(goDotEnvVariable("SMTP_PORT"))
		if err != nil {
			log.Fatal("ERROR could not read port")
		}
		cfg.smtp.port = p
		cfg.smtp.username = goDotEnvVariable("SMTP_USER")
		cfg.smtp.password = goDotEnvVariable("SMTP_PASS")
		cfg.smtp.sender = goDotEnvVariable("SMTP_SEND")
		cfg.crabmin = goDotEnvVariable("CRABMIN")
	}

	if prod {
		cfg.db.tableName = os.Getenv("TABLE_NAME")
		cfg.db.region = os.Getenv("REGION")
		cfg.db.sac = os.Getenv("DB_SAC")
		cfg.db.akid = os.Getenv("DB_AKID")
		cfg.smtp.host = os.Getenv("SMTP_HOST")
		p, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
		if err != nil {
			log.Fatal("ERROR could not read port")
		}
		cfg.smtp.port = p
		cfg.smtp.username = os.Getenv("SMTP_USER")
		cfg.smtp.password = os.Getenv("SMTP_PASS")
		cfg.smtp.sender = os.Getenv("SMTP_SEND")
		cfg.crabmin = os.Getenv("CRABMIN")
	}

	addr := flag.String("addr", ":5000", "HTTP network address") // default:5000

	svc := newItemService(cfg)
	// Initialize a new template cache...
	templateCache, err := w.NewTemplateCache()
	if err != nil {
		fmt.Println("ERROR with template cache: ", err)
		os.Exit(1)
	}

	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our dynamo db table as session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	// Initialize a models.MoltModel instance containing the connection pool
	// and add it to the application dependencies.
	app := &w.Application{
		//Logger:         logger,
		Molts:          &models.MoltModel{SVC: svc},
		Comments:       &models.CommentModel{SVC: svc},
		Crabs:          &models.CrabModel{SVC: svc},
		Follows:        &models.FollowModel{SVC: svc},
		Tokens:         &models.TokenModel{SVC: svc},
		Trench:         &models.TrenchModel{SVC: svc},
		Likes:          &models.LikesModel{SVC: svc},
		Mailer:         mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		Notifications:  &models.NotificationModel{SVC: svc},
		TemplateCache:  templateCache,
		FormDecoder:    formDecoder,
		SessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = srv.ListenAndServe()
	os.Exit(1)
}

func newItemService(cfg conf) models.ItemService {
	dt := createLocalClient(cfg)
	return models.ItemService{
		ItemTable: dt,
	}
}

func createLocalClient(c conf) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(c.db.region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     c.db.akid,
				SecretAccessKey: c.db.sac,
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
