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

func (app *application) usersPage(w http.ResponseWriter, r *http.Request) {
	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offsetLoc, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "offset must be a number", http.StatusBadRequest)
			return
		}
		offset = offsetLoc
	}

	ctx := context.Background()
	users, err := app.user.GetAll(ctx, offset)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Users = users
	app.render(w, http.StatusOK, "users.tmpl.html", data)
}

func (app *application) userPage(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.notFound(w)
		app.errorLog.Print(err)
		return
	}
	subs, err := app.sub.GetAll(ctx, user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courses, err := app.course.GetAllActive(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/users/%s", user.ID)
	data.Subs = subs
	data.Courses = courses
	app.render(w, http.StatusOK, "user.tmpl.html", data)
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	firstname := r.FormValue("firstname")
	if firstname == "" {
		http.Error(w, "must provide firstname", http.StatusBadRequest)
		return
	}
	lastname := r.FormValue("lastname")
	if lastname == "" {
		http.Error(w, "must provide lastname", http.StatusBadRequest)
		return
	}
	phone_number := r.FormValue("phone_number")
	if phone_number == "" {
		http.Error(w, "must provide phone_number", http.StatusBadRequest)
		return
	}
	parent_phone_number := r.FormValue("parent_phone_number")
	if parent_phone_number == "" {
		http.Error(w, "must provide parent_phone_number", http.StatusBadRequest)
		return
	}
	pwd := r.FormValue("pwd")
	if pwd == "" {
		http.Error(w, "must provide pwd", http.StatusBadRequest)
		return
	}
	gender := r.FormValue("gender")
	if gender == "" {
		http.Error(w, "must provide gender", http.StatusBadRequest)
		return
	}
	if gender != "male" && gender != "female" {
		http.Error(w, "gender has to be male or female", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Firstname:         firstname,
		Lastname:          lastname,
		Pwd:               pwd,
		PhoneNumber:       phone_number,
		ParentPhoneNumber: parent_phone_number,
		Gender:            gender,
	}
	ctx := context.Background()
	id, err := app.user.Create(ctx, user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty exam id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%s", id), http.StatusSeeOther)
}

func (app *application) editUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Print(err)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't parse form, error: %v", err), http.StatusBadRequest)
		return
	}
	userUpdates := &models.User{}
	firstname := r.FormValue("firstname")
	if firstname != "" {
		userUpdates.Firstname = firstname
	}
	lastname := r.FormValue("lastname")
	if lastname != "" {
		userUpdates.Firstname = lastname
	}
	phone := r.FormValue("phone_number")
	if phone != "" {
		userUpdates.PhoneNumber = phone
	}
	parentPhone := r.FormValue("parent_phone_number")
	if parentPhone != "" {
		userUpdates.ParentPhoneNumber = parentPhone
	}
	pwd := r.FormValue("pwd")
	if pwd != "" {
		userUpdates.Pwd = pwd
	}
	file, handler, err := r.FormFile("user_img")
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
		path := fmt.Sprintf("users/%s/%s", userId, handler.Filename)
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
		userUpdates.ImgURL = url
		userUpdates.ImgPath = path
	}

	updates := app.createFirestoreUpdateArr(userUpdates, true)
	err = app.user.Update(ctx, userId, updates)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}
	err = app.storage.DeleteFile(ctx, user.ImgPath)
	if err != nil {
		app.errorLog.Print(err)
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%s", userId), http.StatusSeeOther)
}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	user, err := app.user.Get(ctx, userId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.errorLog.Print(err)
		return
	}
	userDocRef := app.user.DB.Collection("users").Doc(user.ID)
	err = models.DeleteAll(ctx, userDocRef)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.storage.DeleteFile(ctx, user.ImgPath)
	if err != nil {
		app.errorLog.Print(err)
	}
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (app *application) createUserPage(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	data := app.newTemplateData(r)
	data.HxMethod = "post"
	data.HxRoute = "/users"
	data.User = user
	app.render(w, http.StatusOK, "createUserPage.tmpl.html", data)
}
