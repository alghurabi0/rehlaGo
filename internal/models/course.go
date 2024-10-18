package models

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	gcloud "cloud.google.com/go/storage"
	"firebase.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type Course struct {
	ID               string       `firestore:"-"`
	Title            string       `firestore:"title"`
	Description      string       `firestore:"description"`
	Teacher          string       `firestore:"teacher"`
	TeacherImg       string       `firestore:"teacher_img"`
	FilePath         string       `firestore:"file_path"`
	Cover            string       `firestore:"cover"`
	CoverPath        string       `firestore:"cover_path"`
	Price            int          `firestore:"price"`
	FolderId         string       `firestore:"folder_id"`
	NumberOfLecs     int          `firestore:"number_of_lecs"`
	Lecs             []Lec        `firestore:"-"`
	Exams            []Exam       `firestore:"-"`
	Materials        []Material   `firestore:"-"`
	UserSubscription Subscription `firestore:"-"`
	UserPayments     []Payment    `firestore:"-"`
	UserLastPayment  Payment      `firestore:"-"`
	UserAmountPaid   int          `firestore:"-"`
	// TODO - add to dashboard
	Active bool `firestore:"active"`
	Free   bool `firestore:"free"`
}

type CourseModel struct {
	DB *firestore.Client
	ST *storage.Client
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

func (c *CourseModel) GetAllActive(ctx context.Context) (*[]Course, error) {
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
		if course.Active {
			course.ID = doc.Ref.ID
			courses = append(courses, course)
		}
	}
	return &courses, nil
}

func (c *CourseModel) Update(ctx context.Context, courseId string, updates []firestore.Update) error {
	_, err := c.DB.Collection("courses").Doc(courseId).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

func (c *CourseModel) Create(ctx context.Context, title, description, teacher string, price int, photo multipart.File, handler multipart.FileHeader, cover multipart.File, coverHand multipart.FileHeader) (string, error) {
	bkt, err := c.ST.DefaultBucket()
	if err != nil {
		return "", err
	}
	// Upload the file to Firebase Storage
	wc := bkt.Object(handler.Filename).NewWriter(ctx)
	defer wc.Close()
	// Copy the uploaded file's content to Firebase storage
	if _, err := io.Copy(wc, photo); err != nil {
		return "", err
	}
	expiration := time.Now().Add(time.Hour * 8640)
	opts := &gcloud.SignedURLOptions{
		Expires: expiration,
		Method:  http.MethodGet,
	}
	url, err := bkt.SignedURL(handler.Filename, opts)
	if err != nil {
		return "", fmt.Errorf("couldn't get signed url: %v", err)
	}
	if url == "" {
		return "", errors.New("empty photo signed file url")
	}
	// Upload the file to Firebase Storage
	wcc := bkt.Object(coverHand.Filename).NewWriter(ctx)
	defer wcc.Close()
	// Copy the uploaded file's content to Firebase storage
	if _, err := io.Copy(wc, cover); err != nil {
		return "", err
	}
	coverUrl, err := bkt.SignedURL(coverHand.Filename, opts)
	if err != nil {
		return "", fmt.Errorf("couldn't get signed url: %v", err)
	}
	if coverUrl == "" {
		return "", errors.New("empty photo signed file url")
	}
	// create wistia folder
	jsonData := []byte(fmt.Sprintf(`{"name": "%s", "public": "true"}`, title))
	folderId, err := wistiaReq("POST", "https://api.wistia.com/v1/projects", jsonData)
	if err != nil {
		return "", fmt.Errorf("couldn't create a wistia folder: %s", err)
	}

	course := &Course{
		Title:       title,
		Description: description,
		Teacher:     teacher,
		TeacherImg:  url,
		FilePath:    handler.Filename,
		Cover:       coverUrl,
		CoverPath:   coverHand.Filename,
		Price:       price,
		FolderId:    folderId,
		Active:      true,
	}
	doc, _, err := c.DB.Collection("courses").Add(ctx, course)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func (c *CourseModel) Delete(ctx context.Context, id string) error {
	_, err := c.DB.Collection("courses").Doc(id).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
