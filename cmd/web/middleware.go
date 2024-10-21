package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alghurabi0/rehla/internal/models"
	"github.com/felixge/httpsnoop"
	"google.golang.org/api/iterator"
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
		session_id := app.session.GetString(r.Context(), "session_id")
		if session_id == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.Background()
		val, err := app.redis.Get(ctx, session_id).Result()
		if err != nil {
			app.errorLog.Printf("couldn't get redis key: %s, error: %v", session_id, err)
			iter := app.user.DB.Collection("users").Where("session_id", "==", session_id).Documents(ctx)
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
				http.Error(w, "no_match", http.StatusBadRequest)
				return
			} else if count > 1 {
				app.serverError(w, fmt.Errorf("more than one user with this session_id: %s", session_id))
				return
			}
			re, err := json.Marshal(user)
			if err != nil {
				app.errorLog.Printf("failed to marshal user to json: %v\n", err)
				return
			}
			err = app.redis.Set(ctx, session_id, re, time.Hour*24).Err()
			if err != nil {
				app.serverError(w, err)
				return
			}

			ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
			ctx = context.WithValue(ctx, userModelContextKey, &user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		var user models.User
		err = json.Unmarshal([]byte(val), &user)
		if err != nil {
			app.errorLog.Printf("can't unmarshal json to user: %v\n", err)
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(r.Context(), isLoggedInContextKey, true)
		ctx = context.WithValue(ctx, userModelContextKey, &user)
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
