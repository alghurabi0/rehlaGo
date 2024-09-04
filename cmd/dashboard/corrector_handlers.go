package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alghurabi0/rehla/internal/models"
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

func (app *application) correctAnswers(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	userIds, err := app.storage.GetAnswers(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	var correctedAnswers []models.Answer
	var uncorrectedAnswers []models.Answer
	for _, userId := range userIds {
		answer, err := app.answer.Get(ctx, userId, courseId, examId)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if answer.Corrected {
			correctedAnswers = append(correctedAnswers, *answer)
		} else {
			uncorrectedAnswers = append(uncorrectedAnswers, *answer)
		}
	}

	data := app.newTemplateData(r)
	data.CorrectedAnswers = &correctedAnswers
	data.UncorrectedAnswers = &uncorrectedAnswers
	app.render(w, http.StatusOK, "answers.tmpl.html", data)
}

func (app *application) correctAnswer(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.errorLog.Println(err)
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
		return
	}
	exam, err := app.exam.Get(ctx, course.ID, examId)
	if err != nil {
		app.errorLog.Println(err)
		app.notFound(w)
		return
	}
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.errorLog.Println(err)
		app.notFound(w)
		return
	}
	answer, err := app.answer.Get(ctx, user.ID, course.ID, exam.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Answer = answer
	data.User = user
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/correct/%s/%s/%s", courseId, examId, userId)
	app.render(w, http.StatusOK, "answer.tmpl.html", data)
}

func (app *application) editAnswer(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
		return
	}
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	fmt.Println(user)
	ctx := context.Background()
	_, err = app.answer.Get(ctx, userId, courseId, examId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Print(err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ans := &models.Answer{}
	gradeStr := r.FormValue("grade")
	if gradeStr != "" {
		grade, err := strconv.Atoi(gradeStr)
		if err != nil {
			http.Error(w, "invalid grade number format", http.StatusBadRequest)
			return
		}
		if grade < 0 {
			http.Error(w, "grade can't be smaller than 0", http.StatusBadRequest)
			return
		}
		ans.Grade = grade
	}
	notes := r.FormValue("notes")
	if notes != "" {
		ans.Notes = notes
	}
	ans.Corrected = true
	ans.Corrector = user.Username

	updates := app.createAnswerUpdateArr(ans)
	err = app.answer.Update(ctx, userId, courseId, examId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/correct/%s/%s", courseId, examId), http.StatusSeeOther)
}
