package models

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Lec struct {
	ID          string `firestore:"-"`
	CourseId    string `firestore:"-"`
	Title       string `firestore:"title"`
	Description string `firestore:"description"`
	Order       int    `firestore:"order"`
	VideoUrl    string `firestore:"video_url"`
}

type LecModel struct {
	DB *firestore.Client
}

func (l *LecModel) Get(ctx context.Context, courseId, lecId string) (*Lec, error) {
	lecDoc, err := l.DB.Collection("courses").Doc(courseId).Collection("lecs").Doc(lecId).Get(ctx)
	if err != nil {
		return &Lec{}, err
	}
	var lec Lec
	err = lecDoc.DataTo(&lec)
	if err != nil {
		return &Lec{}, err
	}
	lec.ID = lecDoc.Ref.ID
	lec.CourseId = courseId
	return &lec, nil
}

func (l *LecModel) GetAll(ctx context.Context, courseId string) (*[]Lec, error) {
	lecsIter := l.DB.Collection("courses").Doc(courseId).Collection("lecs").Documents(ctx)
	var lecs []Lec
	for {
		doc, err := lecsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var lec Lec
		if err := doc.DataTo(&lec); err != nil {
			return nil, err
		}
		lec.ID = doc.Ref.ID
		lec.CourseId = courseId
		lecs = append(lecs, lec)
	}
	return &lecs, nil
}

func (l *LecModel) Create(ctx context.Context, courseId string, lec *Lec) (string, error) {
	doc, _, err := l.DB.Collection("courses").Doc(courseId).Collection("lecs").Add(ctx, lec)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func (l *LecModel) Update(ctx context.Context, courseId, lecId string, updates []firestore.Update) error {
	_, err := l.DB.Collection("courses").Doc(courseId).Collection("lecs").Doc(lecId).Update(ctx, updates)
	if err != nil {
		return err
	}
	return nil
}

func (l *LecModel) Delete(ctx context.Context, courseId, lecId string) error {
	_, err := l.DB.Collection("courses").Doc(courseId).Collection("lecs").Doc(lecId).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
