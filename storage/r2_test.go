package storage

import (
	"bytes"
	"context"
	"os"
	mock_storage "page_matcher/mocks"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpload(t *testing.T) {
	myBackup, _ := os.CreateTemp("", "myBackup.txt")
	defer os.Remove(myBackup.Name())
	myBackup.WriteString(`this is the body`)
	myBackup.Close()

	ctrl := gomock.NewController(t)

	m := mock_storage.NewMockS3Client(ctrl)

	ctx := context.TODO()
	bucket := "bucket"
	os.Setenv("CF_BUCKET_NAME", bucket)
	key := "objectKey"

	expectedBody := []byte("this is the body")
	fakeBucket := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(expectedBody),
	}

	m.EXPECT().PutObject(ctx, fakeBucket)
	Upload(m, "objectKey", myBackup.Name())
}

func TestGetPresignURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mock_storage.NewMockPresignClient(ctrl)

	bucket := "bucket"
	os.Setenv("CF_BUCKET_NAME", bucket)
	key := "key"

	ctx := context.Background()
	fakeBucket := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	fakeRequest := &v4.PresignedHTTPRequest{
		URL: "http://foo.com",
	}
	// optFns is a variadic parameter and it can't be easily mocked. use gomock.Any
	m.EXPECT().PresignGetObject(ctx, fakeBucket, gomock.Any()).Return(fakeRequest, nil)
	assert := assert.New(t)
	got, err := GetPresignURL(m, key)
	assert.NoError(err)
	assert.Equal("http://foo.com", got)
}
