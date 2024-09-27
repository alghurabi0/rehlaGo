package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) lecsPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lecs, err := app.lec.GetAll(ctx, courseId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Lecs = lecs
	data.HxRoute = fmt.Sprintf("/courses/%s/lec", courseId)
	app.render(w, http.StatusOK, "lecs.tmpl.html", data)
}

func (app *application) lecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Lec = lec
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s/lecs/%s", courseId, lecId)
	app.render(w, http.StatusOK, "lec.tmpl.html", data)
}

func (app *application) createLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "must provide title", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "must provide description", http.StatusBadRequest)
		return
	}
	url := r.FormValue("video_url")
	if url == "" {
		http.Error(w, "must provide url", http.StatusBadRequest)
		return
	}
	orderStr := r.FormValue("order")
	if orderStr == "" {
		http.Error(w, "must provide order", http.StatusBadRequest)
		return
	}
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "couldn't convert order to valid integer", http.StatusBadRequest)
		return
	}
	if order < 1 {
		http.Error(w, "order can't be less than 1", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	lec := &models.Lec{
		Title:       title,
		Order:       order,
		Description: description,
		VideoUrl:    url,
		Free:        course.Free,
	}
	id, err := app.lec.Create(ctx, courseId, lec)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty exam id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs/%s", courseId, id), http.StatusSeeOther)
}

func (app *application) editLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	lec := &models.Lec{}
	title := r.FormValue("title")
	if title != "" {
		lec.Title = title
	}
	orderStr := r.FormValue("order")
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err != nil {
			http.Error(w, "invalid order number format", http.StatusBadRequest)
			return
		}
		if order < 1 {
			http.Error(w, "order can't be smaller than 1", http.StatusBadRequest)
			return
		}
		lec.Order = order
	}

	updates := app.createFirestoreUpdateArr(lec, true)
	ctx := context.Background()
	err = app.lec.Update(ctx, courseId, lecId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs/%s", courseId, lecId), http.StatusSeeOther)
}

func (app *application) deleteLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		http.Error(w, fmt.Sprintf("lec with id %s doesn't exist", lecId), http.StatusBadRequest)
		return
	}
	err = app.wistia.DeleteVideo(lec.VideoUrl)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.lec.Delete(ctx, courseId, lecId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs", courseId), http.StatusSeeOther)
}

func (app *application) createLecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		http.Error(w, "course with this id doesn't exist", http.StatusBadRequest)
		return
	}

	lec := &models.Lec{}
	data := app.newTemplateData(r)
	data.Course = course
	data.Lec = lec
	data.HxMethod = "post"
	data.HxRoute = fmt.Sprintf("/courses/%s/lecs", courseId)
	data.WistiaToken = os.Getenv("wistia_token")
	app.render(w, http.StatusOK, "createLecPage.tmpl.html", data)
}
