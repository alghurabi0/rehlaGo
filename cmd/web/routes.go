package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("GET /ping", app.ping)

	// is logged in middleware
	isLoggedIn := alice.New(app.session.LoadAndSave, app.isLoggedIn)
    isSubscribed := alice.New(app.session.LoadAndSave, app.isLoggedIn, app.isSubscribed)
	mux.Handle("GET /", isLoggedIn.ThenFunc(app.home))
	mux.Handle("GET /courses", isLoggedIn.ThenFunc(app.courses))
	mux.Handle("GET /courses/{id}", isLoggedIn.ThenFunc(app.coursePage))
	mux.Handle("GET /courses/{courseId}/lec/{lecId}", isLoggedIn.ThenFunc(app.lecPage))
	mux.Handle("GET /courses/{courseId}/exam/{examId}", isLoggedIn.ThenFunc(app.examPage))
	mux.Handle("POST /answers/{courseId}/{examId}", isLoggedIn.ThenFunc(app.createAnswer))
	mux.Handle("GET /materials", isLoggedIn.ThenFunc(app.materialsPage))
	mux.Handle("GET /materials/{courseId}", isLoggedIn.ThenFunc(app.courseMaterials))
	mux.Handle("GET /progress", isLoggedIn.ThenFunc(app.progressPage))
	mux.Handle("GET /progress/{courseId}", isSubscribed.ThenFunc(app.gradesPage))
	mux.Handle("GET /progress/{courseId}/{examId}", isSubscribed.ThenFunc(app.answerPage))
    mux.Handle("GET /payments", isLoggedIn.ThenFunc(app.paymentsPage))
    mux.Handle("GET /payments/{courseId}", isLoggedIn.ThenFunc(app.paymentHistory))
    mux.Handle("GET /mycourses", isLoggedIn.ThenFunc(app.myCoursesPage))
    mux.Handle("GET /mycourses/{courseId}", isLoggedIn.ThenFunc(app.myCourse))
    mux.Handle("GET /myprofile", isLoggedIn.ThenFunc(app.myprofile))
    mux.Handle("GET /privacy_policy", isLoggedIn.ThenFunc(app.policyPage))

	mux.Handle("GET /signup", isLoggedIn.ThenFunc(app.signUpPage))
	mux.Handle("POST /signup_validate", isLoggedIn.ThenFunc(app.validateSignUp))
	mux.Handle("POST /signup", isLoggedIn.ThenFunc(app.createUser))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
