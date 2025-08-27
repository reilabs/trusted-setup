package storage_test

import (
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/storage"
)

type testWriterTo struct {
	content []byte
}

func (e *testWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(e.content)
	return int64(n), err
}

func TestS3(t *testing.T) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	srv := httptest.NewServer(faker.Server())
	defer srv.Close()

	credentials, _ := os.CreateTemp("", "")
	_, _ = credentials.Write(
		[]byte(
			"[default]\n" +
				"aws_access_key_id = <YOUR_ACCESS_KEY_ID>\n" +
				"aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>"),
	)
	_ = credentials.Close()

	s, err := storage.NewS3(
		storage.WithBucket("test-bucket"),
		storage.WithRegion("us-east-1"),
		storage.WithProfile("default"),
		storage.WithCredentialsFile(credentials.Name()),
		storage.WithEndpoint(srv.URL), // Use mock S3 service
	)
	assert.NoError(t, err)

	const testBucket = "test-bucket"
	const testFile = "test"
	obj := &testWriterTo{content: []byte("Hello world")}

	// Bucket does not exist
	filePath, err := s.Save(testFile, obj)
	assert.Error(t, err)
	assert.Empty(t, filePath)

	// Bucket exists
	err = backend.CreateBucket(testBucket)
	assert.NoError(t, err)
	filePath, err = s.Save(testFile, obj)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("s3://%s/%s", testBucket, testFile), filePath)

	// Object is visible
	objs, err := s.List()
	assert.NoError(t, err)
	assert.Len(t, objs, 1)
	assert.Equal(t, objs[0], filePath)

	// Removing objects with other means is reflected in List return value
	_, err = backend.DeleteObject(testBucket, testFile)
	assert.NoError(t, err)
	objs, err = s.List()
	assert.NoError(t, err)
	assert.Empty(t, objs)

	// Saving objects is immediately reflected in List return value
	for i := 0; i < 5; i++ {
		path, err := s.Save(fmt.Sprintf("testfile%d", i), obj)
		assert.NoError(t, err)
		assert.NotEmpty(t, path)

		objs, err = s.List()
		assert.NoError(t, err)
		assert.Len(t, objs, i+1)
	}

	_ = os.Remove(credentials.Name())
}
