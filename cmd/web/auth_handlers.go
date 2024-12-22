package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.User = user
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
	data := app.newTemplateData(r)
	if data.IsLoggedIn {
		http.Error(w, "user is signed in", http.StatusConflict)
		return
	}
	app.renderAuth(w, http.StatusOK, "signupp.tmpl.html", nil)
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if data.IsLoggedIn {
		http.Error(w, "user is signed in", http.StatusConflict)
		return
	}
	app.renderAuth(w, http.StatusOK, "login.tmpl.html", nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if data.IsLoggedIn {
		http.Redirect(w, r, "/", http.StatusConflict)
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	phone := r.PostFormValue("phone_number")
	if strings.TrimSpace(phone) == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	pass := r.PostFormValue("password")
	if strings.TrimSpace(pass) == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	iter := app.user.DB.Collection("users").Where("phone_number", "==", phone).Documents(ctx)
	count := 0
	user := &models.User{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			app.serverError(w, err)
			return
		}
		count++
		err = doc.DataTo(&user)
		if err != nil {
			app.serverError(w, err)
			return
		}
		user.ID = doc.Ref.ID
	}
	if count == 0 {
		w.Header().Set("HX-Retarget", ".errors")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Write([]byte("لا يوجد حساب بهذا الرقم, يرجى أنشاء حساب"))
		return
	} else if count > 1 {
		app.serverError(w, fmt.Errorf("more than one user with this phone number: %s", phone))
		return
	}
	if user.Pwd != pass {
		w.Header().Set("HX-Retarget", ".errors")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Write([]byte("يوجد خطأ في المعلومات المدخلة"))
		return
	}
	if user.SessionId != "" {
		err = app.redis.Del(ctx, user.SessionId).Err()
		if err != nil {
			app.serverError(w, err)
			return
		}

	}
	session_id := app.GenerateRandomID()
	user.SessionId = session_id
	updates := app.createFirestoreUpdateArr(user, true)
	err = app.user.Update(ctx, user.ID, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}
	re, err := json.Marshal(user)
	if err != nil {
		app.errorLog.Printf("failed to marshal user to json: %v\n", err)
		app.serverError(w, err)
		return
	}

	err = app.redis.Set(ctx, session_id, re, time.Hour*24).Err()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r.Context(), "session_id", session_id)
	// Create a new cookie
	cookie := &http.Cookie{
		Name:     "Login-Success",
		Value:    "true",
		MaxAge:   5,     // Expires after 5 seconds
		Path:     "/",    // Cookie available for all paths (optional, but good practice)
		Secure:   true,  // Set to true if using HTTPS (recommended)
		// HttpOnly: true,  // Recommended for security
	}
	// Set the cookie in the response
	http.SetCookie(w, cookie)
	//http.Redirect(w, r, "/", http.StatusFound)
	w.Header().Set("HX-Redirect", "/")
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	// check if user logged in, redirect to homepage
	isLoggedIn := app.isLoggedInCheck(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", http.StatusConflict)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read body request", http.StatusBadRequest)
		return
	}
	app.infoLog.Println(string(body))

	defer r.Body.Close()
	// get values from json object
	user := &models.User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		http.Error(w, "invalid json format", http.StatusBadRequest)
		return
	}
	app.infoLog.Println(user)
	// validation
	v := validator.Validator{}
	v.Check(validator.NotBlank(user.Firstname), "الاسم الأول", "لا يجب ان يكون الاسم الاول فارغ")
	v.Check(validator.NotBlank(user.Lastname), "الاسم الاخير", "لا يجب ان يكون الاسم الاخير فارغ")
	v.Check(validator.ValidPhoneNumber(user.PhoneNumber), "رقم الهاتف", "phone number must be valid iraqi number consist of 11 digits")
	v.Check(validator.ValidPhoneNumber(user.ParentPhoneNumber), "رقم هاتف ولي الأمر", "should be a valid iraqi number of 11 digits")
	v.Check(validator.ValidGender(user.Gender), "gender", "gender should be male or female")
	v.Check(validator.Password(user.Pwd), "الرمز", "يجب ان يتكون الرمز من 8 حروف, ارقام, او رموز")

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
	user.SessionId = ""
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read body request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	userIdStr := string(body)
	userId := strings.Trim(userIdStr, `"`)

	// TODO - if empty
	app.infoLog.Println(userId)
	if strings.TrimSpace(userId) == "" {
		http.Error(w, "empty request body", http.StatusBadRequest)
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
	user.SessionId = app.GenerateRandomID()
	updates := app.createFirestoreUpdateArr(user, true)
	err = app.user.Update(ctx, userId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.redis.Set(ctx, user.SessionId, userJson, time.Hour*24).Err()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r.Context(), "session_id", user.SessionId)
	w.WriteHeader(http.StatusOK)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		w.Header().Set("HX-Redirect", "/")
		return
	}
	app.session.PopString(r.Context(), "session_id")
	w.Header().Set("HX-Redirect", "/")
}
