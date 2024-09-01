package main

import (
	"context"
	"net/http"
)

func (app *application) correctExams(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exams, err := app.exam.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Exams = exams
	app.render(w, http.StatusOK, "exams.tmpl.html", data)
}
