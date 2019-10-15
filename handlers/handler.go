package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type IndexdInfo struct {
	URL              string `json:"url"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	ExtramuralBucket bool   `json:"extramural_bucket"`

	ExtramuralUploader         *string `json:"extramural_uploader"`
	ExtramuralUploaderS3Owner  bool    `json:"extramural_uploader_s3owner"`
	ExtramuralUploaderManifest *string `json:"extramural_uploader_manifest"`
}

type IndexdRecord struct {
	DID    string `json:"did"`
	BaseID string `json:"baseid"`
	Rev    string `json:"rev"`
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
	if err := json.Unmarshal([]byte(os.Getenv("CONFIG_FILE")), indexdInfo); err != nil {
		return nil, errors.New("Enviroiment variable CONFIG_FILE is not set correctly")
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
		log.Printf("Wrong url format %s\n", s3objectURL)
		return
	}
	bucket, key := u.Host, u.Path

	indexdInfo, _ := getIndexServiceInfo()

	client, err := CreateNewAwsClient()
	if err != nil {
		log.Printf("Can not create AWS client. Detail %s\n\n", err)
		return
	}

	log.Printf("Start to compute hashes for %s", key)
	hashes, objectSize, err := CalculateBasicHashes(client, bucket, key)

	if err != nil {
		log.Printf("Can not compute hashes for %s. Detail %s ", key, err)
		return
	}
	log.Printf("Finish to compute hashes for %s", key)

	var uuid, rev string

	// Create the indexd record if this is an ExtramuralBucket and it doesn't already exist
	if indexdInfo.ExtramuralBucket {

		// search indexd to see if the record already exists
		if foundIndexdRecord, err := GetIndexdRecordByURL(indexdInfo, s3objectURL); err == nil {
			uuid = foundIndexdRecord.DID
			rev = foundIndexdRecord.Rev
		} else {
			var uploader string

			if indexdInfo.ExtramuralUploader != nil {
				uploader = *(indexdInfo.ExtramuralUploader)
			} else if indexdInfo.ExtramuralUploaderS3Owner {
				s3owner, err := GetS3BucketOwner(client, bucket)
				if err != nil {
					panic(err) // Should always be able to fetch owner, something bad happened if not
				}

				uploader = s3owner
			} else if indexdInfo.ExtramuralUploaderManifest != nil {
				// Read from manifest, try to find uploader.
				// if fail, default to empty string
				oo, err := GetS3ObjectOutput(client, bucket, *indexdInfo.ExtramuralUploaderManifest)
				if err == nil {
					uploader = FindUploaderInManifest(key, oo.Body)
				} else {
					log.Println(err)
				}
			}

			body, _ := json.Marshal(struct {
				Uploader string `json:"uploader"`
				Filename string `json:"file_name"`
			}{
				uploader, filepath.Base(key),
			})

			indexdRecord, err := CreateBlankIndexdRecord(indexdInfo, body)
			if err != nil {
				log.Println(err)
				return
			}

			uuid = indexdRecord.DID
			rev = indexdRecord.Rev
		}

	} else {
		// key looks like one of these:
		//
		//     <uuid>/<filename>
		//     <dataguid>/<uuid>/<filename>
		//
		// we want to keep the `<dataguid>/<uuid>` part
		split_key := strings.Split(key, "/")
		if len(split_key) == 2 {
			uuid = split_key[0]
		} else {
			uuid = strings.Join(split_key[:len(split_key)-1], "/")
		}

		if uuid == "" {
			panic("Are you trying to index a non-Gen3 managed S3 bucket? Try setting 'extramural_bucket: true' in the config, no UUID found in object path.")
		}

		rev, err = GetIndexdRecordRev(uuid, indexdInfo.URL)
		if err != nil {
			log.Println(err)
			return
		}
	}

	body, _ := json.Marshal(struct {
		Size   int64     `json:"size"`
		URLs   []string  `json:"urls"`
		Hashes *HashInfo `json:"hashes"`
	}{
		objectSize, []string{s3objectURL}, hashes,
	})

	resp, err := UpdateIndexdRecord(uuid, rev, indexdInfo, body)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Finish updating the record. Response Status: %s", resp.Status)

}

func FindUploaderInManifest(object string, oo io.Reader) string {
	uploader := ""

	manifest := csv.NewReader(oo)
	manifestRecords, err := manifest.ReadAll()
	if err == nil {
		uploaderMap := make(map[string]string)

		for _, row := range manifestRecords {
			uploaderMap[row[0]] = row[1]
		}

		if val, ok := uploaderMap[object]; ok {
			uploader = val
		} else {
			log.Printf("Object %s not found in uploader manifest file", object)
		}
	} else {
		log.Println(err)
	}

	return uploader
}
