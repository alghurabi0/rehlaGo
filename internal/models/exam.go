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

type Exam struct {
	ID       string `firestore:"-"`
	CourseId string `firestore:"-"`
	Title    string `firestore:"title"`
	Order    int    `firestore:"order"`
	URL      string `firestore:"-"`
}

type ExamModel struct {
	DB      *firestore.Client
	ST *storage.Client
}

func (e *ExamModel) Get(ctx context.Context, courseId, examId string) (*Exam, error) {
	examDoc, err := e.DB.Collection("courses").Doc(courseId).Collection("exams").Doc(examId).Get(ctx)
	if err != nil {
		return &Exam{}, err
	}
	var exam Exam
	err = examDoc.DataTo(&exam)
	if err != nil {
		return &Exam{}, err
	}
	exam.ID = examDoc.Ref.ID
	exam.CourseId = courseId

    examPath := fmt.Sprintf("courses/%s/exams/%s", courseId, examId)
    bkt, err := e.ST.DefaultBucket()
    if err != nil {
        return &Exam{}, err
    }
    // send expiration as an arg to this func
    expiration := time.Now().Add(time.Hour)
    opts := &gcloud.SignedURLOptions{
        Expires: expiration,
        Method: http.MethodGet,
    }
    url, err := bkt.SignedURL(examPath, opts)
    if err != nil {
        return &Exam{}, err
    }
    if url == "" {
        return &Exam{}, errors.New("empty exam file url")
    }
    exam.URL = url

	return &exam, nil
}

func (e *ExamModel) GetAll(ctx context.Context, courseId string) (*[]Exam, error) {
	examsIter := e.DB.Collection("courses").Doc(courseId).Collection("exams").Documents(ctx)
	var exams []Exam
	for {
		doc, err := examsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var exam Exam
		if err := doc.DataTo(&exam); err != nil {
			return nil, err
		}
		exam.ID = doc.Ref.ID
		exam.CourseId = courseId
		exams = append(exams, exam)
	}
	return &exams, nil
}
