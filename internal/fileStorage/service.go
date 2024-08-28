package fileStorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	gcloud "cloud.google.com/go/storage"
	"firebase.google.com/go/storage"
)

type StorageModel struct {
	ST *storage.Client
}

func (s *StorageModel) UploadFile(ctx context.Context, file multipart.File, handler multipart.FileHeader, path string) (string, *gcloud.ObjectHandle, error) {
	bkt, err := s.ST.DefaultBucket()
	if err != nil {
		return "", nil, err
	}
	// Upload the file to Firebase Storage
	object := bkt.Object(path)
	wc := object.NewWriter(ctx)
	defer wc.Close()
	// Copy the uploaded file's content to Firebase storage
	if _, err := io.Copy(wc, file); err != nil {
		object.Delete(ctx)
		return "", nil, err
	}
	expiration := time.Now().Add(time.Hour * 8640)
	opts := &gcloud.SignedURLOptions{
		Expires: expiration,
		Method:  http.MethodGet,
	}
	url, err := bkt.SignedURL(path, opts)
	if err != nil {
		object.Delete(ctx)
		return "", nil, fmt.Errorf("couldn't get signed url: %v", err)
	}
	if url == "" {
		object.Delete(ctx)
		return "", nil, errors.New("empty photo signed file url")
	}

	return url, object, nil
}
