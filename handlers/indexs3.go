package handlers

import (
	"bytes"
	"net/http"
)

func MakeARequest(client *http.Client, method string, apiEndpoint string, headers map[string]string, body *bytes.Buffer) (*http.Response, error) {
	/*
		Make http request with header and body
	*/
	req, err := http.NewRequest(method, apiEndpoint, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil

}
