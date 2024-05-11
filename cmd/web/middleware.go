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
		ctx := context.Background()
		cookie, err := r.Cookie("rehlaSessionId")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		cookie1, err := r.Cookie("rehlaUserId")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		err = app.user.VerifySessionId(ctx, cookie1.Value, cookie.Value)
		if err != nil {
			// TODO - token invalid
			next.ServeHTTP(w, r)
			app.errorLog.Println(err)
			return
		}
		ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
		ctx = context.WithValue(ctx, userIdContextKey, cookie1.Value)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
