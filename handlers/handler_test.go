package handlers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

// return Indexd config
func makeImageConfigString() string {
	return `{
				"url": "http://indexd-service/",
				"username": "test",
				"password": "test"
		  	}`
}

// Test that Indexd config can be marshalled into IndexdInfo struct
func TestHandler(t *testing.T) {
	indexdInfo := new(IndexdInfo)
	if err := json.Unmarshal([]byte(makeImageConfigString()), indexdInfo); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, indexdInfo.URL, "http://indexd-service/")
	assert.Equal(t, indexdInfo.Username, "test")
	assert.Equal(t, indexdInfo.Password, "test")

}

// Test getConfigInfo function inputting Indexd and Metadata Service configs
func TestGetConfigInfoUsingIndexdAndMDSCreds(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"url": "http://indexd-service/index",
		"username": "mr happy cat",
		"password": "whiskers",
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

// Test that getConfigInfo function panics when no Indexd or Metadata Service
// config is present
func TestGetConfigInfoWithoutIndexdOrMDSCreds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expecting getConfigInfo to panic since Indexd and Metadata Service creds were not present in CONFIG_FILE")
		}
	}()

	jsonConfigInfo :=
		`
	{}
	`
	err := os.Setenv("CONFIG_FILE", jsonConfigInfo)
	if err != nil {
		t.Fatal(err)
	}
	getConfigInfo()
}

// Test that getConfigInfo function panics when no Metadata Service config is
// present
func TestGetConfigInfoWithoutMDSCreds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expecting getConfigInfo to panic since Metadata Service creds were not present in CONFIG_FILE")
		}
	}()

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
	getConfigInfo()
}

// Test that getConfigInfo function panics when no Indexd config is present
func TestGetConfigInfoWithoutIndexdCreds(t *testing.T) {
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

// Test getConfigInfo function inputting extra service config
func TestGetConfigInfoUsingExtraServiceInfo(t *testing.T) {
	jsonConfigInfo :=
		`
	{
		"url": "http://indexd-service/index",
		"username": "mr happy cat",
		"password": "whiskers",
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

func TestCreateUUID(t *testing.T) {
	keys := []struct {
		key      string
		uuid     string
		filename string
	}{
		{"dg.MD1R/00004ead-7ae8-48d9-9d87-c1e0eabbf651/TEST-12-AR/TEST-12-AR-16434369/01-05-2012-CT TEST TESTTEST TEST W-12234/4.1111111-ARTEST 1MM-758254/8-7654.dcm", "dg.MD1R/00004ead-7ae8-48d9-9d87-c1e0eabbf651", "TEST-12-AR/TEST-12-AR-16434369/01-05-2012-CT TEST TESTTEST TEST W-12234/4.1111111-ARTEST 1MM-758254/8-7654.dcm"},
		{"dg.NACD/da85ab42-53a0-4698-9b38-7ad59b770b47/flies/moth.txt", "dg.NACD/da85ab42-53a0-4698-9b38-7ad59b770b47", "flies/moth.txt"},
		{"dg.4825/da85ab42-53a0-4698-9b38-7ad59b770b47/flies/moth.txt", "dg.4825/da85ab42-53a0-4698-9b38-7ad59b770b47", "flies/moth.txt"},
		{"dg.F738/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test1.txt", "dg.F738/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test1.txt"},
		{"dg.80B6/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test2.txt", "dg.80B6/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test2.txt"},
		{"dg.EA80/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test3.txt", "dg.EA80/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test3.txt"},
		{"dg.5B0D/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test4.txt", "dg.5B0D/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test4.txt"},
		{"dg.4503/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test5.txt", "dg.4503/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test5.txt"},
		{"dg.373F/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test6.txt", "dg.373F/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test6.txt"},
		{"dg.7519/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test7.txt", "dg.7519/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test7.txt"},
		{"dg.7C5B/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test8.txt", "dg.7C5B/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test8.txt"},
		{"dg.0896/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test9.txt", "dg.0896/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test9.txt"},
		{"dg.F82A1A/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test10.txt", "dg.F82A1A/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test10.txt"},
		{"dg.4DFC/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test11.txt", "dg.4DFC/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test11.txt"},
		{"dg.712C/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test12.txt", "dg.712C/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test12.txt"},
		{"dg.6VTS/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test13.txt", "dg.6VTS/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test13.txt"},
		{"dg.63D5/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test14.txt", "dg.63D5/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test14.txt"},
		{"dg.414F/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test15.txt", "dg.414F/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test15.txt"},
		{"dg.6539/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test16.txt", "dg.6539/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test16.txt"},
		{"dg.ANV0/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test17.txt", "dg.ANV0/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test17.txt"},
		{"dg.MD1R/da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test18.txt", "dg.MD1R/da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test18.txt"},
		{"da85ab42-53a0-4698-9b38-7ad59b770b47/test19.txt", "da85ab42-53a0-4698-9b38-7ad59b770b47", "test19.txt"},
		{"dg.MD1R/da85ab42-53a0-4698-9b38-7ad59b770b47/test20.txt", "dg.MD1R/da85ab42-53a0-4698-9b38-7ad59b770b47", "test20.txt"},
		{"da85ab42-53a0-4698-9b38-7ad59b770b47/files/prefixtest/test21.txt", "da85ab42-53a0-4698-9b38-7ad59b770b47", "files/prefixtest/test21.txt"},
	}

	for _, key := range keys {
		uuid, filename, err := resolveUUID(key.key)
		if err != nil {
			t.Fatal(err)
		}
		if uuid != key.uuid {
			t.Errorf("The UUID is invalid")
		}
		if filename != key.filename {
			t.Errorf("The filename is invalid")
		}
	}
}
