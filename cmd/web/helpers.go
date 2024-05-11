package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/alghurabi0/rehla/internal/models"
)

// serverError helper writes an error message and stack trace to the errorLog
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError helper sends a specific status code and corresponding description
// to the user.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) renderFull(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", page))
		return
	}
	buf := new(bytes.Buffer)
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
	buf.WriteTo(w)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", page))
		return
	}
	buf := new(bytes.Buffer)
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(buf, "main", data)
	if err != nil {
		app.serverError(w, err)
	}
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:  time.Now().Year(),
		IsLoggedIn:   app.isLoggedInCheck(r),
		IsSubscribed: app.isSubscribedCheck(r),
	}
}

func (app *application) isLoggedInCheck(r *http.Request) bool {
	isLoggedIn, ok := r.Context().Value(isLoggedInContextKey).(bool)
	if !ok {
		return false
	}
	return isLoggedIn
}

func (app *application) isSubscribedCheck(r *http.Request) bool {
	return false
}

func (app *application) getUserId(r *http.Request) string {
	// TODO
	return "12345"
}

func (app *application) getCourse(ctx context.Context, courseId string) (*models.Course, error) {
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		return &models.Course{}, err
	}

	lecs, err := app.lec.GetAll(ctx, courseId)
	if err != nil {
		return &models.Course{}, err
	}
	course.Lecs = *lecs
	course.NumberOfLecs = len(course.Lecs)

	exams, err := app.exam.GetAll(ctx, courseId)
	if err != nil {
		return &models.Course{}, err
	}
	course.Exams = *exams

	return course, nil
}

func (app *application) GenerateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	id := fmt.Sprintf("%x", b)
	return id
}
