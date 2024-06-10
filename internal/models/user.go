package models

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

type User struct {
	ID                string   `firestore:"-"`
	Firstname         string   `firestore:"firstname"`
	Lastname          string   `firestore:"lastname"`
	PhoneNumber       string   `firestore:"phone_number"`
	ParentPhoneNumber string   `firestore:"parent_phone_number"`
	Pwd               string   `firestore:"pwd"`
	Subscriptions     []string `firestore:"Subscriptions"`
}

type UserModel struct {
	DB   *firestore.Client
	Auth *auth.Client
}

func (u *UserModel) Get(ctx context.Context, userId string) (*User, error) {
	userDoc, err := u.DB.Collection("users").Doc(userId).Get(ctx)
	if err != nil {
		return &User{}, err
	}
	var user User
	err = userDoc.DataTo(&user)
	if err != nil {
		print(err)
		return &User{}, err
	}
	user.ID = userDoc.Ref.ID
	return &user, nil
}

func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	id := fmt.Sprintf("%x", b)
	return id
}

func (u *UserModel) CheckUserExists(ctx context.Context, phone string) error {
	user, err := u.Auth.GetUserByPhoneNumber(ctx, phone)
	print(user)
	if err != nil {
		return nil
	}
	return errors.New("user already exists")
}

func (u *UserModel) Create(ctx context.Context, firstname, lastname, phone, parentPhone, pwd string) (string, string, error) {
	user, err := u.Auth.GetUserByPhoneNumber(ctx, phone)
	if err != nil {
		return "", "", err
	}
	userId := user.UID
	sessionId := generateRandomID()
	if sessionId == "" {
		return "", "", errors.New("could not generate session id")
	}
	userData := User{
		Firstname:         firstname,
		Lastname:          lastname,
		PhoneNumber:       phone,
		ParentPhoneNumber: parentPhone,
		Pwd:               pwd,
	}
	_, err = u.DB.Collection("users").Doc(userId).Set(ctx, userData)
	if err != nil {
		return "", "", err
	}
	return userId, sessionId, nil
}

func (u *UserModel) VerifySessionId(ctx context.Context, userId, sessionId string) error {
	doc, err := u.DB.Collection("users").Doc(userId).Get(ctx)
	if err != nil {
		return err
	}
	var user User
	doc.DataTo(&user)
	return nil
}

func (u *UserModel) ValidateLogin(ctx context.Context, phone, pass string) (*User, error) {
	query := u.DB.Collection("users").Where("phone_number", "==", phone)
	iter := query.Documents(ctx)
    var user User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
            log.Fatalf("failed to iterate: %v", err)
		}
        fmt.Print(doc.Data())
        err = doc.DataTo(&user)
        if err != nil {
            fmt.Print(err)
        }
        fmt.Print(user)
	}
    fmt.Print(phone)
    fmt.Print(pass)
	if user.PhoneNumber != phone {
		return &User{}, errors.New("user not found")
	}
	if user.Pwd != pass {
		return &User{}, errors.New("incorrect password")
	}
    return &user, nil
}
