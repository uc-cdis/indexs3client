package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

type IndexdInfo struct {
	URL      string `url`
	Username string `username`
	Password string `password`
}

func getIndexServiceInfo() (*IndexdInfo, error) {
	indexdInfo := new(IndexdInfo)
	if err := json.Unmarshal([]byte(os.Getenv("IMAGE_CONFIG")), indexdInfo); err != nil {
		return nil, errors.New("Enviroiment variable IMAGE_CONFIG is not set correctly")
	}
	return indexdInfo, nil
}

// IndexS3Object indexes s3 object
// The fuction does several things. It first downloads the object from
// S3, computes size and hashes, and update indexd
func IndexS3Object(s3objectURL string) {

	u, _ := url.Parse(s3objectURL)
	bucket, key := u.Host, u.Path
	uuid := strings.Split(key, "/")[0]

	client, err := CreateNewAwsClient()
	if err != nil {
		log.Printf("Can not create AWS client. Detail %s\n\n", err)
		return
	}

	buff, err := StreamObjectFromS3(client, bucket, key)
	if err != nil {
		log.Printf("Can not download file. Detail %s\n\n", err)
		return
	}

	hashes := CalculateBasicHashes(bytes.NewReader(buff))

	indexdInfo, _ := getIndexServiceInfo()
	rev, err := GetIndexdRecordRev(uuid, indexdInfo.URL)
	if err != nil {
		log.Println(err)
	}

	body := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		len(buff), s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	resp, err := UpdateIndexdRecord(uuid, rev, indexdInfo.URL, indexdInfo.Username, indexdInfo.Password, []byte(body))
	if err != nil {
		log.Println(err)
	}

	log.Printf("Response Status: %s", resp.Status)

}
