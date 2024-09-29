package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/alghurabi0/rehla/internal/models"
	"github.com/alghurabi0/rehla/internal/validator"
	"google.golang.org/api/iterator"
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
	// check if user logged in, redirect to homepage
	isLoggedIn := app.isLoggedInCheck(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}
	// get values from json object
	user := &models.User{}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	user.Firstname = r.PostFormValue("firstname")
	user.Lastname = r.PostFormValue("lastname")
	user.PhoneNumber = r.PostFormValue("phone_number")
	user.ParentPhoneNumber = r.PostFormValue("parent_phone_number")
	user.Pwd = r.PostFormValue("password")
	// validation
	// TODO - check if user with same number exist and send code 409
	v := validator.Validator{}
	v.Check(validator.NotBlank(user.Firstname), "firstname", "firstname shouldn't be empty")
	v.Check(validator.NotBlank(user.Lastname), "lastname", "lastname shouldn't be empty")
	v.Check(validator.ValidPhoneNumber(user.PhoneNumber), "phone_number", "phone_number should be a valid iraqi number of 11 digits")
	v.Check(validator.ValidPhoneNumber(user.ParentPhoneNumber), "parent_phone_number", "parent_phone_number should be a valid iraqi number of 11 digits")
	v.Check(validator.Password(user.Pwd), "password", "password should be at least 8 chars, number, or symbols")

	if v.Errors != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(v.Errors)
		if err != nil {
			app.serverError(w, err)
			return
		}
		return
	}
	// check if user already exist
	ctx := context.Background()
	count := 0
	iter := app.user.DB.Collection("users").Where("phone_number", "==", user.PhoneNumber).Documents(ctx)
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			app.serverError(w, err)
		}
		count++
	}
	if count > 0 {
		app.clientError(w, http.StatusConflict)
		return
	}
	// create the user
	user.Verified = false
	userId, err := app.user.Create(ctx, user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(userId))
}

func (app *application) verifyUser(w http.ResponseWriter, r *http.Request) {
	isLoggedIn := app.isLoggedInCheck(r)
	if isLoggedIn {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}
	userId := r.PostFormValue("userId")
	if userId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	_, err = app.user.Get(ctx, userId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := &models.User{}
	user.Verified = true
	updates := app.createFirestoreUpdateArr(user, true)
	err = app.user.Update(ctx, userId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(ctx, "userId", userId)
	w.WriteHeader(http.StatusOK)
}
