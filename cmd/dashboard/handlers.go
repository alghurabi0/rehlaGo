package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
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

func (app *application) createCoursePage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Course = &models.Course{}
	data.HxMethod = "post"
	data.HxRoute = "/courses"
	app.render(w, http.StatusOK, "createCoursePage.tmpl.html", data)
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
		err := app.storage.DeleteFile(ctx, exam.FilePath)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting storage file, exam: %v, error: %v", exam, err))
			return
		}
		err = app.exam.Delete(ctx, id, exam.ID)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting exam from firestore, exam: %v, error: %v", exam, err))
			return
		}

	}
	for _, material := range *materials {
		err := app.storage.DeleteFile(ctx, material.FilePath)
		if err != nil {
			app.serverError(w, fmt.Errorf("error while deleting storage file, material: %v, error: %v", material, err))
			return
		}
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

func (app *application) lecsPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lecs, err := app.lec.GetAll(ctx, courseId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Lecs = lecs
	data.HxRoute = fmt.Sprintf("/courses/%s/lec", courseId)
	app.render(w, http.StatusOK, "lecs.tmpl.html", data)
}

func (app *application) examsPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exams, err := app.exam.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Exams = exams
	data.HxRoute = fmt.Sprintf("/courses/%s/exam", courseId)
	app.render(w, http.StatusOK, "exams.tmpl.html", data)
}

func (app *application) lecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Lec = lec
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s/lecs/%s", courseId, lecId)
	app.render(w, http.StatusOK, "lec.tmpl.html", data)
}

func (app *application) examPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v\n", err), http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Exam = exam
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s/exams/%s", courseId, examId)
	app.render(w, http.StatusOK, "exam.tmpl.html", data)
}

func (app *application) createExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "must provide title", http.StatusBadRequest)
		return
	}
	orderStr := r.FormValue("order")
	if orderStr == "" {
		http.Error(w, "must provide order", http.StatusBadRequest)
		return
	}
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "couldn't convert order to valid integer", http.StatusBadRequest)
		return
	}
	if order < 1 {
		http.Error(w, "order can't be less than 1", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("exam_file")
	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "must provide exam file", http.StatusBadRequest)
			return
		}
		app.errorLog.Printf("%v\n", err)
		http.Error(w, "error with getting file from form", http.StatusBadRequest)
		return
	}

	defer file.Close()
	path := fmt.Sprintf("courses/%s/exams/%s", courseId, handler.Filename)
	ctx := context.Background()
	file_url, object, err := app.storage.UploadFile(ctx, file, *handler, path)
	if err != nil {
		app.serverError(w, err)
		return
	}

	exam := &models.Exam{
		Title:    title,
		Order:    order,
		URL:      file_url,
		FilePath: path,
	}
	ctx = context.Background()
	id, err := app.exam.Create(ctx, courseId, exam)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty exam id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams/%s", courseId, id), http.StatusSeeOther)
}

func (app *application) editExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	exam := &models.Exam{}
	title := r.FormValue("title")
	if title != "" {
		exam.Title = title
	}
	orderStr := r.FormValue("order")
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err != nil {
			http.Error(w, "invalid order number format", http.StatusBadRequest)
			return
		}
		if order < 1 {
			http.Error(w, "order can't be smaller than 1", http.StatusBadRequest)
			return
		}
		exam.Order = order
	}
	file, handler, err := r.FormFile("exam_file")
	var object *storage.ObjectHandle
	if err != nil {
		if err != http.ErrMissingFile {
			app.errorLog.Printf("%v\n", err)
			http.Error(w, "Error processing file upload", http.StatusBadRequest)
			return
		}
	} else {
		defer file.Close()
		ctx := context.Background()
		path := fmt.Sprintf("courses/%s/exams/%s", courseId, handler.Filename)
		url, obj, err := app.storage.UploadFile(ctx, file, *handler, path)
		object = obj
		if err != nil {
			app.serverError(w, err)
			return
		}
		if url == "" {
			app.serverError(w, errors.New("empty file url after uploading to storage"))
			return
		}
		exam.URL = url
		exam.FilePath = path
	}

	updates := app.createExamUpdateArr(exam)
	ctx := context.Background()
	err = app.exam.Update(ctx, courseId, examId, updates)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams/%s", courseId, examId), http.StatusSeeOther)
}

func (app *application) createExamPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	exam := &models.Exam{}
	data := app.newTemplateData(r)
	data.HxMethod = "post"
	data.HxRoute = fmt.Sprintf("/courses/%s/exams", courseId)
	data.Exam = exam
	app.render(w, http.StatusOK, "createExamPage.tmpl.html", data)
}

func (app *application) createLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "must provide title", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "must provide description", http.StatusBadRequest)
		return
	}
	url := r.FormValue("video_url")
	if url == "" {
		http.Error(w, "must provide url", http.StatusBadRequest)
		return
	}
	orderStr := r.FormValue("order")
	if orderStr == "" {
		http.Error(w, "must provide order", http.StatusBadRequest)
		return
	}
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "couldn't convert order to valid integer", http.StatusBadRequest)
		return
	}
	if order < 1 {
		http.Error(w, "order can't be less than 1", http.StatusBadRequest)
		return
	}

	lec := &models.Lec{
		Title:       title,
		Order:       order,
		Description: description,
		VideoUrl:    url,
	}
	ctx := context.Background()
	id, err := app.lec.Create(ctx, courseId, lec)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty exam id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs/%s", courseId, id), http.StatusSeeOther)
}

func (app *application) createLecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		http.Error(w, "course with this id doesn't exist", http.StatusBadRequest)
		return
	}

	lec := &models.Lec{}
	data := app.newTemplateData(r)
	data.Course = course
	data.Lec = lec
	data.HxMethod = "post"
	data.HxRoute = fmt.Sprintf("/courses/%s/lecs", courseId)
	data.WistiaToken = os.Getenv("wistia_token")
	app.render(w, http.StatusOK, "createLecPage.tmpl.html", data)
}

func (app *application) editLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	lec := &models.Lec{}
	title := r.FormValue("title")
	if title != "" {
		lec.Title = title
	}
	orderStr := r.FormValue("order")
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err != nil {
			http.Error(w, "invalid order number format", http.StatusBadRequest)
			return
		}
		if order < 1 {
			http.Error(w, "order can't be smaller than 1", http.StatusBadRequest)
			return
		}
		lec.Order = order
	}

	updates := app.createLecUpdateArr(lec)
	ctx := context.Background()
	err = app.lec.Update(ctx, courseId, lecId, updates)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs/%s", courseId, lecId), http.StatusSeeOther)
}

func (app *application) deleteLec(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	lecId := r.PathValue("lecId")
	if lecId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		http.Error(w, fmt.Sprintf("lec with id %s doesn't exist", lecId), http.StatusBadRequest)
		return
	}
	err = app.wistia.DeleteVideo(lec.VideoUrl)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.lec.Delete(ctx, courseId, lecId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lecs", courseId), http.StatusSeeOther)
}

func (app *application) deleteExam(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		http.Error(w, fmt.Sprintf("exam with id %s doesn't exist", examId), http.StatusBadRequest)
		return
	}
	err = app.storage.DeleteFile(ctx, exam.FilePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.exam.Delete(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/exams", courseId), http.StatusSeeOther)
}
