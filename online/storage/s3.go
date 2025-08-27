package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client   *s3.Client
	uploader *manager.Uploader
	bucket   string
}

type S3Option func(*s3Options)

type s3Options struct {
	bucket          string
	region          string
	profile         string
	credentialsFile string
	endpoint        string
}

// WithBucket sets the S3 bucket
func WithBucket(bucket string) S3Option {
	return func(o *s3Options) {
		o.bucket = bucket
	}
}

// WithRegion sets the AWS region
func WithRegion(region string) S3Option {
	return func(o *s3Options) {
		o.region = region
	}
}

// WithProfile sets the AWS profile name
func WithProfile(profile string) S3Option {
	return func(o *s3Options) {
		o.profile = profile
	}
}

// WithCredentialsFile sets the AWS credentials file path
func WithCredentialsFile(credentialsFile string) S3Option {
	return func(o *s3Options) {
		o.credentialsFile = credentialsFile
	}
}

// WithEndpoint sets the custom endpoint for AWS service
func WithEndpoint(endpoint string) S3Option {
	return func(o *s3Options) {
		o.endpoint = endpoint
	}
}

// NewS3 creates a new instance of the storage backed by AWS S3.
//
// Bucket, profile, region, credentials file and endpoint can be optionally customized using With### functions.
// Otherwise, the default values are used based on the environment variables,
// shared configuration and shared credentials files.
//
// AWS S3 endpoint can be overridden with the endpoint parameter. For default endpoint use S3DefaultEndpoint.
func NewS3(opts ...S3Option) (*S3, error) {
	options := &s3Options{}

	for _, opt := range opts {
		opt(options)
	}

	var cfgOpts []func(*config.LoadOptions) error

	if options.endpoint != "" {
		cfgOpts = append(cfgOpts, config.WithBaseEndpoint(options.endpoint))
	}
	if options.region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(options.region))
	}
	if options.profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(options.profile))
	}
	if options.credentialsFile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedCredentialsFiles([]string{options.credentialsFile}))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), cfgOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	return &S3{
		client:   client,
		uploader: manager.NewUploader(client),
		bucket:   options.bucket,
	}, nil
}

// Save stores the object in the S3 bucket.
//
// The given id is prepended to the file name of the file. If the file already exists, it will
// be overwritten.
//
// The function returns the path to the file and an error if any.
func (s *S3) Save(id string, obj io.WriterTo) (string, error) {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer func(pipeWriter *io.PipeWriter) {
			err := pipeWriter.Close()
			if err != nil {
				return
			}
		}(pipeWriter)
		_, err := obj.WriteTo(pipeWriter)
		if err != nil {
			err = pipeWriter.CloseWithError(err)
			if err != nil {
				return
			}
		}
	}()

	_, err := s.uploader.Upload(
		context.TODO(), &s3.PutObjectInput{
			Bucket: &s.bucket,
			Key:    &id,
			Body:   pipeReader,
		},
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("s3://%s/%s", s.bucket, id), nil
}

// List returns a list of all files stored with Save.
//
// The function reflects the actual state of the storage. I.e. if the file stored with Save
// is removed with AWS S3 CLI, List will not return it.
//
// Returns an array of strings where each element is a path in the format: s3://<bucket name>/<object id>
// and error, if occurred.
func (s *S3) List() ([]string, error) {
	output, err := s.client.ListObjectsV2(
		context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
		},
	)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, object := range output.Contents {
		files = append(files, fmt.Sprintf("s3://%s/%s", s.bucket, aws.ToString(object.Key)))
	}

	return files, nil
}
