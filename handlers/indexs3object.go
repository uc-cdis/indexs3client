package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// GetIndexdRecordRev gets record rev
func GetIndexdRecordRev(uuid, indexURL string) (string, error) {
	req, err := http.NewRequest("GET", indexURL+"/"+uuid, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Can not get rev of the record %s. IndexURL %s. Status code %d", uuid, indexURL, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var data interface{}

	json.Unmarshal(body, &data)

	iDataMap := data.(map[string]interface{})

	data = iDataMap["rev"]

	if revStr, ok := data.(string); ok {
		return revStr, nil
	}

	return "", fmt.Errorf("Can not get rev of the record with uuid '%v'. IndexURL %s. Status code %d. Key \"rev\" not found in data map: %v", uuid, indexURL, resp.StatusCode, iDataMap)
}

// UpdateIndexdRecord updates the record with size, urls and hashes endcoded in body
func UpdateIndexdRecord(uuid, rev string, indexdInfo *IndexdInfo, body []byte) (*http.Response, error) {
	endpoint := indexdInfo.URL + "/blank/" + uuid + "?rev=" + rev
	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	client := &http.Client{}
	return client.Do(req)
}

func CreateBlankIndexdRecord(indexdInfo *IndexdInfo, body []byte) (*IndexdRecord, error) {
	endpoint := indexdInfo.URL + "/blank"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		record := new(IndexdRecord)
		if err := json.Unmarshal(bodyBytes, record); err != nil {
			return nil, err
		}
		return record, nil
	}

	return nil, fmt.Errorf("Problem creating blank record for %v", string(body))
}

type searchResponse []struct {
	DID  string   `json:"did"`
	URLs []string `json:"urls"`
}

func SearchRecordByURL(indexdInfo *IndexdInfo, url string) (searchResponse, error) {

	baseURL := strings.TrimSuffix(indexdInfo.URL, "/index")
	req, err := http.NewRequest("GET", baseURL+"/_query/urls/q", nil)

	q := req.URL.Query()
	q.Add("include", url)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed getting rev for url %s. IndexURL %s. Status code %d", url, indexdInfo.URL, resp.StatusCode)
	}
	var data searchResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	uErr := json.Unmarshal(body, &data)
	if uErr != nil {
		return data, uErr
	}
	return data, nil

}
