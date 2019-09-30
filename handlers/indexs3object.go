package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	return "", fmt.Errorf("Can not get rev of the record %s. IndexURL %s. Status code %d. Key \"rev\" not found in data map: %v", uuid, indexURL, resp.StatusCode, iDataMap)

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
