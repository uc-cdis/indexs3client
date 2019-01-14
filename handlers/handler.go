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
	if err := json.Unmarshal([]byte(os.Getenv("CONFIG_FILE")), indexdInfo); err != nil {
		return nil, errors.New("Enviroiment variable CONFIG_FILE is not set correctly")
	}
	return indexdInfo, nil
}

// IndexS3Object indexes s3 object
// The fuction does several things. It first downloads the object from
// S3, computes size and hashes, and update indexd
func IndexS3Object(s3objectURL string) {

	u, err := url.Parse(s3objectURL)
	if err != nil {
		log.Printf("Wrong url format %s\n", s3objectURL)
		return
	}
	bucket, key := u.Host, u.Path

	// key looks like one of these:
	//
	//     <uuid>/<filename>
	//     <dataguid>/<uuid>/<filename>
	//
	// we want to keep the `<dataguid>/<uuid>` part
	split_key := strings.Split(key, "/")
	var uuid string
	if len(split_key) == 2 {
		uuid = split_key[0]
	} else {
		uuid = strings.Join(split_key[:len(split_key)-1], "/")
	}

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

	hashes, err := CalculateBasicHashes(bytes.NewReader(buff))
	if err != nil {
		log.Printf("Can not compute hashes. Detail %s\n\n", err)
		return
	}

	indexdInfo, _ := getIndexServiceInfo()
	rev, err := GetIndexdRecordRev(uuid, indexdInfo.URL)
	if err != nil {
		log.Println(err)
		return
	}

	body := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		len(buff), s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	resp, err := UpdateIndexdRecord(uuid, rev, indexdInfo, []byte(body))
	if err != nil {
		log.Println(err)
	}

	log.Printf("Finish updating the record. Response Status: %s", resp.Status)

}
