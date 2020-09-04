package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func apiRequest(endpoint string, target interface{}) {
	url := fmt.Sprintf("https://%s/%s", *host, endpoint)
	log.Print(url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("X-Api-Key", *apiKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(body, target)
}
