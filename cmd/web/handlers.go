package main

import (
	"context"
	"net/http"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	app.renderFull(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) courses(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	ctx := context.Background()
	courses, err := app.course.GetAll(ctx)
	if err != nil {
		app.serverError(w, err)
        return
	}
	data.Courses = courses
	app.renderFull(w, http.StatusOK, "courses.tmpl.html", data)
}

func (app *application) coursePage(w http.ResponseWriter, r *http.Request) {
    courseId := r.PathValue("id")
    data := app.newTemplateData(r)
    ctx := context.Background()
    course, err := app.getCourse(ctx, courseId)
    if err != nil {
        app.serverError(w, err)
    }
    data.Course = course
    data.IsSubscribed = true
    app.renderFull(w, http.StatusOK, "course.tmpl.html", data)
}
