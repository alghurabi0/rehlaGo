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
	isLoggedIn := alice.New(app.isLoggedIn)
	mux.Handle("GET /", isLoggedIn.ThenFunc(app.home))
	mux.Handle("GET /courses", isLoggedIn.ThenFunc(app.courses))
	mux.Handle("GET /courses/{id}", isLoggedIn.ThenFunc(app.coursePage))
	mux.Handle("GET /courses/{courseId}/lec/{lecId}", isLoggedIn.ThenFunc(app.lecPage))
	mux.Handle("GET /courses/{courseId}/exam/{examId}", isLoggedIn.ThenFunc(app.examPage))
	mux.Handle("POST /answers/{courseId}/{examId}", isLoggedIn.ThenFunc(app.createAnswer))
	mux.Handle("GET /progress", isLoggedIn.ThenFunc(app.progressPage))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
