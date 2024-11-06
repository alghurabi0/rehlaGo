package main

import (
	"expvar"
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("GET /service-worker.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "service-worker.js")
	})
	mux.HandleFunc("GET /sw-register.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "sw-register.js")
	})
	mux.HandleFunc("GET /firebase-messaging-sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "firebase-messaging-sw.js")
	})
	mux.HandleFunc("GET /ping", app.ping)

	// is logged in middleware
	isLoggedIn := alice.New(app.session.LoadAndSave, app.isLoggedIn)
	isSubscribed := alice.New(app.session.LoadAndSave, app.isLoggedIn, app.isSubscribed)

	mux.Handle("GET /", isLoggedIn.ThenFunc(app.home))
	mux.Handle("GET /courses", isLoggedIn.ThenFunc(app.courses))
	mux.Handle("GET /courses/{courseId}", isSubscribed.ThenFunc(app.coursePage))

	mux.Handle("GET /courses/{courseId}/lec/{lecId}", isSubscribed.ThenFunc(app.lecPage))

	mux.Handle("GET /courses/{courseId}/exam/{examId}", isSubscribed.ThenFunc(app.examPage))
	mux.Handle("POST /answers/{courseId}/{examId}", isSubscribed.ThenFunc(app.createAnswer))

	mux.Handle("GET /materials", isLoggedIn.ThenFunc(app.materialsPage))
	mux.Handle("GET /materials/free", isLoggedIn.ThenFunc(app.freeMaterials))
	mux.Handle("GET /materials/{courseId}", isSubscribed.ThenFunc(app.courseMaterials))

	mux.Handle("GET /progress", isLoggedIn.ThenFunc(app.progressPage))
	mux.Handle("GET /progress/{courseId}", isSubscribed.ThenFunc(app.gradesPage))
	mux.Handle("GET /progress/{courseId}/{examId}", isSubscribed.ThenFunc(app.answerPage))

	mux.Handle("GET /payments", isLoggedIn.ThenFunc(app.paymentsPage))
	mux.Handle("GET /payments/{courseId}", isLoggedIn.ThenFunc(app.paymentHistory))
	mux.Handle("GET /mycourses", isLoggedIn.ThenFunc(app.myCoursesPage))
	mux.Handle("GET /mycourses/{courseId}", isLoggedIn.ThenFunc(app.myCourse))
	mux.Handle("GET /myprofile", isLoggedIn.ThenFunc(app.myprofile))
	mux.Handle("GET /privacy_policy", isLoggedIn.ThenFunc(app.policyPage))
	mux.Handle("GET /contact", isLoggedIn.ThenFunc(app.contactPage))
	mux.Handle("POST /contact", isLoggedIn.ThenFunc(app.contactMessage))

	mux.Handle("GET /reset", isLoggedIn.ThenFunc(app.resetPasswordPage))
	mux.Handle("POST /reset", isLoggedIn.ThenFunc(app.resetPassword))

	mux.Handle("GET /signup", isLoggedIn.ThenFunc(app.signUpPage))
	mux.Handle("POST /signup", isLoggedIn.ThenFunc(app.createUser))
	mux.Handle("POST /verify_signup", isLoggedIn.ThenFunc(app.verifyUser))

	mux.Handle("GET /login", isLoggedIn.ThenFunc(app.loginPage))
	mux.Handle("POST /login", isLoggedIn.ThenFunc(app.login))
	mux.Handle("POST /logout", isLoggedIn.ThenFunc(app.logout))

	mux.Handle("GET /debug/vars", expvar.Handler())

	standard := alice.New(app.metrics, app.recoverPanic, app.logRequest, app.secureHeaders)

	return standard.Then(mux)
}
