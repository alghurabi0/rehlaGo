package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) materialsPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.renderFull(w, http.StatusOK, "materials.tmpl.html", data)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "materials.tmpl.html", data)
}

func (app *application) courseMaterials(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	courseId := r.PathValue("courseId")
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	var course *models.Course

	course, err := app.getCourseInfo(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if !data.IsSubscribed && !course.Free {
		app.unauthorized(w, "subRequired")
		return
	}
	mats, err := app.getMaterials(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	course.Materials = *mats
	data.Course = course
	app.renderFull(w, http.StatusOK, "courseMaterials.tmpl.html", data)
}

func (app *application) freeMaterials(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	ctx := context.Background()
	mats, err := app.material.GetFree(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.FreeMaterials = mats
	app.renderFull(w, http.StatusOK, "freeMaterials.tmpl.html", data)
}
