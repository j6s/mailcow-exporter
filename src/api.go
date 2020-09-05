package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Common abstraction over api requests to mailcow.
// Given an endpoint, this method will do the HTTP request
// with the correct authentication and unserialize the JSON
// response into a given target reference.
//
// NOTE: This method relies on the global *host and *apiKey
//       existing variables
//
// TODO: Move into struct that is initialized with host and apiKey to
//       remove the reliance on global variables.
//
// FIXME: This method kills the program when an error is encountered.
//        there should be proper error handling here.
//
// Example:
// body := make([]apiItem, 0)
// apiRequest("api/v1/get/foo/all", &body)
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
