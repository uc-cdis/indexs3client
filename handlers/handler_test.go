package handlers

import (
	"encoding/json"
	"testing"

	"github.com/magiconair/properties/assert"
)

func makeImageConfigString() string {
	return `{
				"url": "http://indexd-service/",
				"username": "test",
				"password": "test"
		  	}`
}
func TestHandler(t *testing.T) {
	indexdInfo := new(IndexdInfo)
	if err := json.Unmarshal([]byte(makeImageConfigString()), indexdInfo); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, indexdInfo.URL, "http://indexd-service/")
	assert.Equal(t, indexdInfo.Username, "test")
	assert.Equal(t, indexdInfo.Password, "test")

}
