package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/storage"
	scsfs "github.com/alexedwards/scs/firestore"
	"github.com/alexedwards/scs/v2"
	"github.com/alghurabi0/rehla/internal/models"
	"google.golang.org/api/option"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	course        *models.CourseModel
	lec           *models.LecModel
	exam          *models.ExamModel
	material      *models.MaterialModel
	answer        *models.AnswerModel
	user          *models.UserModel
	sub           *models.SubscriptionModel
	session       *scs.SessionManager
	payment       *models.PaymentModel
	contact       *models.ContactModel
}

const version = "1.0.0"

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	projectId := flag.String("project-id", "rehla-74745", "Google Cloud Project ID")
	credFile := flag.String("cred-file", "./internal/rehla-74745-firebase-adminsdk-m9ksq-dc2a61849d.json", "Path to the credentials file")
	dfBkt := flag.String("default-bucket", "rehla-74745.appspot.com", "Defualt google storage bucket")
	versionDisplay := flag.Bool("version", false, "display version and exit")
	flag.Parse()

	if *versionDisplay {
		fmt.Printf("Version\t%s\n", version)
		os.Exit(0)
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	ctx := context.Background()
	db, auth, strg, err := initDB_AUTH(ctx, *projectId, *credFile, *dfBkt)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	session := scs.New()
	session.Store = scsfs.New(db)
	session.Lifetime = 24 * time.Hour

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		course:        &models.CourseModel{DB: db},
		lec:           &models.LecModel{DB: db},
		exam:          &models.ExamModel{DB: db, ST: strg},
		material:      &models.MaterialModel{DB: db, ST: strg},
		answer:        &models.AnswerModel{DB: db, ST: strg},
		user:          &models.UserModel{DB: db, Auth: auth},
		sub:           &models.SubscriptionModel{DB: db},
		payment:       &models.PaymentModel{DB: db},
		contact:       &models.ContactModel{DB: db},
		session:       session,
	}
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	infoLog.Printf("starting the srv and listening on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func initDB_AUTH(ctx context.Context, projectId, credFile, dfBkt string) (*firestore.Client, *auth.Client, *storage.Client, error) {
	opt := option.WithCredentialsFile(credFile)
	cfg := &firebase.Config{
		StorageBucket: dfBkt,
	}
	app, err := firebase.NewApp(ctx, cfg, opt)
	if err != nil {
		log.Fatalln(err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	//TODO - ping the database to check if it's connected
	docRef := firestoreClient.Collection("ping").Doc("test")
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	var data map[string]interface{}
	if err := docSnapshot.DataTo(&data); err != nil {
		return nil, nil, nil, err
	}
	expectedValue := "pong"
	if value, ok := data["ping"].(string); !ok || value != expectedValue {
		return nil, nil, nil, fmt.Errorf("ping test failed, expected %s, got %s", expectedValue, value)
	}

	return firestoreClient, auth, storageClient, nil
}
