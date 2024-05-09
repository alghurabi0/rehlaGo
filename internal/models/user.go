package models

import (
	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

type User struct {
	ID string `firestore:"-"`
}

type UserModel struct {
	DB   *firestore.Client
	Auth *auth.Client
}
