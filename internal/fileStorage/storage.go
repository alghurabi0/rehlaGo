package fileStorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	gcloud "cloud.google.com/go/storage"
	"firebase.google.com/go/storage"
	"google.golang.org/api/iterator"
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

func (s *StorageModel) DeleteFile(ctx context.Context, path string) error {
	bkt, err := s.ST.DefaultBucket()
	if err != nil {
		return fmt.Errorf("failed to get default bucket: %v", err)
	}

	object := bkt.Object(path)
	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

func (s *StorageModel) GetAnswers(ctx context.Context, courseId, examId string) ([]string, error) {
	bkt, err := s.ST.DefaultBucket()
	if err != nil {
		return nil, fmt.Errorf("failed to get default bucket: %v", err)
	}
	path := fmt.Sprintf("courses/%s/exams/%s/answers/", courseId, examId)

	query := &gcloud.Query{
		Prefix: path,
	}
	objectsIter := bkt.Objects(ctx, query)

	var answers []string
	for {
		obj, err := objectsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("couldn't iterate over objects, error: %v", err)
		}
		objName := strings.TrimPrefix(obj.Name, path)
		answers = append(answers, objName)
	}

	return answers, nil
}
