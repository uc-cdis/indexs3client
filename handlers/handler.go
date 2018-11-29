package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// IndexS3Object indexes s3 object
// The fuction does several things. It first downloads the object from
// S3, computes size and hashes, and update indexd
func IndexS3Object(s3objectURL string, indexURL string) {

	u, _ := url.Parse(s3objectURL)
	bucket, key := u.Host, u.Path
	uuid := strings.Split(key, "/")[0]

	client, err := CreateNewAwsClient()
	if err != nil {
		fmt.Printf("Can not create AWS client. Detail %s\n\n", err)
		return
	}

	buff, err := StreamObjectFromS3(client, bucket, key)
	if err != nil {
		fmt.Printf("Can not download file. Detail %s\n\n", err)
		return
	}

	hashes := CalculateBasicHashes(bytes.NewReader(buff))

	rev, err := GetIndexdRecordRev(uuid, indexURL)
	if err != nil {
		log.Println(err)
	}

	body := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s"}}`, len(buff), s3objectURL, hashes.Md5)
	resp, err := UpdateIndexdRecord(uuid, rev, indexURL, os.Getenv("USERNAME"), os.Getenv("PASSWORD"), []byte(body))
	if err != nil {
		log.Println(err)
	}

	log.Printf("Response Status:", resp.Status)

}
