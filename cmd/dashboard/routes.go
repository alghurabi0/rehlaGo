package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/dashboard/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// is logged in middleware
	isLoggedIn := alice.New(app.isLoggedIn)
	// admin or corrector middleware
	isAdmin := alice.New(app.isLoggedIn, app.isAdmin)
	mux.Handle("GET /", isLoggedIn.ThenFunc(app.home))
	mux.Handle("GET /courses", isAdmin.ThenFunc(app.courses))
	mux.Handle("GET /course", isAdmin.ThenFunc(app.createCoursePage))
	mux.Handle("POST /courses", isAdmin.ThenFunc(app.createCourse))
	mux.Handle("GET /courses/{id}", isAdmin.ThenFunc(app.coursePage))
	mux.Handle("PATCH /courses/{id}", isAdmin.ThenFunc(app.editCourse))
	//mux.Handle("GET /courses/{courseId}/lec/{lecId}", isLoggedIn.ThenFunc(app.lecPage))
	//mux.Handle("GET /courses/{courseId}/exam/{examId}", isSubscribed.ThenFunc(app.examPage))
	//mux.Handle("POST /answers/{courseId}/{examId}", isSubscribed.ThenFunc(app.createAnswer))
	//mux.Handle("GET /materials", isLoggedIn.ThenFunc(app.materialsPage))
	//mux.Handle("GET /materials/{courseId}", isSubscribed.ThenFunc(app.courseMaterials))
	//mux.Handle("GET /progress", isLoggedIn.ThenFunc(app.progressPage))
	//mux.Handle("GET /progress/{courseId}", isSubscribed.ThenFunc(app.gradesPage))
	//mux.Handle("GET /progress/{courseId}/{examId}", isSubscribed.ThenFunc(app.answerPage))
	//mux.Handle("GET /payments", isLoggedIn.ThenFunc(app.paymentsPage))
	//mux.Handle("GET /payments/{courseId}", isLoggedIn.ThenFunc(app.paymentHistory))
	//mux.Handle("GET /mycourses", isLoggedIn.ThenFunc(app.myCoursesPage))
	//mux.Handle("GET /mycourses/{courseId}", isLoggedIn.ThenFunc(app.myCourse))
	//mux.Handle("GET /myprofile", isLoggedIn.ThenFunc(app.myprofile))
	//mux.Handle("GET /privacy_policy", isLoggedIn.ThenFunc(app.policyPage))
	//mux.Handle("GET /contact", isLoggedIn.ThenFunc(app.contactPage))
	//mux.Handle("POST /contact", isLoggedIn.ThenFunc(app.contactMessage))
	//mux.Handle("GET /reset", isLoggedIn.ThenFunc(app.resetPasswordPage))
	//mux.Handle("POST /reset", isLoggedIn.ThenFunc(app.resetPassword))

	//mux.Handle("GET /signup", isLoggedIn.ThenFunc(app.signUpPage))
	//mux.Handle("POST /signup_validate", isLoggedIn.ThenFunc(app.validateSignUp))
	//mux.Handle("POST /signup", isLoggedIn.ThenFunc(app.createUser))
	mux.HandleFunc("GET /login", app.loginPage)
	mux.HandleFunc("POST /login", app.login)

	standard := alice.New(app.recoverPanic, app.logRequest, app.session.LoadAndSave)

	return standard.Then(mux)

}