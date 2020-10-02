package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// MaxRetries maximum number of retries
const MaxRetries = 5

type ConfigInfo struct {
	Indexd          IndexdInfo          `indexd`
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

// Read Indexd and Metadata Service config info from CONFIG_FILE into
// ConfigInfo struct. Panic if Indexd creds could not be found
func getConfigInfo() *ConfigInfo {
	configInfo := new(ConfigInfo)
	configBytes := []byte(os.Getenv("CONFIG_FILE"))
	log.Printf("Attempting to unmarshal both Indexd and Metadata Service configs from JSON in CONFIG_FILE env variable")
	if err := json.Unmarshal(configBytes, configInfo); err != nil {
		log.Panicf("Could not unmarshal JSON in CONFIG_FILE env variable: %s", err)
	}
	if configInfo.Indexd == (IndexdInfo{}) {
		log.Printf("Could not find required Indexd config when unmarshalling both Indexd and Metadata Service configs. Trying again to only unmarshal Indexd config")
		if err := json.Unmarshal(configBytes, &configInfo.Indexd); err != nil {
			log.Panicf("Could not unmarshal JSON in CONFIG_FILE env variable: %s", err)
		}
		if configInfo.Indexd == (IndexdInfo{}) {
			log.Panicf("Could not find required Indexd config in JSON in CONFIG_FILE env variable")
		}
	}
	log.Printf("Indexd config was unmarshalled")
	if configInfo.MetadataService != (MetadataServiceInfo{}) {
		log.Printf("Metadata Service config was unmarshalled")
	} else {
		log.Printf("Metadata Service config was not unmarshalled")
	}

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
	filename := split_key[len(split_key)-1]

	log.Printf("Attempting to get rev for record %s in Indexd", uuid)
	rev, err := GetIndexdRecordRev(uuid, configInfo.Indexd.URL)
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

	indexdHashesBody := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)
	log.Printf("Attempting to update Indexd record %s. Request Body: %s", uuid, indexdHashesBody)
	resp, err := UpdateIndexdRecord(uuid, rev, &configInfo.Indexd, []byte(indexdHashesBody))
	if err != nil {
		log.Panicf("Could not update Indexd record %s. Error: %s", uuid, err)
	} else if resp.StatusCode != http.StatusOK {
		log.Panicf("Could not update Indexd record %s. Response Status Code: %d", uuid, resp.StatusCode)
	}
	log.Printf("Updated Indexd record %s with hash info. Response Status Code: %d", uuid, resp.StatusCode)

	updateMetadataObjectWrapper(uuid, configInfo, mdsUploadedBody)
}
