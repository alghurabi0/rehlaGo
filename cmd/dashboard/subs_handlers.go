package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) subPage(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}
	subId := r.PathValue("subId")
	if subId == "" {
		app.notFound(w)
		return
	}
	sub, err := app.sub.Get(ctx, user.ID, subId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}
	payments, err := app.payment.GetAll(ctx, user.ID, subId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	answers, err := app.answer.GetAll(ctx, user.ID, sub.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Sub = sub
	data.Payments = payments
	data.Answers = answers
	app.render(w, http.StatusOK, "sub.tmpl.html", data)
}

func (app *application) createSub(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	sub := &models.Subscription{}
	courseId := r.FormValue("courseId")
	if courseId == "" {
		http.Error(w, "must provide course id", http.StatusBadRequest)
		return
	} else {
		course, err := app.course.Get(ctx, courseId)
		if err != nil {
			app.errorLog.Println(err)
			app.clientError(w, http.StatusBadRequest)
			return
		}
		sub.CourseTitle = course.Title
	}
	status := r.FormValue("status")
	if status == "active" {
		sub.Active = true
	} else {
		sub.Active = false
	}
	id, err := app.sub.Create(ctx, user.ID, courseId, sub)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("empty id"))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%s", user.ID), http.StatusSeeOther)
}

func (app *application) editSub(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	subId := r.PathValue("subId")
	if subId == "" {
		app.notFound(w)
		return
	}
	sub, err := app.sub.Get(ctx, user.ID, subId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var updates []firestore.Update
	status := r.FormValue("status")
	if status == "active" {
		updates = append(updates, firestore.Update{
			Path:  "active",
			Value: true,
		})
	} else {
		updates = append(updates, firestore.Update{
			Path:  "active",
			Value: false,
		})
	}
	err = app.sub.Update(ctx, user.ID, sub.ID, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%s/%s", user.ID, sub.ID), http.StatusSeeOther)
}

func (app *application) deleteSub(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}
	subId := r.PathValue("subId")
	if subId == "" {
		app.notFound(w)
		return
	}
	sub, err := app.sub.Get(ctx, user.ID, subId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Println(err)
		return
	}
	subDocRef := app.sub.DB.Collection("users").Doc(user.ID).Collection("subs").Doc(sub.ID)
	err = models.DeleteAll(ctx, subDocRef)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%s", user.ID), http.StatusSeeOther)
}

func (app *application) createPayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	subId := r.PathValue("subId")
	if subId == "" {
		app.notFound(w)
		return
	}
	sub, err := app.sub.Get(ctx, userId, subId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	amountStr := r.FormValue("amount")
	if amountStr == "" {
		http.Error(w, "must provide amount paid", http.StatusBadRequest)
		return
	}
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	validUntilStr := r.FormValue("valid_until")
	if validUntilStr == "" {
		http.Error(w, "must provide valid_until paid", http.StatusBadRequest)
		return
	}
	validUntil, err := time.Parse("2006-01-02", validUntilStr)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	payment := &models.Payment{
		AmountPaid:    amount,
		ValidUntil:    validUntil,
		DateOfPayment: time.Now(),
	}
	id, err := app.payment.Create(ctx, user.ID, sub.ID, payment)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("empty id"))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%s/%s", user.ID, sub.ID), http.StatusSeeOther)
}

func (app *application) deletePayment(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	subId := r.PathValue("subId")
	if subId == "" {
		app.notFound(w)
		return
	}
	sub, err := app.sub.Get(ctx, userId, subId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	paymentId := r.PathValue("paymentId")
	if paymentId == "" {
		app.notFound(w)
		return
	}
	payment, err := app.payment.Get(ctx, userId, subId, paymentId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.payment.Delete(ctx, user.ID, sub.ID, payment.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/users/%s/%s", user.ID, sub.ID), http.StatusSeeOther)
}
