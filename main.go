package main

import (
	"fmt"
	"strings"

	"github.com/uc-cdis/indexs3client/handlers"
)

func main() {
	input := "hello world"

	fmt.Println("Input:", input)

	r := strings.NewReader(input)

	hashes := handlers.CalculateBasicHashes(r)

	fmt.Print(input, ",", hashes.Crc32c, ",", hashes.Md5, ",", hashes.Sha1, ",", hashes.Sha256, ",", hashes.Sha512, "\n")
}
