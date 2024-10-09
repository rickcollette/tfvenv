// remotesnap.go

package snaps

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// RemoteSnapConfig stores the configuration for remote snap handling.
type RemoteSnapConfig struct {
	Endpoint string
	Auth     string
	Type     string
}

// initS3Client initializes an S3 client using the provided credentials and region.
func initS3Client(accessKey, secretKey, region string) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing AWS session: %v", err)
	}
	return s3.New(sess), nil
}

// GetS3Bucket retrieves the S3 bucket name from the environment variable REMOTE_SNAP_BUCKET.
// Returns an error if the bucket name is not set.
func GetS3Bucket() (string, error) {
	bucketName := os.Getenv("REMOTE_SNAP_BUCKET")
	if bucketName == "" {
		return "", fmt.Errorf("s3 bucket name not provided")
	}
	return bucketName, nil
}

// SanitizeSnapName ensures that snapName does not contain directory traversal characters.
func SanitizeSnapName(snapName string) (string, error) {
	cleanName := filepath.Base(snapName)
	if cleanName != snapName {
		return "", fmt.Errorf("invalid snap name: contains path traversal characters")
	}
	return cleanName, nil
}

// GetRemoteSnap fetches a snap from the remote S3 storage using context for cancellation and timeouts.
func GetRemoteSnap(ctx context.Context, snapName, accessKey, secretKey, region string) ([]byte, error) {
	s3Client, err := initS3Client(accessKey, secretKey, region)
	if err != nil {
		return nil, fmt.Errorf("error initializing S3 client: %v", err)
	}

	bucketName, err := GetS3Bucket()
	if err != nil {
		return nil, fmt.Errorf("error retrieving S3 bucket name: %v", err)
	}

	// Use GetObjectWithContext to pass the context
	result, err := s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(snapName),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving snap from S3: %v", err)
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading snap data: %v", err)
	}

	return buf.Bytes(), nil
}

// SaveRemoteSnap uploads a snap to the remote S3 storage using context for cancellation and timeouts.
func SaveRemoteSnap(ctx context.Context, snapName string, snapData []byte, accessKey, secretKey, region string) error {
	s3Client, err := initS3Client(accessKey, secretKey, region)
	if err != nil {
		return fmt.Errorf("error initializing S3 client: %v", err)
	}

	bucketName, err := GetS3Bucket()
	if err != nil {
		return fmt.Errorf("error retrieving S3 bucket name: %v", err)
	}

	// Use PutObjectWithContext to pass the context
	_, err = s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(snapName),
		Body:   bytes.NewReader(snapData),
	})
	if err != nil {
		return fmt.Errorf("error uploading snap to S3: %v", err)
	}
	return nil
}

// ListRemoteSnaps lists all snaps stored in the remote S3 bucket.
// It returns a slice of snap names or an error if the operation fails.
// ListRemoteSnaps lists all snaps stored in the remote S3 bucket.
func ListRemoteSnaps(ctx context.Context, accessKey, secretKey, region string) ([]string, error) {
	s3Client, err := initS3Client(accessKey, secretKey, region)
	if err != nil {
		return nil, fmt.Errorf("error initializing S3 client: %v", err)
	}

	bucketName, err := GetS3Bucket()
	if err != nil {
		return nil, fmt.Errorf("error retrieving S3 bucket name: %v", err)
	}

	result, err := s3Client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing snaps in S3 bucket: %v", err)
	}

	var snapsList []string
	for _, item := range result.Contents {
		snapsList = append(snapsList, *item.Key)
	}
	return snapsList, nil
}
// RemoveRemoteSnap deletes a snap from the remote S3 storage using context for cancellation and timeouts.
// It removes the snap identified by snapName using the provided AWS credentials and region.
func RemoveRemoteSnap(ctx context.Context, snapName, accessKey, secretKey, region string) error {
	s3Client, err := initS3Client(accessKey, secretKey, region)
	if err != nil {
		return fmt.Errorf("error initializing S3 client: %v", err)
	}

	bucketName, err := GetS3Bucket()
	if err != nil {
		return fmt.Errorf("error retrieving S3 bucket name: %v", err)
	}

	// Use DeleteObjectWithContext to pass the context
	_, err = s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(snapName),
	})
	if err != nil {
		return fmt.Errorf("error deleting snap from S3: %v", err)
	}
	return nil
}
