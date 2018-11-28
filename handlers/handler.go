package handlers

import (
	"fmt"
	"net/url"
	"strings"
)

func IndexS3Object(s3object string, indexURL string) error {

	u, _ := url.Parse(s3object)
	bucket, key := u.Host, u.Path

	client, err := CreateNewAwsClient()
	if err != nil {
		return err
	}
	DownloadObjectFromS3(client, bucket, key, "./tmp.obj")
	r := strings.NewReader("./tmp.obj")
	hashes := CalculateBasicHashes(r)
	fmt.Print(hashes.Crc32c, ",", hashes.Md5, ",", hashes.Sha1, ",", hashes.Sha256, ",", hashes.Sha512, "\n")
	return nil
}
