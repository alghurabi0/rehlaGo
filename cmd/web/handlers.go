package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
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
	if lec.Order > 3 {
		if !data.IsSubscribed {
			app.unauthorized(w, "subRequired")
			return
		}
	}
	data.Lec = lec
	data.TemplateTitle = lec.Title
	app.renderFull(w, http.StatusOK, "lec.tmpl.html", data)
}

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
	examId := r.PathValue("examId")
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Exam = exam
	data.TemplateTitle = exam.Title
	app.renderFull(w, http.StatusOK, "exam.tmpl.html", data)
}

func (app *application) createAnswer(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
        app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
        app.unauthorized(w, "subRequired")
		return
	}
	var info struct {
		courseId string
		examId   string
		userId   string
		filename string
	}
	info.courseId = r.PathValue("courseId")
	info.examId = r.PathValue("examId")
	userId := app.getUserId(r)
	if userId == "" {
		app.serverError(w, errors.New("userId is not a string"))
		return
	}
	info.userId = userId
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
		return
	}
	err = app.answer.Set(ctx, info.userId, info.courseId, info.examId, exam.Title, info.filename)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "success")
}

func (app *application) progressPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.renderFull(w, http.StatusOK, "progress.tmpl.html", data)
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
	app.renderFull(w, http.StatusOK, "progress.tmpl.html", data)

}

func (app *application) gradesPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
        app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
        app.unauthorized(w, "subRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	answers, err := app.answer.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Answers = answers
	app.renderFull(w, http.StatusOK, "gradesPage.tmpl.html", data)
}

func (app *application) answerPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
        app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
        app.unauthorized(w, "subRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	answer, err := app.answer.Get(ctx, user.ID, courseId, examId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	data.Answer = answer
	examUrl, err := app.exam.GetExamUrl(courseId, examId)
	if err != nil {
		app.serverError(w, errors.New("can't get exam url"))
		return
	}
	data.ExamURL = examUrl
	app.renderFull(w, http.StatusOK, "answer.tmpl.html", data)
}

func (app *application) materialsPage(w http.ResponseWriter, r *http.Request) {
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
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	mats, err := app.material.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.Materials = *mats
	data.Course = course
	app.renderFull(w, http.StatusOK, "materials.tmpl.html", data)
}

func (app *application) signUpPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "signup.tmpl.html", nil)
}

func (app *application) validateSignUp(w http.ResponseWriter, r *http.Request) {
	// get values from json object
	formData := struct {
		Firstname   string `json:"firstname"`
		Lastname    string `json:"lastname"`
		Phone       string `json:"phone_number"`
		ParentPhone string `json:"parent_phone_number"`
		Pwd         string `json:"password"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// validate the form
	ctx := context.Background()
	err = app.user.CheckUserExists(ctx, formData.Phone)
	if err != nil {
		app.clientError(w, http.StatusConflict)
		print(err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("success"))
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	// get values from json object
	formData := struct {
		Firstname   string `json:"firstname"`
		Lastname    string `json:"lastname"`
		Phone       string `json:"phone_number"`
		ParentPhone string `json:"parent_phone_number"`
		Pwd         string `json:"password"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// create the user
	ctx := context.Background()
	userId, sessId, err := app.user.Create(ctx, formData.Firstname, formData.Lastname, formData.Phone, formData.ParentPhone, formData.Pwd)
	if err != nil {
		app.serverError(w, err)
		return
	}
	cookie := &http.Cookie{
		Name:     "rehlaSessionId",
		Value:    sessId,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
	}
	cookie1 := &http.Cookie{
		Name:     "rehlaUserId",
		Value:    userId,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.SetCookie(w, cookie1)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("success"))
}
