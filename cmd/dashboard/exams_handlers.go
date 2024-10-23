package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) examsPage(w http.ResponseWriter, r *http.Request) {
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
	data.HxRoute = fmt.Sprintf("/courses/%s/exam", courseId)
	app.render(w, http.StatusOK, "exams.tmpl.html", data)
}

func (app *application) examPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Exam = exam
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s/exams/%s", courseId, examId)
	app.render(w, http.StatusOK, "exam.tmpl.html", data)
}

func (app *application) createExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "must provide title", http.StatusBadRequest)
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
	file, handler, err := r.FormFile("exam_file")
	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "must provide exam file", http.StatusBadRequest)
			return
		}
		app.errorLog.Printf("%v\n", err)
		http.Error(w, "error with getting file from form", http.StatusBadRequest)
		return
	}

	defer file.Close()
	examId := app.GenerateRandomID()
	path := fmt.Sprintf("courses/%s/exams/%s/%s", courseId, examId, handler.Filename)
	ctx := context.Background()
	file_url, object, err := app.storage.UploadFile(ctx, file, *handler, path)
	if err != nil {
		app.serverError(w, err)
		return
	}

	exam := &models.Exam{
		Title:    title,
		Order:    order,
		URL:      file_url,
		FilePath: path,
	}
	ctx = context.Background()
	id, err := app.exam.Create(ctx, courseId, examId, exam)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty exam id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams/%s", courseId, id), http.StatusSeeOther)
}

func (app *application) editExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	exam := &models.Exam{}
	title := r.FormValue("title")
	if title != "" {
		exam.Title = title
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
		exam.Order = order
	}
	file, handler, err := r.FormFile("exam_file")
	var object *storage.ObjectHandle
	if err != nil {
		if err != http.ErrMissingFile {
			app.errorLog.Printf("%v\n", err)
			http.Error(w, "Error processing file upload", http.StatusBadRequest)
			return
		}
	} else {
		defer file.Close()
		ctx := context.Background()
		path := fmt.Sprintf("courses/%s/exams/%s/%s", courseId, examId, handler.Filename)
		url, obj, err := app.storage.UploadFile(ctx, file, *handler, path)
		object = obj
		if err != nil {
			app.serverError(w, err)
			return
		}
		if url == "" {
			app.serverError(w, errors.New("empty file url after uploading to storage"))
			return
		}
		exam.URL = url
		exam.FilePath = path
	}

	updates := app.createFirestoreUpdateArr(exam, true)
	ctx := context.Background()
	err = app.exam.Update(ctx, courseId, examId, updates)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams/%s", courseId, examId), http.StatusSeeOther)
}

func (app *application) deleteExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		http.Error(w, fmt.Sprintf("exam with id %s doesn't exist", examId), http.StatusBadRequest)
		return
	}
	err = app.storage.DeleteFile(ctx, exam.FilePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.exam.Delete(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.redis.Del(ctx, fmt.Sprintf("course:%s:exam:%s", courseId, examId)).Err()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams", courseId), http.StatusSeeOther)
}

func (app *application) createExamPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	exam := &models.Exam{}
	data := app.newTemplateData(r)
	data.HxMethod = "post"
	data.HxRoute = fmt.Sprintf("/courses/%s/exams", courseId)
	data.Exam = exam
	app.render(w, http.StatusOK, "createExamPage.tmpl.html", data)
}
