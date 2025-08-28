package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *s3.Client
	bucket string
}

func (s *S3) Save(id string, obj io.WriterTo) (string, error) {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()
		_, err := obj.WriteTo(pipeWriter)
		if err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	uploader := manager.NewUploader(s.client)

	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &id,
		Body:   pipeReader,
	})
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}

	return fmt.Sprintf("s3://%s/%s", s.bucket, id), nil
}

func NewS3(bucket, region string) *S3 {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile("test"),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &S3{s3.NewFromConfig(cfg), bucket}
}

// List returns a list of all files stored with Save.
func (s *S3) List() []string {
	output, err := s.client.ListObjectsV2(
		context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, object := range output.Contents {
		files = append(files, aws.ToString(object.Key))
	}

	return files
}
