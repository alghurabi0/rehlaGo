package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	switch user.Role {
	case "admin":
		app.render(w, http.StatusOK, "home.tmpl.html", data)
	case "corrector":
		user, err := app.getUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		corrector_courses, err := app.getCorrectorCourses(user.CorrectorCourses)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.Courses = corrector_courses
		app.render(w, http.StatusOK, "correctorHome.tmpl.html", data)
	default:
		app.clientError(w, http.StatusUnauthorized)
		return
	}
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	if username == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	password := r.PostFormValue("password")
	if password == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userId, err := app.dashboardUser.ValidateLogin(ctx, username, password)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		app.infoLog.Printf("error validating user: %v\n", err)
		return
	}
	app.session.Put(r.Context(), "userId", userId)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) courses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	courses, err := app.course.GetAll(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Courses = courses
	app.render(w, http.StatusOK, "courses.tmpl.html", data)
}

func (app *application) coursePage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("id")
	if courseId == "" {
		app.notFound(w)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Course = course
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s", course.ID)
	app.render(w, http.StatusOK, "course.tmpl.html", data)
}

func (app *application) createCourse(w http.ResponseWriter, r *http.Request) {
	// max form size 10 MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	teacher := r.FormValue("teacher_name")
	if teacher == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	priceStr := r.FormValue("price")
	if priceStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		http.Error(w, "invalid number format", http.StatusBadRequest)
		return
	}
	if price < 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	photo, handler, err := r.FormFile("teacher_img")
	if err != nil {
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		app.errorLog.Printf("%v\n", err)
		return
	}
	defer photo.Close()

	ctx := context.Background()
	id, err := app.course.Create(ctx, title, description, teacher, price, photo, *handler)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("empty course id"))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/courses/%s", id), http.StatusSeeOther)
}

func (app *application) editCourse(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("id")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	course := &models.Course{}
	title := r.FormValue("title")
	if title != "" {
		course.Title = title
	}
	description := r.FormValue("description")
	if description != "" {
		course.Description = description
	}
	teacher := r.FormValue("teacher_name")
	if teacher != "" {
		course.Teacher = teacher
	}
	teacherImg, _, err := r.FormFile("teacher_img")
	if err != nil {
		if err != http.ErrMissingFile {
			app.errorLog.Printf("%v\n", err)
			http.Error(w, "Error processing file upload", http.StatusBadRequest)
			return
		}
	} else {
		defer teacherImg.Close()
	}
	priceStr := r.FormValue("price")
	if priceStr != "" {
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			http.Error(w, "invalid number format", http.StatusBadRequest)
			return
		}
		if price < 0 {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		course.Price = price
	}

	updates := app.createFsUpdateArr(course)
	ctx := context.Background()
	err = app.course.Update(ctx, courseId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/courses/%s", courseId), http.StatusSeeOther)
}

func (app *application) deleteCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, id)
	if err != nil {
		http.Error(w, "course with this id doesn't exist", http.StatusBadRequest)
		return
	}
	lecs, err := app.lec.GetAll(ctx, id)
	if err != nil {
		app.serverError(w, fmt.Errorf("can't get lecs: %v", err))
		return
	}
	exams, err := app.exam.GetAll(ctx, id)
	if err != nil {
		app.serverError(w, fmt.Errorf("can't get exams: %v", err))
		return
	}
	materials, err := app.material.GetAll(ctx, id)
	if err != nil {
		app.serverError(w, fmt.Errorf("can't get materials: %v", err))
		return
	}

	for _, lec := range *lecs {
		err := app.wistia.DeleteVideo(lec.VideoUrl)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting wistai video, lec: %v, error: %v", lec, err))
			return
		}
		err = app.lec.Delete(ctx, id, lec.ID)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting lec from firestore, lec: %v, error: %v", lec, err))
			return
		}

	}
	// TODO - delete wistia folder
	for _, exam := range *exams {
		//err := app.storage.DeleteFile(ctx, exam.FilePath)
		//if err != nil {
		//app.serverError(w, fmt.Errorf("error while deleting storage file, exam: %v, error: %v", exam, err))
		//return
		//}
		err = app.exam.Delete(ctx, id, exam.ID)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting exam from firestore, exam: %v, error: %v", exam, err))
			return
		}
	}
	for _, material := range *materials {
		//err := app.storage.DeleteFile(ctx, material.FilePath)
		//if err != nil {
		//app.serverError(w, fmt.Errorf("error while deleting storage file, material: %v, error: %v", material, err))
		//return
		//}
		err = app.material.Delete(ctx, id, material.ID)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting material from firestore, material: %v, error: %v", material, err))
			return
		}
	}

	err = app.storage.DeleteFile(ctx, course.FilePath)
	if err != nil {
		app.serverError(w, fmt.Errorf("error while deleting storage file, course: %v, error: %v", course, err))
		return
	}
	err = app.course.Delete(ctx, id)
	if err != nil {
		app.serverError(w, fmt.Errorf("error while deleting course from firestore: %v", err))
		return
	}

	http.Redirect(w, r, "/courses", http.StatusSeeOther)
}

func (app *application) createCoursePage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Course = &models.Course{}
	data.HxMethod = "post"
	data.HxRoute = "/courses"
	app.render(w, http.StatusOK, "createCoursePage.tmpl.html", data)
}
