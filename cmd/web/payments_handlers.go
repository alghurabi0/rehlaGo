package main

import (
	"context"
	"net/http"
	"strings"
)

func (app *application) paymentsPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getAllSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "payments.tmpl.html", data)
}

func (app *application) paymentHistory(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
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
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	payments, err := app.payment.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.UserPayments = *payments
	data.Course = course
	app.renderFull(w, http.StatusOK, "paymentHistory.tmpl.html", data)
}
