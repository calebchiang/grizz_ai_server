package services

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadVideoToR2(file *multipart.FileHeader) (string, error) {

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(src)

	filename := fmt.Sprintf("drills/%d-%s.mp4", time.Now().Unix(), uuid.New())

	accessKey := os.Getenv("R2_ACCESS_KEY")
	secretKey := os.Getenv("R2_SECRET_KEY")
	endpoint := os.Getenv("R2_ENDPOINT")
	bucket := os.Getenv("R2_BUCKET")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
	)

	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(buf.Bytes()),
	})

	if err != nil {
		return "", err
	}

	videoURL := publicURL + "/" + filename

	return videoURL, nil
}
