package main

import (
	"context"
	"fmt"
	"net/http"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) isLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := app.session.GetString(r.Context(), "userId")
		if userId == "" {
			fmt.Println("no session cookie")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.Background()
		user, err := app.dashboardUser.Get(ctx, userId)
		if err != nil {
			fmt.Println("didn't find session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
		ctx = context.WithValue(ctx, userModelContextKey, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (app *application) isAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := app.getUser(r)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		if user.Role != "admin" {
			app.clientError(w, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), isAdminContextKey, true)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (app *application) isCorrector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := app.getUser(r)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		if user.Role == "corrector" {
			next.ServeHTTP(w, r)
		} else if user.Role == "admin" {
			ctx := context.WithValue(r.Context(), isAdminContextKey, true)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		app.clientError(w, http.StatusUnauthorized)
		return
	})
}
