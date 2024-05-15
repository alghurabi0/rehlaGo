package models

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Answer struct {
	ID               string    `firestore:"-"`
	CourseId         string    `firestore:"course_id"`
	ExamId           string    `firestore:"exam_id"`
	ExamTitle        string    `firestore:"exam_title"`
	StoragePath      string    `firestore:"storage_path"`
	Grade            int       `firestore:"mark"`
	OutOf            int       `firestore:"out_of"`
	DateOfSubmission time.Time `firestore:"date_of_submission"`
}

type AnswerModel struct {
	DB *firestore.Client
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

func (s *AnswerModel) Set(ctx context.Context, userId, courseId, examId, examTitle, filename string) error {
	answer := Answer{
		CourseId:    courseId,
		ExamId:      examId,
		ExamTitle:   examTitle,
		StoragePath: filename,
	}
	_, err := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(courseId).Collection("answers").Doc(examId).Set(ctx, answer)
	if err != nil {
		return err
	}
	return nil
}
