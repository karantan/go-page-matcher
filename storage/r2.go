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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// client creates and returns a new S3 client instance configured for
// Cloudflare R2 storage. It fetches necessary credentials and account ID
// from environment variables and sets up the custom endpoint resolver
// for Cloudflare's R2 storage.
func client() *s3.Client {
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

	return s3.NewFromConfig(cfg)

}

// Upload uploads a local file to Cloudflare R2 Storage.
//
// Parameters:
//   - objectKey: The key (path) at which the object will be stored in the bucket.
//   - fileName: Path to the local file which needs to be uploaded.
//
// Returns:
//   - error: An error if encountered during the upload process.
//
// Usage:
//
//	err := storage.Upload("path/in/bucket", "/local/path/to/file")
//	if err != nil {
//	    log.Fatalf("Upload error: %v", err)
//	}
func Upload(objectKey, fileName string) error {
	bucketName := os.Getenv("CF_BUCKET_NAME")

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	c := client()
	_, err = c.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}

// GetPresignURL generates a pre-signed URL to access an object in Cloudflare R2 Storage.
// This URL can be shared and provides temporary access to the specified object.
//
// Parameters:
//   - objectKey: The key (path) of the object for which the pre-signed URL needs to be generated.
//
// Returns:
//   - string: The pre-signed URL which provides access to the specified object.
//   - error: An error if encountered during the URL generation process.
//
// Usage:
//
//	url, err := storage.GetPresignURL("path/in/bucket")
//	if err != nil {
//	    log.Fatalf("URL generation error: %v", err)
//	}
//	fmt.Println("Pre-signed URL:", url)
func GetPresignURL(objectKey string) (string, error) {
	bucketName := os.Getenv("CF_BUCKET_NAME")
	c := client()
	presignClient := s3.NewPresignClient(c)
	presignedURL, err := presignClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(time.Hour*168)) // Max 7 days
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}
