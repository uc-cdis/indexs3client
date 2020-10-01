package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"github.com/hashicorp/go-retryablehttp"
)

// GetIndexdRecordRev gets record rev
func GetIndexdRecordRev(uuid, indexURL string) (string, error) {
	// req, err := http.NewRequest("GET", indexURL+"/"+uuid, nil)
    req, err := retryablehttp.NewRequest("GET", indexURL+"/"+uuid, nil)
	// client := &http.Client{}
	client := retryablehttp.NewClient()
	// client.ResponseLogHook = func(logger retryablehttp.Logger, resp *http.Response) {
		// if resp.StatusCode == 200 {
			// successLog := "test_log_pass"
			// // Log something when we get a 200
            // logger.Printf(successLog)
		// } else {
			// // Log the response body when we get a 500
			// body, _ := ioutil.ReadAll(resp.Body)
			// // if err != nil {
				// // // t.Fatalf("err: %v", err)
				// // logger.Panicf("Wrong url format %s\n", s3objectURL)
			// // }
			// failLog := string(body)
			// logger.Printf(failLog)
		// }
	// }
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		// logger.Printf("retryNumber: %d", retryNumber)
		log.Printf("retryNumber: %d", retryNumber)
		// retryCount = retryNumber

		// if logger != client.Logger {
			// t.Fatalf("Client logger was not passed to logging hook")
		// }

		dumpBytes, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			log.Fatal("Dumping requests failed")
		}

		dumpString := string(dumpBytes)
		// logger.Printf("request:\n%s", dumpString)
		log.Printf("request:\n%s", dumpString)
		// if !strings.Contains(dumpString, "PUT /v1/foo") {
			// t.Fatalf("Bad request dump:\n%s", dumpString)
		// }
	}
	resp, err := client.Do(req)
	fmt.Println(client.RetryMax)
	fmt.Println(client.RetryWaitMin)
	fmt.Println(client.RetryWaitMax)
	if err != nil {
		return "", err
	}
	// XXX why?
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Can not get rev of the record %s. IndexURL %s. Status code: %d", uuid, indexURL, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var data interface{}

	json.Unmarshal(body, &data)
	rev := data.(map[string]interface{})["rev"]
	size := data.(map[string]interface{})["size"]
	// return rev.(string), nil

	if size == nil {
		return rev.(string), nil
	}
	return "", nil
}

// UpdateIndexdRecord updates the record with size, urls and hashes endcoded in body
func UpdateIndexdRecord(uuid, rev string, indexdInfo *IndexdInfo, body []byte) (*http.Response, error) {
	endpoint := indexdInfo.URL + "/blank/" + uuid + "?rev=" + rev
	// req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	req, err := retryablehttp.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(indexdInfo.Username, indexdInfo.Password)

	// client := &http.Client{}
	client := retryablehttp.NewClient()
	return client.Do(req)
}
