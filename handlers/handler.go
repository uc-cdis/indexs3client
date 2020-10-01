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
	// "time"
	// "github.com/hashicorp/go-retryablehttp"
)

// MaxRetries maximum number of retries
const MaxRetries = 5

type ConfigInfo struct {
    // Indexd interface{} `indexd`
    // MetadataService interface{} `metadata_service`
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
    fmt.Println(configInfo)
    fmt.Println(err)

	// var retries = 0
	// var rev = ""

	// for {
		// rev, err = GetIndexdRecordRev(uuid, configInfo.Indexd.URL)
		// if err != nil {
			// retries++
			// log.Printf("Error: %s. Retry: %d", err, retries)
			// time.Sleep(5 * time.Second)
		// } else if rev == "" {
			// log.Println("The file already has size and hashes")
			// return
		// } else {
			// break
		// }
		// if retries == MaxRetries {
			// log.Panicf("Can not get record %s from indexd. Error message %s", uuid, err)
		// }
	// }
    rev, err := GetIndexdRecordRev(uuid, configInfo.Indexd.URL)
    fmt.Println(rev, err)
	mdsUploadedBody := fmt.Sprintf(`{"_upload_status": "uploaded", "_filename": "%s"}`, filename)
    if err != nil {
        log.Panicf("Can not get record %s from indexd. Error message %s", uuid, err)
    } else if rev == "" {
        log.Println("The file already has size and hashes")
        // XXX error handling
        UpdateMetadataObject(uuid, &configInfo.MetadataService, []byte(mdsUploadedBody))
        return
    }
	log.Printf("Got rev %s from Indexd for record %s", rev, uuid)

	resp, err := UpdateMetadataObject(uuid, &configInfo.MetadataService, []byte(`{"_upload_status": "calculating hashes and size"}`))
    if err != nil {
		log.Printf("Error: %s", err)
    } else if resp.StatusCode != http.StatusOK {
        log.Printf("Could not update metadata object _upload_status to calculating for metadata object %s. resp.StatusCode: %s", uuid, resp.StatusCode)
    } else {
        log.Printf("Updated _upload_status field to calculating for metadata object %s", uuid)
    }

	// indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		// objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, "1", "2", "3", "4", "5")
    // fmt.Println(indexdHashesBody)
	resp, err = UpdateIndexdRecord(uuid, rev, &configInfo.Indexd, []byte(indexdHashesBody))
    if err != nil {
		log.Panicf("Could not update Indexd record %s with hash info. Hash info %s. Detail %s", uuid, indexdHashesBody, err)
	} else if resp.StatusCode != http.StatusOK {
		log.Panicf("Could not update Indexd record %s with hash info. Hash info: %s. Status code: %d", uuid, indexdHashesBody, resp.StatusCode)
	}
	log.Printf("Finished updating Indexd record %s with hash info. Response status: %d. Hash info: %s", uuid, resp.StatusCode, indexdHashesBody)

    resp, err = UpdateMetadataObject(uuid, &configInfo.MetadataService, []byte(mdsUploadedBody))
    if err != nil {
		log.Printf("Error: %s", err)
    } else if resp.StatusCode != http.StatusOK {
        log.Printf("Could not update metadata object _upload_status to updated for metadata object %s. resp.StatusCode: %s", uuid, resp.StatusCode)
    } else {
        log.Printf("Updated _upload_status field to uploaded for metadata object %s", uuid)
    }

	// retries = 0
	// for {
		// resp, err := UpdateIndexdRecord(uuid, rev, &configInfo.Indexd, []byte(body))
		// if err != nil {
			// retries++
			// log.Printf("Error: %s. Retry: %d", err, retries)
			// time.Sleep(5 * time.Second)
		// } else if resp.StatusCode != 200 {
			// log.Printf("StatusCode: %d. Retry: %d", resp.StatusCode, retries)
			// retries++
			// time.Sleep(5 * time.Second)
		// } else {
			// log.Printf("Finish updating the record %s. Response Status: %d. Body %s", uuid, resp.StatusCode, body)
			// break
		// }

		// if retries == MaxRetries {
			// if err == nil {
				// log.Panicf("Can not update %s with hash info. Body %s. Status code %d. Detail %s", uuid, body, resp.StatusCode, err)
			// } else {
				// log.Panicf("Can not update %s with hash info. Body %s. Detail %s", uuid, body, err)
			// }
			// break
		// }
	// }
}

// func updateMetadataObjectWrapper(uuid string, mdsInfo *MetadataServiceInfo, body string) {
    // resp, err = UpdateMetadataObject(uuid, &configInfo.MetadataService, []byte(body))
    // if err != nil {
		// log.Printf("Error: %s", err)
    // } else if resp.StatusCode != http.StatusOK {
        // log.Printf("Could not update metadata object _upload_status to updated for metadata object %s. resp.StatusCode: %s", uuid, resp.StatusCode)
    // } else {
        // log.Printf("Updated _upload_status field to uploaded for metadata object %s", uuid)
    // }
// }
