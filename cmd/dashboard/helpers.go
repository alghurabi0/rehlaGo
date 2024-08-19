package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
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

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		IsLoggedIn: app.isLoggedInCheck(r),
	}
}

func (app *application) isLoggedInCheck(r *http.Request) bool {
	isLoggedIn, ok := r.Context().Value(isLoggedInContextKey).(bool)
	if !ok {
		return false
	}
	return isLoggedIn
}

//func (app *application) getUserId(r *http.Request) (string, error) {
//user, ok := r.Context().Value(userModelContextKey).(*dashboard_models.DashboardUser)
//if !ok {
//return "", errors.New("can't get user object from context")
//}
//return user.ID, nil
//}
//
//func (app *application) getUser(r *http.Request) (*dashboard_models.DashboardUser, error) {
//print("helper\n")
//fmt.Printf("%v", r.Context().Value(userModelContextKey))
//user, ok := r.Context().Value(userModelContextKey).(*dashboard_models.DashboardUser)
//if !ok {
//return &dashboard_models.DashboardUser{}, errors.New("can't get user object from context")
//}
//return user, nil
//}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
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
