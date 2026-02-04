package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

var ErrIndexdNotFound = errors.New("indexd record not found")

// GetIndexdRecordRev gets record rev.
// Returns:
//   - (rev, nil) when record exists and size is nil (blank record)
//   - ("", nil) when record exists and size is set (already indexed)
//   - ("", ErrIndexdNotFound) when record does not exist (404)
//   - ("", error) for other failures
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

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrIndexdNotFound
	}
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

// IndexdAddEntryRequest models the request for POST /index (addEntry).
// Minimal required fields per swagger: size, urls, hashes, form.
// did is optional but we use it when key has a derived GUID.
type IndexdAddEntryRequest struct {
	DID      string            `json:"did,omitempty"`
	FileName string            `json:"file_name,omitempty"`
	Form     string            `json:"form"`
	Size     int64             `json:"size"`
	URLs     []string          `json:"urls"`
	Hashes   map[string]string `json:"hashes"`
}

type indexdOutputRef struct {
	DID string `json:"did"`
	Rev string `json:"rev"`
}

// AddIndexdEntry creates a new Indexd record via POST /index.
func AddIndexdEntry(indexdInfo *IndexdInfo, payload *IndexdAddEntryRequest) (*indexdOutputRef, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// indexdInfo.URL is expected to be the "/index" base (same as existing code uses for GET/PUT blank).
	endpoint := indexdInfo.URL
	req, err := retryablehttp.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	client := retryablehttp.NewClient()
	client.RetryMax = MaxRetries

	resp, err := client.Do(req)
	if resp != nil && resp.StatusCode == 403 {
		log.Printf("Possible auth issue for Indexd user '%s' (basic auth)", indexdInfo.Username)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Could not add Indexd entry. Status: %d Body: %s", resp.StatusCode, string(respBody))
	}

	var out indexdOutputRef
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateBlankIndexdEntry creates a blank Indexd record via POST /index/blank,
// returning the new GUID.
func CreateBlankIndexdEntry(indexdInfo *IndexdInfo) (*indexdOutputRef, error) {
	endpoint := indexdInfo.URL + "/blank"
	req, err := retryablehttp.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	client := retryablehttp.NewClient()
	client.RetryMax = MaxRetries

	resp, err := client.Do(req)
	if resp != nil && resp.StatusCode == 403 {
		log.Printf("Possible auth issue for Indexd user '%s' (basic auth)", indexdInfo.Username)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Could not create blank Indexd entry. Status: %d Body: %s", resp.StatusCode, string(respBody))
	}

	var out indexdOutputRef
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	if out.DID == "" {
		return nil, fmt.Errorf("Blank create did not return a did. Body: %s", string(respBody))
	}
	return &out, nil
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
	if resp != nil && resp.StatusCode == 403 {
		log.Printf("Possible auth issue for Indexd user '%s' (basic auth)", indexdInfo.Username)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, err
}
