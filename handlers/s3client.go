package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// AWS sesssion wrapper
type AwsClient struct {
	session *session.Session
}

// JobConfig
type JobConfig struct {
	IndexObject map[string]interface{} `json:"index-object"`
}

// CreateNewSession creates an aws s3 session
func CreateNewAwsClient() (*AwsClient, error) {
	client := new(AwsClient)

	var sess *session.Session
	var err error
	var region string
	var awsAccessKeyID string
	var awsSecretAccessKey string

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		region = os.Getenv("AWS_REGION")
		awsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	} else {
		fmt.Println("Try to get credential from the secret")
		buff, err := ioutil.ReadFile("/creds.json")
		if err == nil {
			jobConfig := new(JobConfig)
			if err = json.Unmarshal(buff, &jobConfig); err != nil {
				return nil, errors.New("Wrong job config detected!")
			}
			jobRequired := jobConfig.IndexObject["job_requires"].(map[string]interface{})
			region = jobRequired["region"].(string)
			awsAccessKeyID = jobRequired["aws_access_key_id"].(string)
			awsSecretAccessKey = jobRequired["aws_secret_access_key"].(string)
		}
	}
	// Create aws session with provided aws key pair
	if awsAccessKeyID != "" {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKeyID, awsSecretAccessKey, ""),
		})
	} else {
		// No credential provided, just use default credential
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		sess, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	}

	if err != nil {
		return nil, err
	}
	client.session = sess
	return client, nil
}

// GetS3ObjectOutput gets object output from s3
func GetS3ObjectOutput(client *AwsClient, bucket string, key string) (*s3.GetObjectOutput, error) {

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
func GetObjectSize(client *AwsClient, bucket string, key string) (*int64, error) {
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
