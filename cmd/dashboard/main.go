package dashboard

import (
	"context"
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
	scsfs "github.com/alexedwards/scs/firestore"
	"github.com/alexedwards/scs/v2"
	"github.com/alghurabi0/rehla/internal/models"
	"google.golang.org/api/option"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
	session       *scs.SessionManager
	course        *models.CourseModel
	lec           *models.LecModel
	exam          *models.ExamModel
	material      *models.MaterialModel
	answer        *models.AnswerModel
	user          *models.UserModel
	sub           *models.SubscriptionModel
	payment       *models.PaymentModel
	contact       *models.ContactModel
}

var version string

func main() {
	addr := flag.String("addr", ":4001", "HTTP network address")
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

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// firestore db and google storage initilizations
	ctx := context.Background()
	db, strg, err := getShit(ctx, *credFile, *dfBkt)
	if err != nil {
		errorLog.Fatal(err)
	}

	// sessions manager
	session := scs.New()
	session.Store = scsfs.New(db)
	session.Lifetime = 100 * time.Hour

	app := &application{
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: templateCache,
		course:        &models.CourseModel{DB: db},
		lec:           &models.LecModel{DB: db},
		exam:          &models.ExamModel{DB: db, ST: strg},
		material:      &models.MaterialModel{DB: db, ST: strg},
		answer:        &models.AnswerModel{DB: db, ST: strg},
		sub:           &models.SubscriptionModel{DB: db},
		payment:       &models.PaymentModel{DB: db},
		contact:       &models.ContactModel{DB: db},
		session:       session,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	infoLog.Printf("starting dashboard server and listening on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func getShit(ctx context.Context, credFile, dfBdkt string) (*firestore.Client, *storage.Client, error) {
	opt := option.WithCredentialsFile(credFile)
	cfg := &firebase.Config{
		StorageBucket: dfBdkt,
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

	//ping db
	err = pingDB(ctx, firestoreClient)
	if err != nil {
		log.Fatalln(err)
	}

	return firestoreClient, storageClient, nil
}

func pingDB(ctx context.Context, firestoreClient *firestore.Client) error {
	// ping the database to check if it's connected
	docRef := firestoreClient.Collection("ping").Doc("test")
	docSnapshot, err := docRef.Get(ctx)
	if err != nil {
		return err
	}
	var data map[string]interface{}
	if err := docSnapshot.DataTo(&data); err != nil {
		return err
	}
	expectedValue := "pong"
	if value, ok := data["ping"].(string); !ok || value != expectedValue {
		return fmt.Errorf("ping test failed, expected %s, got %s", expectedValue, value)
	}

	return nil
}
