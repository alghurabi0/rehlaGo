package models

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type User struct {
	ID                string `firestore:"-"`
	Firstname         string `firestore:"firstname"`
	Lastname          string `firestore:"lastname"`
	PhoneNumber       string `firestore:"phone_number"`
	ParentPhoneNumber string `firestore:"parent_phone_number"`
	Pwd               string `firestore:"pwd"`
	// --
	Verified      bool     `firestore:"verified"`
	Subscriptions []string `firestore:"subscriptions"`
	ImgURL        string   `firestore:"img_url"`
	ImgPath       string   `firestore:"img_path"`
	NumSubs       int      `firestore:"-"`
}

type UserModel struct {
	DB *firestore.Client
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

func (u *UserModel) Create(ctx context.Context, user *User) (string, error) {
	doc, _, err := u.DB.Collection("users").Add(ctx, user)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

func (u *UserModel) Update(ctx context.Context, userId string, updates []firestore.Update) error {
	_, err := u.DB.Collection("users").Doc(userId).Update(ctx, updates)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAll(ctx context.Context, docRef *firestore.DocumentRef) error {
	subcollections := docRef.Collections(ctx)
	for {
		subcolRef, err := subcollections.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}

		err = DeleteCollection(ctx, subcolRef)
		if err != nil {
			return err
		}
	}
	_, err := docRef.Delete(ctx)
	return err
}

func DeleteCollection(ctx context.Context, colRef *firestore.CollectionRef) error {
	iter := colRef.Documents(ctx)
	for {
		docSnapshot, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}
		err = DeleteAll(ctx, docSnapshot.Ref)
		if err != nil {
			return err
		}
	}
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
