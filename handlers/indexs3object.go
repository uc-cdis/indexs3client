package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func GetIndexdRecordByHash(indexdInfo *IndexdInfo, hash *HashInfo) (*IndexdRecord, error) {

	// http://indexd-service/index
	u, err := url.Parse(indexdInfo.URL)
	if err != nil {
		return nil, err
	}
	u.Path = "/urls"

	req, err := http.NewRequest("GET", u.String(), nil)

	q := req.URL.Query()
	q.Add("hash", "md5:"+hash.Md5)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed getting rev for hash %s. IndexURL %s. Status code %d", hash.Md5, indexdInfo.URL, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var data interface{}

	json.Unmarshal(body, &data)

	iDataMap := data.(map[string]interface{})

	urls, ok := iDataMap["urls"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Can not get urls for hash %s. IndexURL %s", hash.Md5, indexdInfo.URL)
	}
	metadata, ok2 := urls[0]["metadata"].(map[string]interface{})
	if !ok2 {
		return nil, fmt.Errorf("Can not get metadata for hash %s. IndexURL %s", hash.Md5, indexdInfo.URL)
	}

	record := new(IndexdRecord)
	record.DID = metadata["did"].(string)
	record.BaseID = metadata["baseid"].(string)
	record.Rev = metadata["rev"].(string)

	return record, nil
}
