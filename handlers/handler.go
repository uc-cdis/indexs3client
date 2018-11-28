package handlers

import (
	"fmt"
	"strings"
)

func IndexS3Object(bucket string, key string) error {
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
