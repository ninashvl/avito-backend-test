package reports

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	linkTTL = time.Hour * 24
)

type Repo struct {
	bucket string
	client *minio.Client
}

func NewRepo(bucket string, client *minio.Client) *Repo {
	return &Repo{
		bucket: bucket,
		client: client,
	}
}

func (r *Repo) SaveReportFile(ctx context.Context, file bytes.Buffer) (string, error) {
	filename := fmt.Sprintf("%s.csv", uuid.New().String())

	_, err := r.client.PutObject(ctx, r.bucket, filename, &file, int64(file.Len()),
		minio.PutObjectOptions{
			//RetainUntilDate: time.Now().Add(fileTTL),
			ContentType: "application/csv",
		})
	if err != nil {
		return "", fmt.Errorf("error saving report: %v", err)
	}

	reqParams := make(url.Values)
	reportUrl, err := r.client.PresignedGetObject(ctx, r.bucket, filename, linkTTL, reqParams)
	if err != nil {
		return "", fmt.Errorf("error preparing report link: %v", err)
	}

	return reportUrl.String(), nil
}
