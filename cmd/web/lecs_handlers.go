package main

import (
	"context"
	"net/http"
	"strings"
)

func (app *application) lecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	lecId := r.PathValue("lecId")
	if strings.TrimSpace(lecId) == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	if lec.Order > 3 && !data.IsSubscribed && !lec.Free {
		app.unauthorized(w, "subRequired")
		return
	}
	data.Lec = lec
	data.TemplateTitle = lec.Title
	app.renderFull(w, http.StatusOK, "lec.tmpl.html", data)
}
