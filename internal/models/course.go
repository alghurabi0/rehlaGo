package models

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Course struct {
	ID               string       `firestore:"-"`
	Title            string       `firestore:"title"`
	Description      string       `firestore:"description"`
	Teacher          string       `firestore:"teacher"`
	TeacherImg       string       `firestore:"-"`
	Price            int          `firestore:"price"`
	Duration         string       `firestore:"duration"`
	NumberOfLecs     int          `firestore:"-"`
	Lecs             []Lec        `firestore:"-"`
	Exams            []Exam       `firestore:"-"`
	Materials        []Material   `firestore:"-"`
	UserSubscription Subscription `firestore:"-"`
	UserPayments     []Payment    `firestore:"-"`
	UserLastPayment  Payment      `firestore:"-"`
	UserAmountPaid   int          `firestore:"-"`
	Active           bool         `firestore:"active"`
}

type CourseModel struct {
	DB *firestore.Client
}

func (c *CourseModel) Get(ctx context.Context, courseId string) (*Course, error) {
	courseDoc, err := c.DB.Collection("courses").Doc(courseId).Get(ctx)
	if err != nil {
		return &Course{}, err
	}
	var course Course
	err = courseDoc.DataTo(&course)
	if err != nil {
		return &Course{}, err
	}
	course.ID = courseDoc.Ref.ID
	return &course, nil
}

func (c *CourseModel) GetAll(ctx context.Context) (*[]Course, error) {
	coursesIter := c.DB.Collection("courses").Documents(ctx)
	var courses []Course
	for {
		doc, err := coursesIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var course Course
		if err := doc.DataTo(&course); err != nil {
			return nil, err
		}
		course.ID = doc.Ref.ID
		courses = append(courses, course)
	}
	return &courses, nil
}

func (c *CourseModel) Update(ctx context.Context, id, title, description, teacher string, price int) (string, error) {
	return "", nil
}

func (c *CourseModel) Create(ctx context.Context, title, description, teacher string, price int) (string, error) {
	return "", nil
}
