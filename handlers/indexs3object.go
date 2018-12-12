package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

	log.Println("Start get rev")
	log.Println("requested url:", indexURL+"/"+uuid)

	log.Println("response Status:", resp.Status)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Can not get record rev of %s. IndexURL %s", uuid, indexURL)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var data interface{}

	json.Unmarshal(body, &data)
	data = data.(map[string]interface{})["rev"]

	return data.(string), nil
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
