package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/alghurabi0/rehla/internal/models"
)

// serverError helper writes an error message and stack trace to the errorLog
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError helper sends a specific status code and corresponding description
// to the user.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) renderFull(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", page))
		return
	}
	buf := new(bytes.Buffer)
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
	buf.WriteTo(w)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", page))
		return
	}
	buf := new(bytes.Buffer)
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(buf, "main", data)
	if err != nil {
		app.serverError(w, err)
	}
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:  time.Now().Year(),
		IsLoggedIn:   app.isLoggedInCheck(r),
		IsSubscribed: app.isSubscribedCheck(r),
	}
}

func (app *application) isLoggedInCheck(r *http.Request) bool {
	isLoggedIn, ok := r.Context().Value(isLoggedInContextKey).(bool)
	if !ok {
		return false
	}
	return isLoggedIn
}

func (app *application) isSubscribedCheck(r *http.Request) bool {
	isSubscribed, ok := r.Context().Value(isSubscribedContextKey).(bool)
	if !ok {
		return false
	}
	return isSubscribed
}

func (app *application) getCourse(ctx context.Context, courseId string) (*models.Course, error) {
	var course *models.Course
	var lecs *[]models.Lec
	var exams *[]models.Exam

	foo, err := app.redis.Get(ctx, fmt.Sprintf("course:%s", courseId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), course)
		if err != nil {
			course, err = app.course.Get(ctx, courseId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		course, err = app.course.Get(ctx, courseId)
		if err != nil {
			return nil, err
		}
	}
	course.ID = courseId

	foo, err = app.redis.Get(ctx, fmt.Sprintf("course:%s:lecs", courseId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), lecs)
		if err != nil {
			lecs, err = app.lec.GetAll(ctx, courseId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		lecs, err = app.lec.GetAll(ctx, courseId)
		if err != nil {
			return nil, err
		}
	}
	course.Lecs = *lecs
	course.NumberOfLecs = len(course.Lecs)

	foo, err = app.redis.Get(ctx, fmt.Sprintf("course:%s:exams", courseId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), exams)
		if err != nil {
			exams, err = app.exam.GetAll(ctx, courseId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		exams, err = app.exam.GetAll(ctx, courseId)
		if err != nil {
			return nil, err
		}
	}
	course.Exams = *exams

	return course, nil
}

func (app *application) getCourseInfo(ctx context.Context, courseId string) (*models.Course, error) {
	var course *models.Course

	foo, err := app.redis.Get(ctx, fmt.Sprintf("course:%s", courseId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), course)
		if err != nil {
			course, err = app.course.Get(ctx, courseId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		course, err = app.course.Get(ctx, courseId)
		if err != nil {
			return nil, err
		}
	}
	course.ID = courseId

	return course, nil
}

func (app *application) getLec(ctx context.Context, courseId, lecId string) (*models.Lec, error) {
	var lec *models.Lec
	foo, err := app.redis.Get(ctx, fmt.Sprintf("course:%s:lec:%s", courseId, lecId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), lec)
		if err != nil {
			lec, err = app.lec.Get(ctx, courseId, lecId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		lec, err = app.lec.Get(ctx, courseId, lecId)
		if err != nil {
			return nil, err
		}
	}
	lec.ID = lecId
	lec.CourseId = courseId

	return lec, nil
}

func (app *application) getExam(ctx context.Context, courseId, examId string) (*models.Exam, error) {
	var exam *models.Exam
	foo, err := app.redis.Get(ctx, fmt.Sprintf("course:%s:exam:%s", courseId, examId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), exam)
		if err != nil {
			exam, err = app.exam.Get(ctx, courseId, examId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		exam, err = app.exam.Get(ctx, courseId, examId)
		if err != nil {
			return nil, err
		}
	}
	exam.ID = examId
	exam.CourseId = courseId

	return exam, nil
}

func (app *application) getMaterials(ctx context.Context, courseId string) (*[]models.Material, error) {
	var materials *[]models.Material
	foo, err := app.redis.Get(ctx, fmt.Sprintf("course:%s:mats", courseId)).Result()
	if err == nil {
		err = json.Unmarshal([]byte(foo), materials)
		if err != nil {
			materials, err = app.material.GetAll(ctx, courseId)
			if err != nil {
				return nil, err
			}
		}
	} else {
		materials, err = app.material.GetAll(ctx, courseId)
		if err != nil {
			return nil, err
		}
	}

	return materials, nil
}

func (app *application) getUserId(r *http.Request) string {
	user, ok := r.Context().Value(userModelContextKey).(*models.User)
	if !ok {
		return ""
	}
	return user.ID
}

func (app *application) getUser(r *http.Request) (*models.User, error) {
	user, ok := r.Context().Value(userModelContextKey).(*models.User)
	if !ok {
		return &models.User{}, errors.New("can't get user object from context")
	}
	return user, nil
}

func (app *application) unauthorized(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(msg))
}

func (app *application) GenerateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	id := fmt.Sprintf("%x", b)
	return id
}

func (app *application) createFirestoreUpdateArr(data interface{}, excludeZeroValues bool) []firestore.Update {
	var updates []firestore.Update
	val := reflect.ValueOf(data).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// Skip if the field is zero and excludeZeroValues is true
		if excludeZeroValues && isZeroValue(fieldVal) {
			continue
		}

		// Append update only if the value is non-zero
		updates = append(updates, firestore.Update{
			Path:  fieldType.Tag.Get("firestore"), // Use struct tags for field names
			Value: fieldVal.Interface(),
		})
	}

	return updates
}

// Helper function to check if a value is a zero value
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
