package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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

// parseGuidFromKey enforces that key is either:
//   - <uuid>/<path to file...>
//   - <prefix>/<uuid>/<path to file...>
//
// It returns (guid, filePath, matched).
func parseGuidFromKey(key string) (string, string, bool) {
	key = strings.Trim(key, "/")
	parts := strings.Split(key, "/")

	// Must be at least "<uuid>/<file>" or "<prefix>/<uuid>/<file>"
	if len(parts) < 2 {
		return "", "", false
	}

	// Case B: "<prefix>/<uuid>/<path...>"
	if len(parts) >= 3 {
		if u, err := id.Parse(parts[1]); err == nil && u != id.Nil {
			return parts[0] + "/" + parts[1], strings.Join(parts[2:], "/"), true
		}
	}

	// Case A: "<uuid>/<path...>"
	if u, err := id.Parse(parts[0]); err == nil && u != id.Nil {
		return parts[0], strings.Join(parts[1:], "/"), true
	}

	return "", "", false
}

// IndexS3Object indexes s3 object. The function does several things:
// it downloads the object from S3, computes size and hashes, and updates Indexd
// and potentially Metadata Service.
func IndexS3Object(s3objectURL string) {
	configInfo := getConfigInfo()

	s3objectURL, _ = url.QueryUnescape(s3objectURL)
	u, err := url.Parse(s3objectURL)
	if err != nil {
		log.Panicf("Wrong url format %s\n", s3objectURL)
	}
	bucket, key := u.Host, u.Path

	key = strings.Trim(key, "/")

	// Enforce the required key formats for derived GUIDs
	derivedDID, derivedFilePath, matched := parseGuidFromKey(key)

	var did string
	var rev string
	var notFound bool

	// If key matches "<uuid>/<path>" or "<prefix>/<uuid>/<path>":
	// check for indexd record matching that DID.
	if matched {
		did = derivedDID

		log.Printf("Attempting to get rev for record %s in Indexd", did)
		rev, err = GetIndexdRecordRev(did, configInfo.Indexd.URL)

		var mdsUploadedBody string = `{"_upload_status": "uploaded"}`
		if err != nil {
			if err == ErrIndexdNotFound {
				notFound = true
			} else {
				log.Panicf("Can not get record %s from Indexd. Error message %s", did, err)
			}
		} else if rev == "" {
			// Existing behavior: if size != nil in Indexd, GetIndexdRecordRev returns "" and we exit.
			log.Printf("Indexd record with guid %s already has size and hashes", did)
			updateMetadataObjectWrapper(did, configInfo, mdsUploadedBody)
			return
		} else {
			log.Printf("Got rev %s from Indexd for record %s", rev, did)
		}
	} else {
		// Key doesn't match required derived-guid formats:
		// create a blank record and use its DID.
		log.Printf("Key does not match <uuid>/<path> or <prefix>/<uuid>/<path>. Creating blank Indexd record.")
		out, err := CreateBlankIndexdEntry(&configInfo.Indexd)
		if err != nil {
			log.Panicf("Could not create blank Indexd entry. Error: %s", err)
		}
		did = out.DID
		log.Printf("Created blank Indexd record with guid %s", did)

		log.Printf("Attempting to get rev for newly created blank record %s in Indexd", did)
		rev, err = GetIndexdRecordRev(did, configInfo.Indexd.URL)
		if err != nil {
			log.Panicf("Can not get record %s from Indexd after blank create. Error message %s", did, err)
		} else if rev == "" {
			// Unlikely for a newly created blank entry, but keep behavior consistent.
			log.Printf("Indexd record with guid %s already has size and hashes (unexpected for blank create)", did)
			updateMetadataObjectWrapper(did, configInfo, `{"_upload_status": "uploaded"}`)
			return
		}
		log.Printf("Got rev %s from Indexd for record %s", rev, did)
	}

	updateMetadataObjectWrapper(did, configInfo, `{"_upload_status": "processing"}`)

	var mdsUploadedBody string = `{"_upload_status": "uploaded"}`
	var mdsErrorBody string = `{"_upload_status": "error"}`

	client, err := CreateNewAwsClient()
	if err != nil {
		updateMetadataObjectWrapper(did, configInfo, mdsErrorBody)
		log.Panicf("Can not create AWS client. Detail %s\n\n", err)
	}

	log.Printf("Start to compute hashes for %s", key)
	hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)
	if err != nil {
		updateMetadataObjectWrapper(did, configInfo, mdsErrorBody)
		log.Panicf("Can not compute hashes for %s. Detail %s ", key, err)
	}
	log.Printf("Finish to compute hashes for %s", key)

	// Payload for blank update (PUT /index/blank/{did}?rev=...)
	indexdHashesBody := fmt.Sprintf(
		`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c,
	)

	// If we matched a derived DID but the record didn't exist, create it using addEntry.
	if matched && notFound {
		log.Printf("Indexd record %s does not exist; creating via addEntry with derived guid", did)

		addReq := &IndexdAddEntryRequest{
			DID:  did,
			Form: "object",
			Size: objectSize,
			URLs: []string{s3objectURL},
			Hashes: map[string]string{
				"md5":    hashes.Md5,
				"sha1":   hashes.Sha1,
				"sha256": hashes.Sha256,
				"sha512": hashes.Sha512,
				"crc":    hashes.Crc32c,
			},
			FileName: derivedFilePath,
		}

		_, err := AddIndexdEntry(&configInfo.Indexd, addReq)
		if err != nil {
			updateMetadataObjectWrapper(did, configInfo, mdsErrorBody)
			log.Panicf("Could not create Indexd record %s via addEntry. Error: %s", did, err)
		}

		log.Printf("Created Indexd record %s via addEntry", did)
		updateMetadataObjectWrapper(did, configInfo, mdsUploadedBody)
		return
	}

	// If we matched derived guid and we had a blank record, rev was set earlier.
	// If we created blank guid (non-matching key), rev was also set earlier.
	if rev == "" {
		// If we get here, it means we don't have a rev but also didn't take the addEntry path.
		// That generally means the record already had size/hashes (handled earlier) OR something unexpected.
		// Be conservative and stop to avoid writing to the wrong endpoint.
		log.Printf("No rev available for guid %s; skipping blank update.", did)
		updateMetadataObjectWrapper(did, configInfo, mdsUploadedBody)
		return
	}

	log.Printf("Attempting to update Indexd record %s. Request Body: %s", did, indexdHashesBody)
	resp, err := UpdateIndexdRecord(did, rev, &configInfo.Indexd, []byte(indexdHashesBody))
	if err != nil {
		updateMetadataObjectWrapper(did, configInfo, mdsErrorBody)
		log.Panicf("Could not update Indexd record %s. Error: %s", did, err)
	} else if resp.StatusCode != http.StatusOK {
		updateMetadataObjectWrapper(did, configInfo, mdsErrorBody)
		log.Panicf("Could not update Indexd record %s. Response Status Code: %d", did, resp.StatusCode)
	}
	log.Printf("Updated Indexd record %s with hash info. Response Status Code: %d", did, resp.StatusCode)

	updateMetadataObjectWrapper(did, configInfo, mdsUploadedBody)
}
