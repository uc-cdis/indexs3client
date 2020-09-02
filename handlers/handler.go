package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

// MaxRetries maximum number of retries
const MaxRetries = 5

type IndexdInfo struct {
	URL      string `url`
	Username string `username`
	Password string `password`
}

func minOf(vars ...int64) int64 {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}
func getIndexServiceInfo() (*IndexdInfo, error) {
	indexdInfo := new(IndexdInfo)
	if os.Getenv("CONFIG_FILE") != "" {
		if err := json.Unmarshal([]byte(os.Getenv("CONFIG_FILE")), indexdInfo); err != nil {
			return nil, errors.New("Enviroiment variable CONFIG_FILE is not set correctly")
		}
	} else {
		buff, err := ioutil.ReadFile("/creds.json")
		if err == nil {
			dataMap := new(JobConfig)
			_ = json.Unmarshal(buff, &dataMap)
			indexdInfo.Username = dataMap.IndexObject["indexd_user"].(string)
			indexdInfo.Password = dataMap.IndexObject["indexd_password"].(string)
			indexdInfo.URL = dataMap.IndexObject["indexd_url"].(string)
		}
	}
	return indexdInfo, nil
}

// IndexS3Object indexes s3 object
// The fuction does several things. It first downloads the object from
// S3, computes size and hashes, and update indexd
func IndexS3Object(s3objectURL string) {
	s3objectURL, _ = url.QueryUnescape(s3objectURL)
	u, err := url.Parse(s3objectURL)
	if err != nil {
		log.Panicf("Wrong url format %s\n", s3objectURL)
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
		log.Panicf("Can not create AWS client. Detail %s\n\n", err)
	}

	log.Printf("Start to compute hashes for %s", key)
	hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)

	if err != nil {
		log.Panicf("Can not compute hashes for %s. Detail %s ", key, err)
	}
	log.Printf("Finish to compute hashes for %s", key)

	indexdInfo, _ := getIndexServiceInfo()

	var retries = 0
	var rev = ""

	for {
		rev, err = GetIndexdRecordRev(uuid, indexdInfo.URL)
		if err != nil {
			retries++
			log.Printf("Error: %s. Retry: %d", err, retries)
			time.Sleep(5 * time.Second)
		} else if rev == "" {
			log.Println("The file already has size and hashes")
			return
		} else {
			break
		}
		if retries == MaxRetries {
			log.Panicf("Can not get record %s from indexd. Error message %s", uuid, err)
		}
	}

	body := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)

	retries = 0
	for {
		resp, err := UpdateIndexdRecord(uuid, rev, indexdInfo, []byte(body))
		if err != nil {
			retries++
			log.Printf("Error: %s. Retry: %d", err, retries)
			time.Sleep(5 * time.Second)
		} else if resp.StatusCode != 200 {
			log.Printf("StatusCode: %d. Retry: %d", resp.StatusCode, retries)
			retries++
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Finish updating the record %s. Response Status: %d. Body %s", uuid, resp.StatusCode, body)
			break
		}

		if retries == MaxRetries {
			if err == nil {
				log.Panicf("Can not update %s with hash info. Body %s. Status code %d. Detail %s", uuid, body, resp.StatusCode, err)
			} else {
				log.Panicf("Can not update %s with hash info. Body %s. Detail %s", uuid, body, err)
			}
			break
		}
	}
}
