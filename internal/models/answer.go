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

type Answer struct {
	ID               string    `firestore:"-"`
	CourseId         string    `firestore:"course_id"`
	ExamTitle        string    `firestore:"exam_title"`
	StoragePath      string    `firestore:"storage_path"`
	Grade            int       `firestore:"mark"`
	OutOf            int       `firestore:"out_of"`
	Notes            string    `firestore:"notes"`
	DateOfSubmission time.Time `firestore:"date_of_submission"`
	URL              string    `firestore:"-"`
}

type AnswerModel struct {
	DB *firestore.Client
	ST *storage.Client
}

func (s *AnswerModel) Get(ctx context.Context, userId, courseId, ansId string) (*Answer, error) {
	ansDoc, err := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(courseId).Collection("answers").Doc(ansId).Get(ctx)
	if err != nil {
		return &Answer{}, err
	}
	var ans Answer
	err = ansDoc.DataTo(&ans)
	if err != nil {
		return &Answer{}, err
	}
	ans.ID = ansDoc.Ref.ID
	url, err := s.GetAnswerUrl(userId, courseId, ansId)
	if err == nil {
		return &Answer{}, err
	}
	ans.URL = url
	return &ans, nil
}

func (s *AnswerModel) GetAll(ctx context.Context, userId, courseId string) (*[]Answer, error) {
	ansIterator := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(courseId).Collection("answers").Documents(ctx)
	var answers []Answer
	for {
		doc, err := ansIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var ans Answer
		if err := doc.DataTo(&ans); err != nil {
			return nil, err
		}
		ans.ID = doc.Ref.ID
		answers = append(answers, ans)
	}
	return &answers, nil
}

func (s *AnswerModel) Set(ctx context.Context, userId, courseId, examId, examTitle, filename string) error {
	answer := Answer{
		CourseId:    courseId,
		ExamTitle:   examTitle,
		StoragePath: filename,
	}
	_, err := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(courseId).Collection("answers").Doc(examId).Set(ctx, answer)
	if err != nil {
		return err
	}
	return nil
}

func (s *AnswerModel) GetAnswerUrl(userId, courseId, examId string) (string, error) {
	examPath := fmt.Sprintf("users/%s/courses/%s/asnwers/%s.pdf", userId, courseId, examId)
	bkt, err := s.ST.DefaultBucket()
	if err != nil {
		return "", err
	}
	// send expiration as an arg to this func
	expiration := time.Now().Add(time.Hour)
	opts := &gcloud.SignedURLOptions{
		Expires: expiration,
		Method:  http.MethodGet,
	}
	url, err := bkt.SignedURL(examPath, opts)
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", errors.New("empty exam file url")
	}
	return url, nil
}
