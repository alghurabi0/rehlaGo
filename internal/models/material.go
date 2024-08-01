package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	gcloud "cloud.google.com/go/storage"
	"firebase.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type Material struct {
	ID       string `firestore:"-"`
	CourseId string `firestore:"-"`
	Title    string `firestore:"title"`
	Order    int    `firestore:"order"`
	URL      string `firestore:"-"`
}

type MaterialModel struct {
	DB *firestore.Client
	ST *storage.Client
}

func (m *MaterialModel) Get(ctx context.Context, courseId, matId string) (*Material, error) {
	matDoc, err := m.DB.Collection("courses").Doc(courseId).Collection("materials").Doc(matId).Get(ctx)
	if err != nil {
		return &Material{}, err
	}
	var mat Material
	err = matDoc.DataTo(&mat)
	if err != nil {
		return &Material{}, err
	}
	mat.ID = matDoc.Ref.ID
	mat.CourseId = courseId

	url, err := m.GetMaterialUrl(courseId, matId)
	if err != nil {
		return &Material{}, err
	}
	mat.URL = url

	return &mat, nil
}

func (m *MaterialModel) GetAll(ctx context.Context, courseId string) (*[]Material, error) {
	matsIter := m.DB.Collection("courses").Doc(courseId).Collection("materials").Documents(ctx)
	var mats []Material
	for {
		doc, err := matsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var mat Material
		if err := doc.DataTo(&mat); err != nil {
			return nil, err
		}
		mat.ID = doc.Ref.ID
		mat.CourseId = courseId
		url, err := m.GetMaterialUrl(courseId, mat.ID)
		if err != nil {
			return nil, err
		}
		mat.URL = url
		mats = append(mats, mat)
	}
	return &mats, nil
}

func (m *MaterialModel) GetMaterialUrl(courseId, matId string) (string, error) {
	matPath := fmt.Sprintf("courses/%s/materials/%s.pdf", courseId, matId)
	bkt, err := m.ST.DefaultBucket()
	if err != nil {
		return "", err
	}
	// send expiration as an arg to this func
	expiration := time.Now().Add(time.Hour)
	opts := &gcloud.SignedURLOptions{
		Expires: expiration,
		Method:  http.MethodGet,
	}
	url, err := bkt.SignedURL(matPath, opts)
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", errors.New("empty material file url")
	}
	return url, nil
}
