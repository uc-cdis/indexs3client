package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	id "github.com/google/uuid"
)

// MaxRetries maximum number of retries
const MaxRetries = 5

type ConfigInfo struct {
	Indexd          IndexdInfo          `json:"indexd"`
	MetadataService MetadataServiceInfo `json:"metadataService"`
}

type IndexdInfo struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MetadataServiceInfo struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Read Indexd and Metadata Service config info from CONFIG_FILE into
// ConfigInfo struct. Panic if both Indexd config and Metadata Service
// configs can't be unmarshalled
func getConfigInfo() *ConfigInfo {
	configInfo := new(ConfigInfo)
	configBytes := []byte(os.Getenv("CONFIG_FILE"))

	log.Printf("Attempting to unmarshal Indexd config from JSON in CONFIG_FILE env variable")
	if err := json.Unmarshal(configBytes, &configInfo.Indexd); err != nil {
		log.Panicf("Could not unmarshal JSON in CONFIG_FILE env variable: %s", err)
	}
	if configInfo.Indexd == (IndexdInfo{}) {
		log.Panicf("Could not find Indexd config in JSON in CONFIG_FILE env variable. Both Indexd and Metadata Service configs are required")
	}

	log.Printf("Attempting to unmarshal Metadata Service config from JSON in CONFIG_FILE env variable")
	if err := json.Unmarshal(configBytes, configInfo); err != nil {
		log.Panicf("Could not unmarshal JSON in CONFIG_FILE env variable: %s", err)
	}
	if configInfo.MetadataService == (MetadataServiceInfo{}) {
		log.Panicf("Could not find Metadata Service config in JSON in CONFIG_FILE env variable. Both Indexd and Metadata Service configs are required")
	}
	log.Printf("Both Indexd and Metadata Service configs were unmarshalled")

	return configInfo
}

// IndexS3Object indexes s3 object The fuction does several things. It
// downloads the object from S3, computes size and hashes, and updates Indexd
// and potentially Metadata Service
func IndexS3Object(s3objectURL string) {
	configInfo := getConfigInfo()

	s3objectURL, _ = url.QueryUnescape(s3objectURL)
	u, err := url.Parse(s3objectURL)
	if err != nil {
		log.Panicf("Wrong url format %s\n", s3objectURL)
	}
	scheme, bucket, key := u.Scheme, u.Host, u.Path
	bucketURL := fmt.Sprintf(`%s://%s`, scheme, bucket)

	// key looks like one of these:
	//
	//     <uuid>/<filename>
	//     <dataguid>/<uuid>/<filename>
	//
	// we want to keep the `<dataguid>/<uuid>` part
	key = strings.Trim(key, "/")
	var uuid, filename, errUUID = resolveUUID(key)
	if errUUID != nil {
		log.Panicf(errUUID.Error())
	}

	fileExtension := filepath.Ext(filename)
	if len(fileExtension) > 0 {
		fileExtension = fileExtension[1:]
	}

	log.Printf("Attempting to get rev for record %s in Indexd", uuid)
	rev, err := GetIndexdRecordRev(uuid, configInfo.Indexd.URL)
	mdsUploadedBody := fmt.Sprintf(`{"_bucket": "%s", "_filename": "%s", "_file_extension": "%s", "_upload_status": "uploaded"}`, bucketURL, filename, fileExtension)
	if err != nil {
		log.Panicf("Can not get record %s from Indexd. Error message %s", uuid, err)
	} else if rev == "" {
		log.Printf("Indexd record with guid %s already has size and hashes", uuid)
		updateMetadataObjectWrapper(uuid, configInfo, mdsUploadedBody)
		return
	}
	log.Printf("Got rev %s from Indexd for record %s", rev, uuid)

	updateMetadataObjectWrapper(uuid, configInfo, `{"_upload_status": "processing"}`)

	var mdsErrorBody string = `{"_upload_status": "error"}`
	client, err := CreateNewAwsClient()
	if err != nil {
		updateMetadataObjectWrapper(uuid, configInfo, mdsErrorBody)
		log.Panicf("Can not create AWS client. Detail %s\n\n", err)
	}

	log.Printf("Start to compute hashes for %s", key)
	hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)
	if err != nil {
		updateMetadataObjectWrapper(uuid, configInfo, mdsErrorBody)
		log.Panicf("Can not compute hashes for %s. Detail %s ", key, err)
	}
	log.Printf("Finish to compute hashes for %s", key)

	indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	log.Printf("Attempting to update Indexd record %s. Request Body: %s", uuid, indexdHashesBody)
	resp, err := UpdateIndexdRecord(uuid, rev, &configInfo.Indexd, []byte(indexdHashesBody))
	if err != nil {
		updateMetadataObjectWrapper(uuid, configInfo, mdsErrorBody)
		log.Panicf("Could not update Indexd record %s. Error: %s", uuid, err)
	} else if resp.StatusCode != http.StatusOK {
		updateMetadataObjectWrapper(uuid, configInfo, mdsErrorBody)
		log.Panicf("Could not update Indexd record %s. Response Status Code: %d", uuid, resp.StatusCode)
	}
	log.Printf("Updated Indexd record %s with hash info. Response Status Code: %d", uuid, resp.StatusCode)

	updateMetadataObjectWrapper(uuid, configInfo, mdsUploadedBody)
}

func resolveUUID(key string) (string, string, error) {
	keyParts := strings.Split(key, "/")
	uuidIndex := -1
	var foundUUID id.UUID
	var err error
	var fullUUID string
	var filename string
	for i, part := range keyParts {
		foundUUID, err = id.Parse(part)
		if err == nil && foundUUID != id.Nil {
			uuidIndex = i
			break
		}
	}
	if uuidIndex == -1 {
		return "", "", fmt.Errorf("Cannot process the UUID")
	}
	fullUUID = strings.Join(keyParts[:uuidIndex+1], "/")
	filename = strings.Join(keyParts[uuidIndex+1:], "/")
	return fullUUID, filename, nil
}
