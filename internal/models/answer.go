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
	UserId           string    `firestore:"user_id"`
	CourseId         string    `firestore:"course_id"`
	ExamId           string    `firestore:"exam_id"`
	ExamTitle        string    `firestore:"exam_title"`
	StoragePath      string    `firestore:"storage_path"`
	Grade            int       `firestore:"grade"`
	Notes            string    `firestore:"notes"`
	DateOfSubmission time.Time `firestore:"date_of_submission"`
	URL              string    `firestore:"url"`
	Corrected        bool      `firestore:"corrected"`
	Corrector        string    `firestore:"corrector"`
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

func (s *AnswerModel) Create(ctx context.Context, answer *Answer) error {
	_, err := s.DB.Collection("users").Doc(answer.UserId).Collection("subs").Doc(answer.CourseId).Collection("answers").Doc(answer.ExamId).Set(ctx, answer)
	if err != nil {
		return err
	}
	return nil
}

func (s *AnswerModel) Update(ctx context.Context, userId, courseId, examId string, updates []firestore.Update) error {
	_, err := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(courseId).Collection("answers").Doc(examId).Update(ctx, updates)
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
