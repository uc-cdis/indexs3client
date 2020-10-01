package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"log"
	"net/url"
	"os"
	"strings"
)

// MaxRetries maximum number of retries
const MaxRetries = 5

type ConfigInfo struct {
    Indexd IndexdInfo `indexd`
    MetadataService MetadataServiceInfo `metadata_service`
}

type IndexdInfo struct {
	URL      string `url`
	Username string `username`
	Password string `password`
}

type MetadataServiceInfo struct {
    URL      string `url`
    Username string `username`
    Password string `password`
}

func getConfigInfo() (*ConfigInfo, error) {
    configInfo := new(ConfigInfo)
	configBytes := []byte(os.Getenv("CONFIG_FILE"))
    if err := json.Unmarshal(configBytes, configInfo); err != nil {
        return nil, errors.New("Environment variable CONFIG_FILE is not set correctly")
    }
	if configInfo.Indexd == (IndexdInfo{}) {
		// XXX log
		fmt.Println("Defaulting to Indexd creds")
		if err := json.Unmarshal(configBytes, &configInfo.Indexd); err != nil {
			return nil, errors.New("Environment variable CONFIG_FILE is not set correctly")
		}
	}

    return configInfo, nil
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
	_, key := u.Host, u.Path

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
    // fmt.Println(key)
    // fmt.Println(split_key)
    // fmt.Println(len(split_key))
    // fmt.Println(uuid)
	filename := split_key[len(split_key)-1]
    // fmt.Println(filename)

	// client, err := CreateNewAwsClient()
	// if err != nil {
		// log.Panicf("Can not create AWS client. Detail %s\n\n", err)
	// }

	// log.Printf("Start to compute hashes for %s", key)
	// hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)
	_, objectSize := new(HashInfo), 42

	if err != nil {
		log.Panicf("Can not compute hashes for %s. Detail %s ", key, err)
	}
	log.Printf("Finish to compute hashes for %s", key)

    // XXX error handling
	configInfo, err := getConfigInfo()
	if err != nil {
		log.Panicf("%s ", err)
	}
    // XXX take out
    fmt.Println(configInfo)
    fmt.Println(err)

    rev, err := GetIndexdRecordRev(uuid, configInfo.Indexd.URL)
    // fmt.Println(rev, err)
	mdsUploadedBody := fmt.Sprintf(`{"_upload_status": "uploaded", "_filename": "%s"}`, filename)
    if err != nil {
        log.Panicf("Can not get record %s from Indexd. Error message %s", uuid, err)
    } else if rev == "" {
        log.Printf("Indexd record with guid %s already has size and hashes", uuid)
        updateMetadataObjectWrapper(uuid, configInfo, mdsUploadedBody)
        return
    }
	log.Printf("Got rev %s from Indexd for record %s", rev, uuid)

    updateMetadataObjectWrapper(uuid, configInfo, `{"_upload_status": "indexs3client job calculating hashes and size"}`)

    // XXX use real hashes
	// indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		// objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, "1", "2", "3", "4", "5")
    resp, err := UpdateIndexdRecord(uuid, rev, &configInfo.Indexd, []byte(indexdHashesBody))
    if err != nil {
        log.Panicf("Could not update Indexd record %s. Request Body: %s. Error: %s", uuid, indexdHashesBody, err)
	} else if resp.StatusCode != http.StatusOK {
        log.Panicf("Could not update Indexd record %s. Request Body: %s. Response Status Code: %d", uuid, indexdHashesBody, resp.StatusCode)
	}
    log.Printf("Updated Indexd record %s with hash info. Request Body: %s, Response Status Code: %d", uuid, indexdHashesBody, resp.StatusCode)

    updateMetadataObjectWrapper(uuid, configInfo, mdsUploadedBody)
}
