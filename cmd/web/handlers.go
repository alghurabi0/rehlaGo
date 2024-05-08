package main

import (
	"context"
	"fmt"
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

func (app *application) lecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	lecId := r.PathValue("lecId")
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Lec = lec
	data.TemplateTitle = lec.Title
	app.renderFull(w, http.StatusOK, "lec.tmpl.html", data)
}

func (app *application) examPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	examId := r.PathValue("examId")
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.Exam = exam
	data.TemplateTitle = exam.Title
	app.renderFull(w, http.StatusOK, "exam.tmpl.html", data)
}

func (app *application) createAnswer(w http.ResponseWriter, r *http.Request) {
	var info struct {
		courseId string
		examId   string
		userId   string
		filename string
	}
	info.courseId = r.PathValue("courseId")
	info.examId = r.PathValue("examId")
	info.userId = "12345"
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	info.filename = r.PostFormValue("filename")
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, info.courseId, info.examId)
	if err != nil {
		app.serverError(w, err)
	}
	err = app.answer.Set(ctx, info.userId, info.courseId, info.examId, exam.Title, info.filename)
	if err != nil {
		app.serverError(w, err)
	}
	fmt.Fprintf(w, "success")
}

func (app *application) progressPage(w http.ResponseWriter, r *http.Request) {
}
