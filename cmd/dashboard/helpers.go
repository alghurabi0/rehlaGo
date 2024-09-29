package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"

	"cloud.google.com/go/firestore"
	"github.com/alghurabi0/rehla/internal/dashboard_models"
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

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		IsLoggedIn: app.isLoggedInCheck(r),
		IsAdmin:    app.isAdminCheck(r),
	}
}

func (app *application) isLoggedInCheck(r *http.Request) bool {
	isLoggedIn, ok := r.Context().Value(isLoggedInContextKey).(bool)
	if !ok {
		return false
	}
	return isLoggedIn
}

func (app *application) isAdminCheck(r *http.Request) bool {
	isAdmin, ok := r.Context().Value(isAdminContextKey).(bool)
	if !ok {
		return false
	}
	return isAdmin
}

// func (app *application) getUserId(r *http.Request) (string, error) {
// user, ok := r.Context().Value(userModelContextKey).(*dashboard_models.DashboardUser)
// if !ok {
// return "", errors.New("can't get user object from context")
// }
// return user.ID, nil
// }
func (app *application) getUser(r *http.Request) (*dashboard_models.DashboardUser, error) {
	user, ok := r.Context().Value(userModelContextKey).(*dashboard_models.DashboardUser)
	if !ok {
		return &dashboard_models.DashboardUser{}, errors.New("can't get user object from context")
	}
	return user, nil
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
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

func (app *application) getCorrectorCourses(courseIds []string) (*[]models.Course, error) {
	var courses []models.Course
	ctx := context.Background()
	for _, courseId := range courseIds {
		course, err := app.course.Get(ctx, courseId)
		if err != nil {
			return &[]models.Course{}, fmt.Errorf("couldn't get a corretor course, courseId: %s, error: %v", courseId, err)
		}
		if course.Active {
			courses = append(courses, *course)
		}
	}
	return &courses, nil
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

func (app *application) removeString(slice []string, str string) []string {
	for i, v := range slice {
		if v == str {
			// Remove the string by appending the elements before and after the index
			return append(slice[:i], slice[i+1:]...)
		}
	}
	// Return the original slice if the string is not found
	return slice
}
