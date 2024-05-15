package main

import (
	"context"

	"github.com/alghurabi0/rehla/internal/models"
)


func (app *application) getSubscribedCourses(ctx context.Context, user models.User) (*[]models.Course, error) {
    var courses []models.Course
    for _, subId := range user.Subscriptions {
        valid := app.sub.IsActive(ctx, user.ID, subId)
        if !valid {
            continue
        }
        course, err := app.course.Get(ctx, subId)
        if err != nil {
            return &[]models.Course{}, err
        }
        courses = append(courses, *course)
    }
    return &courses, nil
}
