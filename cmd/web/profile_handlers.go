package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (app *application) changeProfileImg(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		//http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		app.serverError(w, err)
		return
	}
	photo, handler, err := r.FormFile("userPic")
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer photo.Close()

	ctx := context.Background()
	storagePath := fmt.Sprintf("users/%s/%s", user.ID, handler.Filename)
	url, object, err := app.storage.UploadFile(ctx, photo, *handler, storagePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user.ImgURL = url
	user.ImgPath = storagePath
	updates := app.createFirestoreUpdateArr(user, true)
	err = app.user.Update(ctx, user.ID, updates)
	if err != nil {
		app.serverError(w, err)
		object.Delete(ctx)
		return
	}
	re, err := json.Marshal(user)
	if err != nil {
		app.errorLog.Printf("failed to marshal user to json: %v\n", err)
		app.serverError(w, err)
		return
	}

	err = app.redis.Set(ctx, user.SessionId, re, time.Hour*24).Err()
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("HX-Redirect", "/myprofile")
}
