package models

import (
	"context"

	"cloud.google.com/go/firestore"
)

type Course struct {
	ID           string     `firestore:"-"`
	Title        string     `firestore:"title"`
	Description  string     `firestore:"description"`
	Teacher      string     `firestore:"teacher"`
	TeacherImg   string     `firestore:"-"`
	Price        int        `firestore:"price"`
	Duration     string     `firestore:"duration"`
	NumberOfLecs int        `firestore:"number_of_lecs"`
	Lecs         []Lec      `firestore:"-"`
	Exams        []Exam     `firestore:"-"`
	Materials    []Material `firestore:"-"`
}

type CourseModel struct {
	DB *firestore.Client
}

func (c *CourseModel) Get(ctx context.Context, courseId string) (Course, error) {
	courseDoc, err := c.DB.Collection("courses").Doc(courseId).Get(ctx)
	if err != nil {
		return Course{}, err
	}
	var course Course
	courseDoc.DataTo(&course)
	course.ID = courseDoc.Ref.ID
	return course, nil
}
