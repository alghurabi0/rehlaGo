package main

import (
	"context"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	scsfs "github.com/alexedwards/scs/firestore"
	"github.com/alexedwards/scs/v2"
	"github.com/alghurabi0/rehla/internal/fileStorage"
	"github.com/alghurabi0/rehla/internal/models"
	"github.com/redis/go-redis/v9"
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
	payment       *models.PaymentModel
	contact       *models.ContactModel
	session       *scs.SessionManager
	storage       *fileStorage.StorageModel
	redis         *redis.Client
}

var version string

func main() {
	// cmd flags
	addr := flag.String("addr", ":4000", "HTTP network address")
	credFile := flag.String("cred-file", "./internal/rehla-74745-firebase-adminsdk-m9ksq-dc2a61849d.json", "Path to the credentials file")
	dfBkt := flag.String("default-bucket", "rehla-74745.appspot.com", "Defualt google storage bucket")
	versionDisplay := flag.Bool("version", false, "display version and exit")
	flag.Parse()

	if *versionDisplay {
		fmt.Printf("Version\t%s\n", version)
		os.Exit(0)
	}

	// loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	ctx := context.Background()
	db, strg, err := initDB_AUTH(ctx, *credFile, *dfBkt)
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
	session.Lifetime = 7200 * time.Hour
	session.Cookie.Secure = true

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		errorLog.Fatalf("failed to ping redis: %v\n", err)
	}
	infoLog.Println("redis connected")

	// metrics
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
		course:        &models.CourseModel{DB: db},
		lec:           &models.LecModel{DB: db},
		exam:          &models.ExamModel{DB: db, ST: strg},
		material:      &models.MaterialModel{DB: db, ST: strg},
		answer:        &models.AnswerModel{DB: db, ST: strg},
		user:          &models.UserModel{DB: db},
		sub:           &models.SubscriptionModel{DB: db},
		payment:       &models.PaymentModel{DB: db},
		contact:       &models.ContactModel{DB: db},
		session:       session,
		storage:       &fileStorage.StorageModel{ST: strg},
		redis:         rdb,
	}
	/*
		tlsConfig := &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		}
	*/
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
		//TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//
	courses, err := app.course.GetAll(context.Background())
	if err != nil {
		app.errorLog.Printf("failed to get courses--main\n")
		return
	}
	for _, course := range *courses {
		foo, err := json.Marshal(&course)
		if err != nil {
			app.errorLog.Printf("failed to marshal firestore doc to redis --main, err: %v \n", err)
			return
		}

		err = app.redis.Set(context.Background(), fmt.Sprintf("course:%s", course.ID), foo, 0).Err()
		if err != nil {
			app.errorLog.Printf("failed to add course %v to redis, err: %v\n", course, err)
			return
		}
	}
	foo, err := json.Marshal(*courses)
	if err != nil {
		app.errorLog.Printf("failed to marshal firestore docs to redis --main, err: %v \n", err)
		return
	}

	err = app.redis.Set(context.Background(), "courses", foo, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to add courses--main to redis, err: %v\n", err)
		return
	}

	infoLog.Printf("starting the srv and listening on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func initDB_AUTH(ctx context.Context, credFile, dfBkt string) (*firestore.Client, *storage.Client, error) {
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

	return firestoreClient, storageClient, nil
}
