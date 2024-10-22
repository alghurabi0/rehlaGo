package main

import (
	"context"
	"net/http"
	"strings"
)

func (app *application) examPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	courseId := r.PathValue("courseId")
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if strings.TrimSpace(examId) == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	exam, err := app.getExam(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Exam = exam
	data.TemplateTitle = exam.Title
	app.renderFull(w, http.StatusOK, "exam.tmpl.html", data)
}
