package handlers

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AwsClient struct {
	session *session.Session
}

// createNewSession creats a aws s3 session
func CreateNewAwsClient() (*AwsClient, error) {
	client := new(AwsClient)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})

	if err != nil {
		return nil, err
	}

	client.session = sess
	return client, nil
}

func StreamObjectFromS3(client *AwsClient, bucket string, key string) ([]byte, error) {
	buff := &aws.WriteAtBuffer{}
	s3dl := s3manager.NewDownloader(client.session)
	_, err := s3dl.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return buff.Bytes(), nil

}
