package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func ping(w http.ResponseWriter, r *http.Request) {
	if r != nil {
		w.Write([]byte("OK"))
	}
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	app.renderFull(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) courses(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	ctx := context.Background()
	courses, err := app.course.GetAll(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Courses = courses
	app.renderFull(w, http.StatusOK, "courses.tmpl.html", data)
}

func (app *application) coursePage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("id")
	data := app.newTemplateData(r)
	ctx := context.Background()
	course, err := app.getCourse(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
	}
	data.Course = course
	app.renderFull(w, http.StatusOK, "course.tmpl.html", data)
}

func (app *application) lecPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	lecId := r.PathValue("lecId")
	ctx := context.Background()
	lec, err := app.lec.Get(ctx, courseId, lecId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	if lec.Order > 3 {
		if !data.IsSubscribed {
			app.unauthorized(w, "subRequired")
			return
		}
	}
	data.Lec = lec
	data.TemplateTitle = lec.Title
	app.renderFull(w, http.StatusOK, "lec.tmpl.html", data)
}

func (app *application) examPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	courseId := r.PathValue("courseId")
	examId := r.PathValue("examId")
	ctx := context.Background()
	exam, err := app.exam.Get(ctx, courseId, examId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Exam = exam
	data.TemplateTitle = exam.Title
	app.renderFull(w, http.StatusOK, "exam.tmpl.html", data)
}

func (app *application) createAnswer(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	var info struct {
		courseId string
		examId   string
		userId   string
		filename string
	}
	info.courseId = r.PathValue("courseId")
	info.examId = r.PathValue("examId")
	userId := app.getUserId(r)
	if userId == "" {
		app.serverError(w, errors.New("user id is empty string"))
		return
	}
	info.userId = userId

	ctx := context.Background()
	exam, err := app.exam.Get(ctx, info.courseId, info.examId)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("answer_file")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("courses/%s/exams/%s/answers/%s", info.courseId, info.examId, info.userId)
	url, object, err := app.storage.UploadFile(ctx, file, *handler, path)
	if err != nil {
		app.serverError(w, err)
		return
	}

	answer := &models.Answer{
		UserId:           info.userId,
		CourseId:         info.courseId,
		ExamId:           info.examId,
		ExamTitle:        exam.Title,
		URL:              url,
		StoragePath:      path,
		Corrected:        false,
		DateOfSubmission: time.Now(),
	}
	err = app.answer.Create(ctx, answer)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "success")
}

func (app *application) progressPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.renderFull(w, http.StatusOK, "progress.tmpl.html", data)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "progress.tmpl.html", data)
}

func (app *application) gradesPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	answers, err := app.answer.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Answers = answers
	app.renderFull(w, http.StatusOK, "grades.tmpl.html", data)
}

func (app *application) answerPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	examId := r.PathValue("examId")
	if examId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	answer, err := app.answer.Get(ctx, user.ID, courseId, examId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	data.Answer = answer
	examUrl, err := app.exam.GetExamUrl(courseId, examId)
	if err != nil {
		app.serverError(w, errors.New("can't get exam url"))
		return
	}
	data.ExamURL = examUrl
	app.renderFull(w, http.StatusOK, "answer.tmpl.html", data)
}

func (app *application) materialsPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.renderFull(w, http.StatusOK, "materials.tmpl.html", data)
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "materials.tmpl.html", data)
}

func (app *application) courseMaterials(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	if !data.IsSubscribed {
		app.unauthorized(w, "subRequired")
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	mats, err := app.material.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.Materials = *mats
	data.Course = course
	app.renderFull(w, http.StatusOK, "courseMaterials.tmpl.html", data)
}

func (app *application) paymentsPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getAllSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "payments.tmpl.html", data)
}

func (app *application) paymentHistory(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	payments, err := app.payment.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.UserPayments = *payments
	data.Course = course
	app.renderFull(w, http.StatusOK, "paymentHistory.tmpl.html", data)
}

func (app *application) myCoursesPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	ctx := context.Background()
	subedCourses, err := app.getAllSubscribedCourses(ctx, *user)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SubscribedCourses = subedCourses
	app.renderFull(w, http.StatusOK, "mycourses.tmpl.html", data)
}

func (app *application) myCourse(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	course, err := app.course.Get(ctx, courseId)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	sub, err := app.sub.Get(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	payments, err := app.payment.GetAll(ctx, user.ID, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	course.UserSubscription = *sub
	if len(*payments) > 0 {
		course.UserLastPayment = (*payments)[0] // check in template
		totalPaid := 0
		for _, payment := range *payments {
			totalPaid += payment.AmountPaid
		}
		course.UserAmountPaid = totalPaid // check in template
	}
	data.Course = course
	app.renderFull(w, http.StatusOK, "mycourse.tmpl.html", data)
}

func (app *application) policyPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.renderFull(w, http.StatusOK, "policy.tmpl.html", data)
}

func (app *application) myprofile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	user, err := app.getUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.User = user
	app.renderFull(w, http.StatusOK, "myprofile.tmpl.html", data)
}

func (app *application) contactPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	contactInfo, err := app.contact.GetContactInfo(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r)
	data.ContactInfo = contactInfo
	app.renderFull(w, http.StatusOK, "contact.tmpl.html", data)
}

func (app *application) contactMessage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	fullname := r.PostFormValue("fullname")
	phone_number := r.PostFormValue("phone_number")
	message := r.PostFormValue("message")
	if fullname == "" || phone_number == "" || message == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// TODO - validate
	ctx := context.Background()
	err = app.contact.SendInquiry(ctx, fullname, phone_number, message)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
		// TODO - send errors
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) resetPasswordPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	app.renderFull(w, http.StatusOK, "reset_password.tmpl.html", data)
}

func (app *application) resetPassword(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if !data.IsLoggedIn {
		app.unauthorized(w, "loginRequired")
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	//currect_password := r.PostFormValue("current_password")
	//new_password := r.PostFormValue("new_password")
	//confirm := r.PostFormValue("confirm_new_password")
	// TODO - validate
	w.WriteHeader(http.StatusOK)
}

func (app *application) signUpPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "signup.tmpl.html", nil)
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.tmpl.html", nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	if data.IsLoggedIn {
		w.Write([]byte("already logged in"))
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	phone := r.PostFormValue("phone_number")
	pass := r.PostFormValue("password")
	ctx := context.Background()
	user, err := app.user.ValidateLogin(ctx, phone, pass)
	if err != nil {
		fmt.Print(err)
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	fmt.Printf("user id is: %s\n", user.ID)
	app.session.Put(r.Context(), "userId", user.ID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	// get values from json object
	formData := &models.User{}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	formData.PhoneNumber = r.PostFormValue("phone_number")
	formData.Pwd = r.PostFormValue("password")
	formData.ParentPhoneNumber = r.PostFormValue("parent_phone_number")
	formData.Firstname = r.PostFormValue("firstname")
	formData.Lastname = r.PostFormValue("lastname")
	// create the user
	ctx := context.Background()
	userId, err := app.user.Create(ctx, formData)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r.Context(), "userId", userId)
	http.Redirect(w, r, "/", http.StatusFound)
}
