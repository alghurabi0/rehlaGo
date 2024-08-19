package dashboard_models

import (
	"context"

	"cloud.google.com/go/firestore"
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
