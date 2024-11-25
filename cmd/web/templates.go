package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alghurabi0/rehla/internal/models"
)

type templateData struct {
	ContactInfo       *models.ContactInfo
	CurrentYear       int
	Course            *models.Course
	Courses           *[]models.Course
	SubscribedCourses *[]models.Course
	Lec               *models.Lec
	Exam              *models.Exam
	ExamURL           string
	Answer            *models.Answer
	Answers           *[]models.Answer
	FreeMaterials     *[]models.Material
	HxRoute           string
	IsLoggedIn        bool
	IsSubscribed      bool
	TemplateTitle     string
	User              *models.User
}

var functions = template.FuncMap{
	"subtract":  subtract,
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/html/pages/**/*.tmpl.html")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob("./ui/html/partials/**/*.tmpl.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	ts, err := template.New("signupp.tmpl.html").ParseFiles("./ui/html/auth.tmpl.html")
	if err != nil {
		return nil, err
	}
	ts, err = ts.ParseFiles("./ui/html/partials/auth/signupForm.tmpl.html")
	if err != nil {
		return nil, err
	}
	ts, err = ts.ParseFiles("./ui/html/partials/auth/verifyForm.tmpl.html")
	if err != nil {
		return nil, err
	}
	ts, err = ts.ParseFiles("./ui/html/signupp.tmpl.html")
	if err != nil {
		return nil, err
	}
	cache["signupp.tmpl.html"] = ts

	return cache, nil
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

func (app *application) renderAuth(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", page))
		return
	}
	buf := new(bytes.Buffer)
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(buf, "auth", data)
	if err != nil {
		app.serverError(w, err)
	}
	buf.WriteTo(w)
}

func subtract(a, b int) int {
	return a - b
}

func humanDate(t time.Time) string {
	return t.Format("2006-01-02")
}
