package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// Checks for presence of MDS config, calls UpdateMetadataObject, and logs info
func updateMetadataObjectWrapper(uuid string, configInfo *ConfigInfo, body string) {
	if configInfo.MetadataService != (MetadataServiceInfo{}) {
		log.Printf("Attempting to update object with guid %s in Metadata Service. Request Body: %s", uuid, body)
		resp, err := UpdateMetadataObject(uuid, &configInfo.MetadataService, []byte(body))
		if err != nil {
			log.Printf("Could not update object with guid %s in Metadata Service. Error: %s", uuid, err)
		} else if resp.StatusCode != http.StatusOK {
			log.Printf(
				`Could not update object with guid %s in Metadata Service. Response Status Code: %d.
				If using gen3-client, a 404 here does not necessarily indicate a problem, as we would
				only expect there to be a corresponding object in the Metadata Service if --metadata
				was supplied to the gen3-client upload command.`,
				uuid,
				resp.StatusCode)
		} else {
			log.Printf("Updated object with guid %s in Metadata Service. Response Status Code: %d", uuid, resp.StatusCode)
		}
	} else {
		log.Printf("Not attempting to update object with guid %s in Metadata Service because Metadata Service creds were not configured", uuid)
	}
}

// Updates Metadata Service object specified by uuid with provided body
func UpdateMetadataObject(uuid string, mdsInfo *MetadataServiceInfo, body []byte) (*http.Response, error) {
	endpoint := mdsInfo.URL + "/metadata/" + uuid + "?merge=True"
	req, err := retryablehttp.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(mdsInfo.Username, mdsInfo.Password)

	client := retryablehttp.NewClient()
	client.RetryMax = MaxRetries
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, err
}
