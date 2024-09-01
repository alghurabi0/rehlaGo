package main

import (
	"context"
	"fmt"
	"net/http"

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
	ctx := context.Background()
	answer, err := app.answer.Get(ctx, userId, courseId, examId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}
	user, err := app.user.Get(ctx, userId)
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
	http.Redirect(w, r, fmt.Sprintf("/correct/%s/%s/%s", courseId, examId, userId), http.StatusSeeOther)
}
