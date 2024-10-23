package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *application) updateCourseCache(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// add single redis course
	jsonCourse, err := json.Marshal(course)
	if err != nil {
		app.errorLog.Printf("failed to marshal course %v to json, err: %v\n", course, err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s", course.ID))
		app.redis.Del(ctx, "courses")
	}
	err = app.redis.Set(ctx, fmt.Sprintf("course:%s", course.ID), jsonCourse, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to save this course %v to redis, err: %v\n", course, err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s", course.ID))
		app.redis.Del(ctx, "courses")
	}
	// update all courses redis
	courses, err := app.course.GetAll(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	jsonCourses, err := json.Marshal(*courses)
	if err != nil {
		app.errorLog.Printf("failed to marshal courses %v to json, err: %v\n", courses, err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s", course.ID))
		app.redis.Del(ctx, "courses")
	}
	err = app.redis.Set(ctx, "courses", jsonCourses, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to add all jsonCourses %s to redis, err: %v\n", jsonCourses, err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s", course.ID))
		app.redis.Del(ctx, "courses")
	}

	http.Redirect(w, r, "/cache", http.StatusOK)
}

func (app *application) updateLecsCache(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	// add all lecs to redis
	lecs, err := app.lec.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, lec := range *lecs {
		jsonLec, err := json.Marshal(lec)
		if err != nil {
			app.redis.Del(ctx, fmt.Sprintf("course:%s:lecs", courseId))
			app.serverError(w, err)
			return
		}
		err = app.redis.Set(ctx, fmt.Sprintf("course:%s:lec:%s", courseId, lec.ID), jsonLec, 0).Err()
		if err != nil {
			app.redis.Del(ctx, fmt.Sprintf("course:%s:lecs", courseId))
			app.serverError(w, err)
			return
		}
	}

	jsonLecs, err := json.Marshal(lecs)
	if err != nil {
		app.errorLog.Printf("failed to marshal json -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:lecs", courseId))
	}
	err = app.redis.Set(ctx, fmt.Sprintf("course:%s:lecs", courseId), jsonLecs, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to set lecs -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:lecs", courseId))
	}

	http.Redirect(w, r, "/cache", http.StatusOK)
}

func (app *application) updateExamsCache(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	// add all lecs to redis
	lecs, err := app.exam.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, lec := range *lecs {
		jsonLec, err := json.Marshal(lec)
		if err != nil {
			app.redis.Del(ctx, fmt.Sprintf("course:%s:exams", courseId))
			app.serverError(w, err)
			return
		}
		err = app.redis.Set(ctx, fmt.Sprintf("course:%s:exam:%s", courseId, lec.ID), jsonLec, 0).Err()
		if err != nil {
			app.redis.Del(ctx, fmt.Sprintf("course:%s:exams", courseId))
			app.serverError(w, err)
			return
		}
	}

	jsonLecs, err := json.Marshal(lecs)
	if err != nil {
		app.errorLog.Printf("failed to marshal json -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:exams", courseId))
	}
	err = app.redis.Set(ctx, fmt.Sprintf("course:%s:exams", courseId), jsonLecs, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to set lecs -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:exams", courseId))
	}
	http.Redirect(w, r, "/cache", http.StatusOK)
}

func (app *application) updateMatsCache(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	// add all lecs to redis
	lecs, err := app.material.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	jsonLecs, err := json.Marshal(lecs)
	if err != nil {
		app.errorLog.Printf("failed to marshal json -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:mats", courseId))
	}
	err = app.redis.Set(ctx, fmt.Sprintf("course:%s:mats", courseId), jsonLecs, 0).Err()
	if err != nil {
		app.errorLog.Printf("failed to set lecs -redis, err: %v\n", err)
		app.redis.Del(ctx, fmt.Sprintf("course:%s:mats", courseId))
	}
}
