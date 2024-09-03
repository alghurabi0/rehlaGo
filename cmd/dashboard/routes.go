package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/dashboard/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// is logged in middleware
	isLoggedIn := alice.New(app.isLoggedIn)
	// admin or corrector middleware
	isAdmin := alice.New(app.isLoggedIn, app.isAdmin)
	isCorrector := alice.New(app.isLoggedIn, app.isCorrector)

	mux.Handle("GET /", isLoggedIn.ThenFunc(app.home))
	mux.Handle("GET /courses", isAdmin.ThenFunc(app.courses))
	mux.Handle("GET /course", isAdmin.ThenFunc(app.createCoursePage))
	mux.Handle("POST /courses", isAdmin.ThenFunc(app.createCourse))
	mux.Handle("GET /courses/{id}", isAdmin.ThenFunc(app.coursePage))
	mux.Handle("PATCH /courses/{id}", isAdmin.ThenFunc(app.editCourse))
	mux.Handle("DELETE /courses/{id}", isAdmin.ThenFunc(app.deleteCourse))

	mux.Handle("GET /courses/{courseId}/lecs", isAdmin.ThenFunc(app.lecsPage))
	mux.Handle("POST /courses/{courseId}/lecs", isAdmin.ThenFunc(app.createLec))
	mux.Handle("GET /courses/{courseId}/lecs/{lecId}", isAdmin.ThenFunc(app.lecPage))
	mux.Handle("PATCH /courses/{courseId}/lecs/{lecId}", isAdmin.ThenFunc(app.editLec))
	mux.Handle("DELETE /courses/{courseId}/lecs/{lecId}", isAdmin.ThenFunc(app.deleteLec))
	mux.Handle("GET /courses/{courseId}/lec", isAdmin.ThenFunc(app.createLecPage))

	mux.Handle("GET /courses/{courseId}/exams", isAdmin.ThenFunc(app.examsPage))
	mux.Handle("POST /courses/{courseId}/exams", isAdmin.ThenFunc(app.createExam))
	mux.Handle("GET /courses/{courseId}/exams/{examId}", isAdmin.ThenFunc(app.examPage))
	mux.Handle("PATCH /courses/{courseId}/exams/{examId}", isAdmin.ThenFunc(app.editExam))
	mux.Handle("DELETE /courses/{courseId}/exams/{examId}", isAdmin.ThenFunc(app.deleteExam))
	mux.Handle("GET /courses/{courseId}/exam", isAdmin.ThenFunc(app.createExamPage))

	mux.Handle("GET /courses/{courseId}/materials", isAdmin.ThenFunc(app.materialsPage))
	mux.Handle("POST /courses/{courseId}/materials", isAdmin.ThenFunc(app.createMaterial))
	mux.Handle("GET /courses/{courseId}/materials/{materialId}", isAdmin.ThenFunc(app.materialPage))
	mux.Handle("PATCH /courses/{courseId}/materials/{materialId}", isAdmin.ThenFunc(app.editMaterial))
	mux.Handle("DELETE /courses/{courseId}/materials/{materialId}", isAdmin.ThenFunc(app.deleteMaterial))
	mux.Handle("GET /courses/{courseId}/material", isAdmin.ThenFunc(app.createMaterialPage))

	mux.Handle("GET /users", isAdmin.ThenFunc(app.usersPage))
	mux.Handle("POST /users", isAdmin.ThenFunc(app.createUser))
	mux.Handle("GET /users/{userId}", isAdmin.ThenFunc(app.userPage))
	mux.Handle("PATCH /users/{userId}", isAdmin.ThenFunc(app.editUser))
	mux.Handle("DELETE /users/{userId}", isAdmin.ThenFunc(app.deleteUser))
	mux.Handle("GET /user", isAdmin.ThenFunc(app.createUserPage))

	mux.Handle("GET /users/{userId}/{subId}", isAdmin.ThenFunc(app.subPage))
	mux.Handle("POST /users/{userId}", isAdmin.ThenFunc(app.createSub))
	mux.Handle("PATCH /users/{userId}/{subId}", isAdmin.ThenFunc(app.editSub))
	mux.Handle("DELETE /users/{userId}/{subId}", isAdmin.ThenFunc(app.deleteSub))
	mux.Handle("POST /users/{userId}/{subId}", isAdmin.ThenFunc(app.createPayment))
	mux.Handle("DELETE /users/{userId}/{subId}/{paymentId}", isAdmin.ThenFunc(app.deletePayment))

	mux.Handle("GET /correct/{courseId}", isCorrector.ThenFunc(app.correctExams))
	mux.Handle("GET /correct/{courseId}/{examId}", isCorrector.ThenFunc(app.correctAnswers))
	mux.Handle("GET /correct/{courseId}/{examId}/{userId}", isCorrector.ThenFunc(app.correctAnswer))
	mux.Handle("PATCH /correct/{courseId}/{examId}/{userId}", isCorrector.ThenFunc(app.editAnswer))

	mux.HandleFunc("GET /login", app.loginPage)
	mux.HandleFunc("POST /login", app.login)

	standard := alice.New(app.recoverPanic, app.logRequest, app.session.LoadAndSave)

	return standard.Then(mux)

}
