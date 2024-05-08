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
	"firebase.google.com/go/storage"
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
	answer        *models.AnswerModel
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	projectId := flag.String("project-id", "rehla-74745", "Google Cloud Project ID")
	credFile := flag.String("cred-file", "./internal/rehla-74745-firebase-adminsdk-m9ksq-dc2a61849d.json", "Path to the credentials file")
	dfBkt := flag.String("default-bucket", "rehla-74745.appspot.com", "Defualt google storage bucket")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	ctx := context.Background()
	db, auth, strg, err := initDB_AUTH(ctx, *projectId, *credFile, *dfBkt)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	infoLog.Println(auth)
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		course:        &models.CourseModel{DB: db},
		lec:           &models.LecModel{DB: db},
		exam:          &models.ExamModel{DB: db, ST: strg},
		answer:        &models.AnswerModel{DB: db},
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

func initDB_AUTH(ctx context.Context, projectId, credFile, dfBkt string) (*firestore.Client, *firebase.App, *storage.Client, error) {
	opt := option.WithCredentialsFile(credFile)
	cfg := &firebase.Config{
		StorageBucket: dfBkt,
	}
	app, err := firebase.NewApp(ctx, cfg, opt)
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

	return firestoreClient, app, storageClient, nil
}
