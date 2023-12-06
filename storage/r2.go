// Package storage provides functionalities to interact with
// Cloudflare R2 Storage using the AWS SDK. This package provides
// utility functions to upload files and generate pre-signed URLs.
//
// Usage of this package requires specific environment variables to be set:
//   - CF_ACCOUNT_ID: Represents the Cloudflare Account ID.
//   - CF_ACCESS_KEY_ID: Represents the access key ID for authentication.
//   - CF_ACCESS_KEY_SECRET: Represents the access key secret for authentication.
//   - CF_BUCKET_NAME: Represents the bucket name in Cloudflare R2 Storage.
//
// For more details on Cloudflare R2 Storage with AWS SDK, refer to:
// https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/
package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type PresignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type Bucket struct {
	S3Client      S3Client
	PresignClient PresignClient
}

// NewR2Client creates and returns a new Bucket with S3 client and presigned client
// instance configured for Cloudflare R2 storage. It fetches necessary credentials
// and account ID from environment variables and sets up the custom endpoint resolver
// for Cloudflare's R2 storage.
func NewR2Client() Bucket {
	accountID := os.Getenv("CF_ACCOUNT_ID")
	accessKeyID := os.Getenv("CF_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("CF_ACCESS_KEY_SECRET")

	if accountID == "" {
		log.Fatalf("Missing CF_ACCOUNT_ID env. var.")
	}
	if accessKeyID == "" {
		log.Fatalf("Missing CF_ACCESS_KEY_ID env. var.")
	}
	if accessKeySecret == "" {
		log.Fatalf("Missing CF_ACCESS_KEY_SECRET env. var.")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
	)
	if err != nil {
		log.Fatal(err)
	}
	c := s3.NewFromConfig(cfg)
	return Bucket{S3Client: c, PresignClient: s3.NewPresignClient(c)}
}

// Upload uploads a local file to Cloudflare R2 Storage.
func Upload(s3Client S3Client, objectKey, fileName string) error {
	bucketName := os.Getenv("CF_BUCKET_NAME")

	body, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(body),
	})
	return err
}

// GetPresignURL generates a pre-signed URL to access an object in Cloudflare R2 Storage.
// This URL can be shared and provides temporary access to the specified object.
func GetPresignURL(client PresignClient, objectKey string) (string, error) {
	ctx := context.Background()
	bucketName := os.Getenv("CF_BUCKET_NAME")
	bucketObject := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	options := func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(time.Hour * 168) // Max 7 days
	}
	request, err := client.PresignGetObject(ctx, bucketObject, options)
	if err != nil {
		return "", fmt.Errorf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request.URL, nil
}
