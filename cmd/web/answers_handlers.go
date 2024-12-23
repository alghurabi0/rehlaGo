package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alghurabi0/rehla/internal/models"
	"github.com/alghurabi0/rehla/internal/validator"
)

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
	courseId := r.PathValue("courseId")
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	examId := r.PathValue("examId")
	if strings.TrimSpace(examId) == "" {
		app.notFound(w)
		return
	}
	fail_route := fmt.Sprintf("/courses/%s/exam/%s", courseId, examId)
	exam, err := app.getExam(ctx, courseId, examId)
	if err != nil {
		app.serverErrorLog(err)
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}
	userId := app.getUserId(r)
	if userId == "" {
		app.serverErrorLog(errors.New("user id is empty string while submitting answer"))
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}
	file, handler, err := r.FormFile("answer_file")
	if err != nil {
		if err == http.ErrMissingFile {
			data.HxRoute = fail_route
			app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
			app.infoLog.Println("missing answer file")
			return
		}
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		app.errorLog.Println(err)
		return
	}
	defer file.Close()

	// validation
	v := validator.Validator{}
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		app.serverErrorLog(err)
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}
	file.Seek(0, io.SeekStart)
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
	v.Check(validator.FileTypeAllowed(buf, allowedTypes), "file_type", "file type is not allowed")
	v.Check(validator.FileSize(handler, 10*1024*1024), "file_size", "file size must be 10MB or less")
	if v.Errors != nil {
		err = json.NewEncoder(w).Encode(v.Errors)
		if err != nil {
			app.serverErrorLog(err)
			data.HxRoute = fail_route
			app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
			return
		}
		app.errorLog.Println(v.Errors)
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}

	path := fmt.Sprintf("courses/%s/exams/%s/answers/%s", courseId, examId, userId)
	url, object, err := app.storage.UploadFile(ctx, file, *handler, path)
	if err != nil {
		app.serverErrorLog(err)
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}

	answer := &models.Answer{
		UserId:           userId,
		CourseId:         courseId,
		ExamId:           examId,
		ExamTitle:        exam.Title,
		URL:              url,
		StoragePath:      path,
		Corrected:        false,
		DateOfSubmission: time.Now(),
	}
	err = app.answer.Create(ctx, answer)
	if err != nil {
		deleterErr := object.Delete(ctx)
		if deleterErr != nil {
			app.serverErrorLog(deleterErr)
		}
		app.serverErrorLog(err)
		data.HxRoute = fail_route
		app.render(w, http.StatusOK, "fail_exam.tmpl.html", data)
		return
	}

	data.HxRoute = fmt.Sprintf("/courses/%s", courseId)
	app.render(w, http.StatusOK, "success_exam.tmpl.html", data)
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
	data.User = user
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
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	answers, err := app.answer.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Answers = answers
	data.User = user
	app.renderFull(w, http.StatusOK, "grades.tmpl.html", data)
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
		app.serverError(w, errors.New("can't get exam url"))
		return
	}
	answer, err := app.answer.Get(ctx, user.ID, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.ExamURL = exam.URL
	data.Answer = answer
	data.User = user
	app.renderFull(w, http.StatusOK, "answer.tmpl.html", data)
}
