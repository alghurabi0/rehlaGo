package main

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) resetPasswordPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	app.renderFull(w, http.StatusOK, "reset_password.tmpl.html", data)
}

func (app *application) resetPassword(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	currect_password := r.PostFormValue("current_password")
	new_password := r.PostFormValue("new_password")
	confirm := r.PostFormValue("confirm_new_password")
	if new_password != confirm {
		http.Error(w, "new password doesn't match confirm password", http.StatusBadRequest)
		return
	}
	if currect_password != user.Pwd {
		fmt.Println(currect_password)
		fmt.Println(user.Pwd)
		http.Error(w, "current password is wrong", http.StatusBadRequest)
		return
	}
	var updates []firestore.Update
	updates = append(updates, firestore.Update{
		Path:  "pwd",
		Value: new_password,
	})
	ctx := context.Background()
	err = app.user.Update(ctx, user.ID, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) signUpPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "signup.tmpl.html", nil)
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.tmpl.html", nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if data.IsLoggedIn {
		w.Write([]byte("already logged in"))
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	phone := r.PostFormValue("phone_number")
	pass := r.PostFormValue("password")
	ctx := context.Background()
	user, err := app.user.ValidateLogin(ctx, phone, pass)
	if err != nil {
		fmt.Print(err)
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	fmt.Printf("user id is: %s\n", user.ID)
	app.session.Put(r.Context(), "userId", user.ID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	// get values from json object
	formData := &models.User{}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	formData.PhoneNumber = r.PostFormValue("phone_number")
	formData.Pwd = r.PostFormValue("password")
	formData.ParentPhoneNumber = r.PostFormValue("parent_phone_number")
	formData.Firstname = r.PostFormValue("firstname")
	formData.Lastname = r.PostFormValue("lastname")
	// create the user
	ctx := context.Background()
	userId, err := app.user.Create(ctx, formData)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r.Context(), "userId", userId)
	http.Redirect(w, r, "/", http.StatusFound)
}
