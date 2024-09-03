package models

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Subscription struct {
	ID          string    `firestore:"-"`
	CourseTitle string    `firestore:"course_title"`
	Active      bool      `firestore:"active"`
	Answers     *[]Answer `firestore:"-"`
}

type SubscriptionModel struct {
	DB *firestore.Client
}

func (s *SubscriptionModel) Get(ctx context.Context, userId, subId string) (*Subscription, error) {
	subDoc, err := s.DB.Collection("users").Doc(userId).Collection("subs").Doc(subId).Get(ctx)
	if err != nil {
		return &Subscription{}, err
	}
	var sub Subscription
	err = subDoc.DataTo(&sub)
	if err != nil {
		return &Subscription{}, err
	}
	sub.ID = subDoc.Ref.ID
	return &sub, nil
}

func (s *SubscriptionModel) GetAll(ctx context.Context, userId string) (*[]Subscription, error) {
	subIterator := s.DB.Collection("users").Doc(userId).Collection("subs").Documents(ctx)
	var subs []Subscription
	for {
		doc, err := subIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sub Subscription
		if err := doc.DataTo(&sub); err != nil {
			return nil, err
		}
		sub.ID = doc.Ref.ID
		subs = append(subs, sub)
	}
	return &subs, nil
}

func (s *SubscriptionModel) Create(ctx context.Context, userId string, sub *Subscription) (string, error) {
	doc, _, err := s.DB.Collection("users").Doc(userId).Collection("subs").Add(ctx, sub)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func (s *SubscriptionModel) IsActive(ctx context.Context, userId, subId string) bool {
	sub, err := s.Get(ctx, userId, subId)
	if err != nil {
		return false
	}
	if sub.Active {
		return true
	}
	return false
}
