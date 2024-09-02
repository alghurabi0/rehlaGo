package models

import (
	"context"
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
	ImgURL            string   `firestore:"img_url"`
	NumSubs           int      `firestore:"-"`
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
	user.NumSubs = len(user.Subscriptions)
	return &user, nil
}

func (u *UserModel) GetAll(ctx context.Context, offset int) (*[]User, error) {
	usersIter := u.DB.Collection("users").Offset(offset).Documents(ctx)
	var users []User
	for {
		doc, err := usersIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var user User
		if err := doc.DataTo(&user); err != nil {
			return nil, err
		}
		user.ID = doc.Ref.ID
		user.NumSubs = len(user.Subscriptions)
		users = append(users, user)
	}
	return &users, nil
}

//func (u *UserModel) CheckUserExists(ctx context.Context, phone string) error {
//user, err := u.Auth.GetUserByPhoneNumber(ctx, phone)
//print(user)
//if err != nil {
//return nil
//}
//return errors.New("user already exists")
//}

func (u *UserModel) Create(ctx context.Context, firstname, lastname, phone, parentPhone, pwd string) (string, error) {
	userData := User{
		Firstname:         firstname,
		Lastname:          lastname,
		PhoneNumber:       phone,
		ParentPhoneNumber: parentPhone,
		Pwd:               pwd,
	}
	doc, _, err := u.DB.Collection("users").Add(ctx, userData)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
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
		err = doc.DataTo(&user)
		if err != nil {
			fmt.Print(err)
		}
		user.ID = doc.Ref.ID
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
