package main

import (
	"context"
	"fmt"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(("Referrer-Policy"), "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

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
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.Background()
		user, err := app.user.Get(ctx, userId)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
		ctx = context.WithValue(ctx, userModelContextKey, user)
		print("middleware\n")
		fmt.Printf("%v", user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (app *application) isSubscribed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isLoggedIn := app.isLoggedInCheck(r)
		if !isLoggedIn {
			next.ServeHTTP(w, r)
			return
		}
		courseId := r.PathValue("courseId")
		if courseId == "" {
			next.ServeHTTP(w, r)
			return
		}
		user, err := app.getUser(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.Background()
		isSubscribed := app.sub.IsActive(ctx, user.ID, courseId)
		if !isSubscribed {
			next.ServeHTTP(w, r)
			return
		}
		ctx = context.WithValue(r.Context(), isSubscribedContextKey, true)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
