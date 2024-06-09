package models

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	//"google.golang.org/api/iterator"
)

type Contact struct {
    Fullname string `firestore:"full_name"`
    Phone_number string `firestore:"phone_number"`
    Message string `firestre:"message"`
}

type ContactInfo struct {
    Email string `firestore:"email"`
    Phone string `firestore:"phone_number"`
    Location string `firestore:"location"`
    Instagram string `firestore:"instagram"`
    Facebook string `firestore:"facebook"`
}

type ContactModel struct {
	DB *firestore.Client
}

func (c *ContactModel) GetContactInfo(ctx context.Context) (*ContactInfo, error) {
    contactDoc, err := c.DB.Collection("contact").Doc("contactInfo").Get(ctx)
    if err != nil {
        return nil, errors.New("no contactInfo doc in contact collection")
    }
    var info ContactInfo
    err = contactDoc.DataTo(info)
    if err != nil {
        return &ContactInfo{}, err
    }
    return &info, nil
}

func (c *ContactModel) SendInquiry(ctx context.Context, fullname, phone_number, message string) (error) {
    data := Contact{
        Fullname: fullname,
        Phone_number: phone_number,
        Message: message,
    }
    newDoc := c.DB.Collection("contact").NewDoc()
    _, err := newDoc.Set(ctx, data)
    if err != nil {
        return err
    }
    return nil
}
