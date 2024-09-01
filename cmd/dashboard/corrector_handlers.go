package main

import (
	"context"
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
