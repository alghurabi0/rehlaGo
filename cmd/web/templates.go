package main

import (
	"html/template"
	"path/filepath"

	"github.com/alghurabi0/rehla/internal/models"
)

type templateData struct {
	ContactInfo       *models.ContactInfo
	CurrentYear       int
	Course            *models.Course
	Courses           *[]models.Course
	SubscribedCourses *[]models.Course
	Lec               *models.Lec
	Exam              *models.Exam
	ExamURL           string
	Answer            *models.Answer
	Answers           *[]models.Answer
	FreeMaterials     *[]models.Material
	IsLoggedIn        bool
	IsSubscribed      bool
	TemplateTitle     string
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
	return cache, nil
}

func subtract(a, b int) int {
	return a - b
}
