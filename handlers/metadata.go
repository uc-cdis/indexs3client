package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
    "fmt"
	"net/http"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type MetadataServiceInfo struct {
	URL      string `url`
	Username string `username`
	Password string `password`
}

func UpdateMetadataObject(s3objectURL string) {
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

    // Diverges from here down
    // ================
	// client, err := CreateNewAwsClient()
	// if err != nil {
		// log.Panicf("Can not create AWS client. Detail %s\n\n", err)
	// }

	// log.Printf("Start to compute hashes for %s", key)
	// hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)

	// if err != nil {
		// log.Panicf("Can not compute hashes for %s. Detail %s ", key, err)
	// }
	// log.Printf("Finish to compute hashes for %s", key)

	// indexdInfo, _ := getIndexServiceInfo()
    mdsInfo, _ := getMetadataServiceInfo()
    fmt.Println(mdsInfo)
    // fmt.Println(mdsInfo.URL)
    // fmt.Println(mdsInfo.Username)
    // fmt.Println(mdsInfo.Password)

	// var retries = 0
	// var rev = ""

	// for {
		// rev, err = GetIndexdRecordRev(uuid, indexdInfo.URL)
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

	// body := fmt.Sprintf(`{"size": %d, "urls": ["%s"], "hashes": {"md5": "%s", "sha1":"%s", "sha256": "%s", "sha512": "%s", "crc": "%s"}}`,
		// objectSize, s3objectURL, hashes.Md5, hashes.Sha1, hashes.Sha256, hashes.Sha512, hashes.Crc32c)

    body := `{"_upload_status": "uploaded"}`
    var retries = 0
	for {
		resp, err := MakePutRequest(uuid, mdsInfo, []byte(body))
		if err != nil {
			retries++
			log.Printf("Error: %s. Retry: %d", err, retries)
			time.Sleep(5 * time.Second)
		} else if resp.StatusCode != 200 {
			log.Printf("StatusCode: %d. Retry: %d", resp.StatusCode, retries)
			retries++
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Finish updating the object %s. Response Status: %d. Body %s", uuid, resp.StatusCode, body)
			break
		}

		if retries == MaxRetries {
			if err == nil {
				log.Panicf("Can not update metadata object for %s. Body %s. Status code %d. Detail %s", uuid, body, resp.StatusCode, err)
			} else {
				log.Panicf("Can not update metadata object for %s. Body %s. Detail %s", uuid, body, err)
			}
			break
		}
	}
}

func getMetadataServiceInfo() (*MetadataServiceInfo, error) {
	configInfo := new(ConfigInfo)
	if err := json.Unmarshal([]byte(os.Getenv("CONFIG_FILE")), configInfo); err != nil {
		return nil, errors.New("MDS - Environment variable CONFIG_FILE is not set correctly")
	}

    mds_bytes, _ := json.Marshal(configInfo.MetadataService)

	mdsInfo := new(MetadataServiceInfo)
    json.Unmarshal(mds_bytes, &mdsInfo)

	return mdsInfo, nil
}

func MakePutRequest(uuid string, mdsInfo *MetadataServiceInfo, body []byte) (*http.Response, error) {
	endpoint := mdsInfo.URL + "/metadata/" + uuid + "?merge=True"
	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
    req.SetBasicAuth(mdsInfo.Username, mdsInfo.Password)

	client := &http.Client{}
	return client.Do(req)
}
