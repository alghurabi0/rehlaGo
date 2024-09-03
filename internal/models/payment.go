package models

import (
	"context"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Payment struct {
	ID            string    `firestore:"-"`
	UserId        string    `firestore:"-"`
	SubId         string    `firestore:"-"`
	AmountPaid    int       `firestore:"amount_paid"`
	DateOfPayment time.Time `firestore:"date_of_payment"`
	ValidUntil    time.Time `firestore:"valid_until"`
}

type PaymentModel struct {
	DB *firestore.Client
}

func (p *PaymentModel) Get(ctx context.Context, userId, subId, payId string) (*Payment, error) {
	payDoc, err := p.DB.Collection("users").Doc(userId).Collection("subs").Doc(subId).Collection("payments").Doc(payId).Get(ctx)
	if err != nil {
		return &Payment{}, err
	}
	var payment Payment
	err = payDoc.DataTo(&payment)
	if err != nil {
		return &Payment{}, err
	}
	payment.ID = payDoc.Ref.ID
	payment.UserId = userId
	payment.SubId = subId
	return &payment, nil
}

func (p *PaymentModel) GetAll(ctx context.Context, userId, subId string) (*[]Payment, error) {
	payIterator := p.DB.Collection("users").Doc(userId).Collection("subs").Doc(subId).Collection("payments").Documents(ctx)
	var payments []Payment
	for {
		doc, err := payIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var payment Payment
		if err := doc.DataTo(&payment); err != nil {
			return nil, err
		}
		payment.ID = doc.Ref.ID
		payment.UserId = userId
		payment.SubId = subId
		payments = append(payments, payment)
	}
	sort.Slice(payments, func(i, j int) bool {
		return payments[i].DateOfPayment.After(payments[j].DateOfPayment)
	})
	return &payments, nil
}
