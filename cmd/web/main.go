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
	"github.com/alghurabi0/rehla/internal/models"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	course        *models.CourseModel
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	projectId := flag.String("project-id", "xxxx", "Google Cloud Project ID")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	ctx := context.Background()
	db, auth, err := initDB_AUTH(ctx, *projectId)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	tempateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: tempateCache,
		course:        &models.CourseModel{DB: db},
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
	infoLog.Printf("starting the srv and listening on &s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func initDB_AUTH(ctx context.Context, projectId string) (*firestore.Client, *firebase.App, error) {
	conf := &firebase.Config{ProjectID: projectId}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	//TODO - ping the database to check if it's connected
	docRef := firestoreClient.Collection("ping").Doc("test")
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	var data map[string]interface{}
	if err := docSnapshot.DataTo(&data); err != nil {
		return nil, nil, err
	}
	expectedValue := "pong"
	if value, ok := data["ping"].(string); !ok || value != expectedValue {
		return nil, nil, fmt.Errorf("ping test failed, expected %s, got %s", expectedValue, value)
	}

	return firestoreClient, app, nil
}
