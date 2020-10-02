package handlers

import (
	"encoding/json"
	"os"
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

func TestGetConfigInfoUsingOnlyIndexdCreds(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"url": "http://indexd-service/index",
		"username": "mr happy cat",
		"password": "whiskers"
	}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	configInfo := getConfigInfo()
	assert.Equal(t, configInfo.Indexd.URL, "http://indexd-service/index")
	assert.Equal(t, configInfo.Indexd.Username, "mr happy cat")
	assert.Equal(t, configInfo.Indexd.Password, "whiskers")
	assert.Equal(t, configInfo.MetadataService, MetadataServiceInfo{})
}

func TestGetConfigInfoUsingOnlyNestedIndexdCreds(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"indexd": {
			"url": "http://indexd-service/index",
			"username": "mr happy cat",
			"password": "whiskers"
		}
	}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	configInfo := getConfigInfo()
	assert.Equal(t, configInfo.Indexd.URL, "http://indexd-service/index")
	assert.Equal(t, configInfo.Indexd.Username, "mr happy cat")
	assert.Equal(t, configInfo.Indexd.Password, "whiskers")

	assert.Equal(t, configInfo.MetadataService, MetadataServiceInfo{})
}

func TestGetConfigInfoUsingIndexdAndMDSCreds(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"indexd": {
			"url": "http://indexd-service/index",
			"username": "mr happy cat",
			"password": "whiskers"
		},
		"metadataService": {
			"url": "http://revproxy-service/mds",
			"username": "mr friendly cat",
			"password": "paws"
		}
	}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	configInfo := getConfigInfo()
	assert.Equal(t, configInfo.Indexd.URL, "http://indexd-service/index")
	assert.Equal(t, configInfo.Indexd.Username, "mr happy cat")
	assert.Equal(t, configInfo.Indexd.Password, "whiskers")

	assert.Equal(t, configInfo.MetadataService.URL, "http://revproxy-service/mds")
	assert.Equal(t, configInfo.MetadataService.Username, "mr friendly cat")
	assert.Equal(t, configInfo.MetadataService.Password, "paws")
}

func TestGetConfigInfoUsingOnlyMDSCreds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expecting getConfigInfo to panic since Indexd creds were not present in CONFIG_FILE")
		}
	}()

	jsonConfigInfo :=
		`
	{
		"metadataService": {
			"url": "http://revproxy-service/mds",
			"username": "mr friendly cat",
			"password": "paws"
		}
	}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	getConfigInfo()
}

func TestGetConfigInfoUsingExtraServiceInfo(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"indexd": {
			"url": "http://indexd-service/index",
			"username": "mr happy cat",
			"password": "whiskers"
		},
		"metadataService": {
			"url": "http://revproxy-service/mds",
			"username": "mr friendly cat",
			"password": "paws",
			"extra": "Always happy to see a human!"
		},
		"futureService": {
			"url": "http://future-service",
			"username": "mr futuristic cat",
			"password": "hover craft",
			"extra": "2100 is going to be purrfectly awesome!"
		}
	}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	configInfo := getConfigInfo()
	assert.Equal(t, configInfo.Indexd.URL, "http://indexd-service/index")
	assert.Equal(t, configInfo.Indexd.Username, "mr happy cat")
	assert.Equal(t, configInfo.Indexd.Password, "whiskers")

	assert.Equal(t, configInfo.MetadataService.URL, "http://revproxy-service/mds")
	assert.Equal(t, configInfo.MetadataService.Username, "mr friendly cat")
	assert.Equal(t, configInfo.MetadataService.Password, "paws")
}
