package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// GetIndexdRecordRev gets record rev
func GetIndexdRecordRev(uuid, indexURL string) (string, error) {
	req, err := retryablehttp.NewRequest("GET", indexURL+"/"+uuid, nil)
	if err != nil {
		log.Printf("Error %s", err)
	}
	client := retryablehttp.NewClient()
	client.RetryMax = MaxRetries
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Can not get rev of the record %s. IndexURL %s. Status code: %d", uuid, indexURL, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var data interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Unmarshall error %s", err)
	}
	rev := data.(map[string]interface{})["rev"]
	size := data.(map[string]interface{})["size"]

	if size == nil {
		return rev.(string), nil
	}
	return "", nil
}

// UpdateIndexdRecord updates the record with size, urls and hashes encoded in body
func UpdateIndexdRecord(uuid, rev string, indexdInfo *IndexdInfo, body []byte) (*http.Response, error) {
	endpoint := indexdInfo.URL + "/blank/" + uuid + "?rev=" + rev
	req, err := retryablehttp.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	client := retryablehttp.NewClient()
	client.RetryMax = MaxRetries
	resp, err := client.Do(req)
	if resp.StatusCode == 403 {
		log.Printf("Possible auth issue for Indexd user '%s' (basic auth)", indexdInfo.Username)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, err
}
