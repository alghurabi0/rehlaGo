package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	switch user.Role {
	case "admin":
		app.render(w, http.StatusOK, "home.tmpl.html", data)
	case "corrector":
		app.render(w, http.StatusOK, "correctorHome.tmpl.html", data)
	default:
		app.clientError(w, http.StatusUnauthorized)
		return
	}
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	if username == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	password := r.PostFormValue("password")
	if password == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userId, err := app.dashboardUser.ValidateLogin(ctx, username, password)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		app.infoLog.Printf("error validating user: %v\n", err)
		return
	}
	app.session.Put(r.Context(), "userId", userId)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) courses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	courses, err := app.course.GetAll(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Courses = courses
	app.render(w, http.StatusOK, "courses.tmpl.html", data)
}

func (app *application) coursePage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("id")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Course = course
	app.render(w, http.StatusOK, "course.tmpl.html", data)
}

func (app *application) editCourse(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("id")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	teacher := r.FormValue("teacher_name")
	if teacher == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// teacherImg := r.FormValue("teacher_img")
	priceStr := r.FormValue("price")
	if priceStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		http.Error(w, "invalid number format", http.StatusBadRequest)
		return
	}
	if price < 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	_, err = app.course.Update(ctx, courseId, title, description, teacher, price)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		print(err)
		return
	}
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Course = course
	app.render(w, http.StatusOK, "course.tmpl.html", data)
}

func (app *application) createCourse(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	teacher := r.FormValue("teacher_name")
	if teacher == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// teacherImg := r.FormValue("teacher_img")
	priceStr := r.FormValue("price")
	if priceStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		http.Error(w, "invalid number format", http.StatusBadRequest)
		return
	}
	if price < 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	id, err := app.course.Create(ctx, title, description, teacher, price)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		print(err)
		return
	}
	if id == "" {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/courses/%s", id), http.StatusSeeOther)
}
