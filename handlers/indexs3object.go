package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
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
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Can not get record rev of %s. IndexURL %s", uuid, indexURL))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var data interface{}

	json.Unmarshal(body, &data)
	data = data.(map[string]interface{})["rev"]

	return data.(string), nil
}

// UpdateIndexdRecord updates the record with size, urls and hashes endcoded in body
func UpdateIndexdRecord(uuid, rev, indexURL, username, password string, body []byte) (*http.Response, error) {
	endpoint := indexURL + "/blank/" + uuid + "?rev=" + rev
	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	return client.Do(req)
}
