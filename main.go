package main

import (
	"os"

	"github.com/uc-cdis/indexs3client/handlers"
)

func main() {

	handlers.IndexS3Object(os.Getenv("INPUT_URL"))

}
