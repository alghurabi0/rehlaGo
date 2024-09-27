package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"strconv"

	"github.com/felixge/httpsnoop"
)

func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
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

func (app *application) metrics(next http.Handler) http.Handler {
	totalReqs := expvar.NewInt("total_requests_received")
	totalRes := expvar.NewInt("total_respones_send")
	totalTime := expvar.NewInt("total_processing_time")
	totalResStat := expvar.NewMap("total_responses_send_by_status")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalReqs.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		totalRes.Add(1)
		totalTime.Add(metrics.Duration.Microseconds())
		totalResStat.Add(strconv.Itoa(metrics.Code), 1)
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
			app.errorLog.Printf("couldn't get user: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
		ctx = context.WithValue(ctx, userModelContextKey, user)
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
			app.notFound(w)
			return
		}
		user, err := app.getUser(r)
		if err != nil {
			app.errorLog.Printf("couldn't get user struct from context: %v\n", err)
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
