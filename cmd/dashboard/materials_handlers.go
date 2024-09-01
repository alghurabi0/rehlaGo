package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) materialsPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	materials, err := app.material.GetAll(ctx, courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Materials = materials
	data.HxRoute = fmt.Sprintf("/courses/%s/material", courseId)
	app.render(w, http.StatusOK, "materials.tmpl.html", data)
}

func (app *application) materialPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	materialId := r.PathValue("materialId")
	if materialId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	material, err := app.material.Get(ctx, courseId, materialId)
	if err != nil {
		app.notFound(w)
		return
	}

	data := app.newTemplateData(r)
	data.Material = material
	data.HxMethod = "patch"
	data.HxRoute = fmt.Sprintf("/courses/%s/materials/%s", courseId, materialId)
	app.render(w, http.StatusOK, "material.tmpl.html", data)
}

func (app *application) createMaterial(w http.ResponseWriter, r *http.Request) {
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
	file, handler, err := r.FormFile("material_file")
	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "must provide material file", http.StatusBadRequest)
			return
		}
		app.errorLog.Printf("%v\n", err)
		http.Error(w, "error with getting file from form", http.StatusBadRequest)
		return
	}

	defer file.Close()
	path := fmt.Sprintf("courses/%s/materials/%s", courseId, handler.Filename)
	ctx := context.Background()
	file_url, object, err := app.storage.UploadFile(ctx, file, *handler, path)
	if err != nil {
		app.serverError(w, err)
		return
	}

	material := &models.Material{
		Title:    title,
		Order:    order,
		URL:      file_url,
		FilePath: path,
	}
	ctx = context.Background()
	id, err := app.material.Create(ctx, courseId, material)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}
	if id == "" {
		app.serverError(w, errors.New("got empty material id from firestore"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/materials/%s", courseId, id), http.StatusSeeOther)
}

func (app *application) editMaterial(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
		return
	}
	materialId := r.PathValue("materialId")
	if materialId == "" {
		app.notFound(w)
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data (max 10 mb)", http.StatusBadRequest)
		return
	}
	material := &models.Material{}
	title := r.FormValue("title")
	if title != "" {
		material.Title = title
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
		material.Order = order
	}
	file, handler, err := r.FormFile("material_file")
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
		path := fmt.Sprintf("courses/%s/materials/%s", courseId, handler.Filename)
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
		material.URL = url
		material.FilePath = path
	}

	updates := app.createMaterialUpdateArr(material)
	ctx := context.Background()
	err = app.material.Update(ctx, courseId, materialId, updates)
	if err != nil {
		object.Delete(ctx)
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/materials/%s", courseId, materialId), http.StatusSeeOther)
}

func (app *application) deleteMaterial(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	materialId := r.PathValue("materialId")
	if materialId == "" {
		app.notFound(w)
	}
	ctx := context.Background()
	material, err := app.material.Get(ctx, courseId, materialId)
	if err != nil {
		http.Error(w, fmt.Sprintf("material with id %s doesn't exist", materialId), http.StatusBadRequest)
		return
	}
	err = app.storage.DeleteFile(ctx, material.FilePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.material.Delete(ctx, courseId, materialId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s/materials", courseId), http.StatusSeeOther)
}

func (app *application) createMaterialPage(w http.ResponseWriter, r *http.Request) {
	courseId := r.PathValue("courseId")
	if courseId == "" {
		app.notFound(w)
	}
	material := &models.Material{}
	data := app.newTemplateData(r)
	data.HxMethod = "post"
	data.HxRoute = fmt.Sprintf("/courses/%s/materials", courseId)
	data.Material = material
	app.render(w, http.StatusOK, "createMaterialPage.tmpl.html", data)
}
