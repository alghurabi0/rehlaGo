package main

import (
	"context"
	"net/http"
	"strconv"
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

	data := app.newTemplateData(r)
	data.User = user
	app.render(w, http.StatusOK, "user.tmpl.html", data)
}
