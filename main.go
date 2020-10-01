package main

import (
	"log"
	"os"

	"github.com/uc-cdis/indexs3client/handlers"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	handlers.IndexS3Object(os.Getenv("INPUT_URL"))
}
