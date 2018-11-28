package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Credentials contains AWS credentials
type AWSCredentials struct {
	region             string
	awsAccessKeyID     string
	awsSecretAccessKey string
}

type AwsClient struct {
	session *session.Session
}

// loadCredentialFromConfigFile loads AWS credentials from the config file
func loadCredentialFromConfigFile(path string) (*AWSCredentials, error) {
	credentials := new(AWSCredentials)
	// Read data file
	jsonBytes, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mapping map[string]interface{}
	json.Unmarshal(jsonBytes, &mapping)
	if region, err := GetValueFromDict(mapping, []string{"region"}); err != nil {
		panic("Can not read region from credential file")
	} else {
		credentials.region = region.(string)
	}

	if awsKeyID, err := GetValueFromDict(mapping, []string{"aws_access_key_id"}); err != nil {
		panic("Can not read aws key from credential file")
	} else {
		credentials.awsAccessKeyID = awsKeyID.(string)
	}

	if awsSecret, err := GetValueFromDict(mapping, []string{"aws_secret_access_key"}); err != nil {
		panic("Can not read aws key from credential file")
	} else {
		credentials.awsSecretAccessKey = awsSecret.(string)
	}

	return credentials, nil
}

// createNewSession creats a aws s3 session
func CreateNewAwsClient() (*AwsClient, error) {
	creds, err := loadCredentialFromConfigFile(CREDENTIAL_PATH)
	if err != nil {
		return nill, err
	}

	client := new(AwsClient)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(creds.region),
		Credentials: credentials.NewStaticCredentials(
			creds.awsAccessKeyID, creds.awsSecretAccessKey, ""),
	})
	client.session = sess

	return client, nil
}

// DownloadObjectFromS3 : Download an object from S3
func DownloadObjectFromS3(client *AwsClient, bucket string, key string, filename string) error {
	if client.session == nil {
		log.Printf("Awsclient is not initialized \n\n")
		return err
	}
	// The session the S3 Downloader will use
	sess := client.session

	// Create a downloader with the session and custom options
	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = 64 * 1024 * 1024 // 64MB per part
	})

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %q, %v", filename, err)
	}

	// Write the contents of S3 Object to the file
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err

}
