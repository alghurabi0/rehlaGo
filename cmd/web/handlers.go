package main

import (
	"context"
	"net/http"
	"strings"
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
	if data.IsLoggedIn {
		user, err := app.getUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}

		data.User = user
	}
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
	if data.IsLoggedIn {
		user, err := app.getUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.User = user
	}

	data.Courses = courses
	app.renderFull(w, http.StatusOK, "courses.tmpl.html", data)
}

func (app *application) coursePage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if strings.TrimSpace(courseId) == "" {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	ctx := context.Background()
	course, err := app.getCourse(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
	}
	if data.IsLoggedIn {
		user, err := app.getUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.User = user
	}

	data.Course = course
	app.renderFull(w, http.StatusOK, "course.tmpl.html", data)
}

func (app *application) myCoursesPage(w http.ResponseWriter, r *http.Request) {
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
	app.renderFull(w, http.StatusOK, "mycourses.tmpl.html", data)
}

// TODO - wtf is this handler/page?
func (app *application) myCourse(w http.ResponseWriter, r *http.Request) {
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
		app.clientError(w, http.StatusNotFound)
		return
	}
	sub, err := app.sub.Get(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	payments, err := app.payment.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.UserSubscription = *sub
	if len(*payments) > 0 {
		course.UserLastPayment = (*payments)[0] // check in template
		totalPaid := 0
		for _, payment := range *payments {
			totalPaid += payment.AmountPaid
		}
		course.UserAmountPaid = totalPaid // check in template
	}
	data.Course = course
	app.renderFull(w, http.StatusOK, "mycourse.tmpl.html", data)
}

func (app *application) policyPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.renderFull(w, http.StatusOK, "policy.tmpl.html", data)
}

func (app *application) myprofile(w http.ResponseWriter, r *http.Request) {
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
	data.User = user
	app.renderFull(w, http.StatusOK, "myprofile.tmpl.html", data)
}

func (app *application) contactPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	contactInfo, err := app.contact.GetContactInfo(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.ContactInfo = contactInfo
	app.renderFull(w, http.StatusOK, "contact.tmpl.html", data)
}

func (app *application) contactMessage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	fullname := r.PostFormValue("fullname")
	phone_number := r.PostFormValue("phone_number")
	message := r.PostFormValue("message")
	if fullname == "" || phone_number == "" || message == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// TODO - validate
	ctx := context.Background()
	err = app.contact.SendInquiry(ctx, fullname, phone_number, message)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
		// TODO - send errors
	}
	w.WriteHeader(http.StatusOK)
}
