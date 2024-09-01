package main

import (
	"html/template"
	"path/filepath"

	"github.com/alghurabi0/rehla/internal/models"
)

type templateData struct {
	ContactInfo        *models.ContactInfo
	Course             *models.Course
	Courses            *[]models.Course
	Lec                *models.Lec
	Lecs               *[]models.Lec
	Exam               *models.Exam
	Exams              *[]models.Exam
	Material           *models.Material
	Materials          *[]models.Material
	Answer             *models.Answer
	UncorrectedAnswers *[]models.Answer
	CorrectedAnswers   *[]models.Answer
	User               *models.User
	IsLoggedIn         bool
	IsAdmin            bool
	TemplateTitle      string
	HxMethod           string
	HxRoute            string
	WistiaToken        string
}

var functions = template.FuncMap{}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/dashboard/html/pages/**/*.tmpl.html")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/dashboard/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob("./ui/dashboard/html/partials/**/*.tmpl.html")
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
