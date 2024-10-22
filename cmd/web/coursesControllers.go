package main

import (
	"context"

	"github.com/alghurabi0/rehla/internal/models"
)

func (app *application) getSubscribedCourses(ctx context.Context, user models.User) (*[]models.Course, error) {
	var courses []models.Course
	for _, subId := range user.Subscriptions {
		active := app.sub.IsActive(ctx, user.ID, subId)
		if !active {
			continue
		}
		course, err := app.getCourseInfo(ctx, subId)
		if err != nil {
			return &[]models.Course{}, err
		}
		courses = append(courses, *course)
	}

	if len(courses) == 0 {
		return &[]models.Course{}, nil
	}
	return &courses, nil
}

func (app *application) getAllSubscribedCourses(ctx context.Context, user models.User) (*[]models.Course, error) {
	var courses []models.Course
	for _, subId := range user.Subscriptions {
		course, err := app.getCourseInfo(ctx, subId)
		if err != nil {
			return &[]models.Course{}, err
		}
		courses = append(courses, *course)
	}
	return &courses, nil
}
