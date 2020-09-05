package mailcowApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Client for mailcow API
type MailcowApiClient struct {
	Host   string
	ApiKey string
}

// Given an endpoint, this method will do the HTTP request
// with the correct authentication and unserialize the JSON
// response into a given target reference.
func (api MailcowApiClient) Get(endpoint string, target interface{}) error {
	url := fmt.Sprintf("https://%s/%s", api.Host, endpoint)
	log.Print(url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("X-Api-Key", api.ApiKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}
