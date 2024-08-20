package dashboard_models

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type DashboardUser struct {
	ID       string `firestore:"-"`
	Username string `firestore:"username"`
	Role     string `firestore:"role"`
	Password string `firestore:"password"`
}

type DashboardUserModel struct {
	DB *firestore.Client
}

func (u *DashboardUserModel) Get(ctx context.Context, userId string) (*DashboardUser, error) {
	userDoc, err := u.DB.Collection("dashboard_users").Doc(userId).Get(ctx)
	if err != nil {
		return &DashboardUser{}, err
	}
	var user DashboardUser
	err = userDoc.DataTo(&user)
	if err != nil {
		return &DashboardUser{}, err
	}
	user.ID = userDoc.Ref.ID
	return &user, nil
}

func (u *DashboardUserModel) Create(ctx context.Context, username, role, pwd string) (string, error) {
	userData := DashboardUser{
		Username: username,
		Role:     role,
		Password: pwd,
	}
	doc, _, err := u.DB.Collection("dashboard_users").Add(ctx, userData)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func (u *DashboardUserModel) ValidateLogin(ctx context.Context, username, password string) (string, error) {
	query := u.DB.Collection("dashboard_users").Where("username", "==", username)
	iter := query.Documents(ctx)
	var user DashboardUser
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("failed to iterate: %v", err)
			return "", err
		}
		err = doc.DataTo(&user)
		if err != nil {
			return "", err
		}
		user.ID = doc.Ref.ID
	}
	if user.Username != username {
		return "", errors.New("user not found")
	}
	if user.Password != password {
		return "", errors.New("incorrect password")
	}
	return user.ID, nil
}
