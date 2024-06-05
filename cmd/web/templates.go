package main

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/alghurabi0/rehla/internal/models"
)

type templateData struct {
	CurrentYear       int
	Course            *models.Course
	Courses           *[]models.Course
	Lec               *models.Lec
	Exam              *models.Exam
	ExamURL           string
	IsLoggedIn        bool
	IsSubscribed      bool
	TemplateTitle     string
	SubscribedCourses *[]models.Course
	Answer            *models.Answer
	Answers           *[]models.Answer
	User              *models.User
}

var functions = template.FuncMap{
	"subtract": subtract,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/html/pages/**/*.tmpl.html")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob("./ui/html/partials/**/*.tmpl.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	fmt.Println(cache)
	return cache, nil
}

func subtract(a, b int) int {
	return a - b
}
