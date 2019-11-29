package handlers

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// AWS sesssion wrapper
type AwsClient struct {
	config  *aws.Config
	session *session.Session
}

// CreateNewSession creates an aws s3 session
func CreateNewAwsClient() (*AwsClient, error) {
	client := new(AwsClient)
	client.config = &aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}

	sess, err := session.NewSession(client.config)
	if err != nil {
		return nil, err
	}

	client.session = sess
	return client, nil
}

// Creates AWS Session if it's nil (embedded entry point)
func (client *AwsClient) init() error {
	if client.session == nil {
		sess, err := session.NewSession(client.config)
		if err != nil {
			return err
		}
		client.session = sess
	}
	return nil
}

func (client *AwsClient) GetS3BucketOwner(bucket string) (string, error) {

	if iErr := client.init(); iErr != nil {
		return "", iErr
	}

	svc := s3.New(client.session)
	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}

	result, err := svc.GetBucketAcl(input)
	if err != nil {
		return "", err
	}

	return *result.Owner.DisplayName, nil
}

// GetS3ObjectOutput gets object output from s3
func (client *AwsClient) GetS3ObjectOutput(bucket string, key string) (*s3.GetObjectOutput, error) {

	if iErr := client.init(); iErr != nil {
		return nil, iErr
	}

	svc := s3.New(client.session)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result, nil

}

// GetObjectSize returns object size in bytes
func (client *AwsClient) GetObjectSize(bucket string, key string) (*int64, error) {

	if iErr := client.init(); iErr != nil {
		return nil, iErr
	}

	svc := s3.New(client.session)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result.ContentLength, nil
}
