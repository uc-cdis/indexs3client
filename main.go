package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/uc-cdis/indexs3client/handlers"
)

type InputData struct {
	URL string `url`
}

func main() {
	if os.Getenv("INPUT_URL") != "" {
		handlers.IndexS3Object(os.Getenv("INPUT_URL"))
	} else if os.Getenv("INPUT_DATA") != "" {
		var s = os.Getenv("INPUT_DATA")
		fmt.Println(s)
		inputData := new(InputData)
		if err := json.Unmarshal([]byte(os.Getenv("INPUT_DATA")), inputData); err != nil {
			fmt.Println("Enviroinment variable CONFIG_FILE is not set correctly")
			return
		}
		fmt.Printf("URL: %s", inputData.URL)
		handlers.IndexS3Object(inputData.URL)

	} else {
		fmt.Println("No URL found")
	}

}
