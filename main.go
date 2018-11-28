package main

import (
	"log"
	"os"

	"github.com/uc-cdis/indexs3client/handlers"
)

func main() {
	argsWithProg := os.Args

	if len(argsWithProg) < 3 {
		log.Panicf("Required 3 parameters. Only %d provide", len(argsWithProg))
	}

	s3object := argsWithProg[1]
	indexURL := argsWithProg[2]

	handlers.IndexS3Object(s3object, indexURL)

}
