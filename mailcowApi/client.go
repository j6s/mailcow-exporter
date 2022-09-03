package mailcowApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Client for mailcow API
type MailcowApiClient struct {
	Scheme       string
	Host         string
	ApiKey       string
	ResponseTime prometheus.GaugeVec
	ResponseSize prometheus.GaugeVec
	Success      prometheus.GaugeVec
}

func NewMailcowApiClient(scheme string, host string, apiKey string) MailcowApiClient {
	return MailcowApiClient{
		Scheme: scheme,
		Host:   host,
		ApiKey: apiKey,
		ResponseTime: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "mailcow_api_response_time",
			Help:        "Response time of the API in milliseconds (1/1000s of a second)",
			ConstLabels: map[string]string{"host": host},
		}, []string{"endpoint", "statusCode"}),
		ResponseSize: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "mailcow_api_response_size",
			Help:        "Size of API response in bytes",
			ConstLabels: map[string]string{"host": host},
		}, []string{"endpoint", "statusCode"}),
		Success: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "mailcow_api_success",
			Help:        "1, if request was sucessful, 0 if not",
			ConstLabels: map[string]string{"host": host},
		}, []string{"endpoint"}),
	}
}

// Given an endpoint, this method will do the HTTP request
// with the correct authentication and unserialize the JSON
// response into a given target reference.
func (api MailcowApiClient) Get(endpoint string, target interface{}) error {
	url := fmt.Sprintf("%s://%s/%s", api.Scheme, api.Host, endpoint)
	log.Print(url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"Could not prepare API request to `%s`: %#v",
			endpoint,
			err.Error(),
		)
	}

	request.Header.Add("X-Api-Key", api.ApiKey)
	start := time.Now()

	// API Request
	response, err := (&http.Client{}).Do(request)
	if err != nil {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"could not execute API request to `%s`: %#v",
			endpoint,
			err.Error(),
		)
	}

	// Metric collection about the API request
	statusCodeString := strconv.FormatInt(int64(response.StatusCode), 10)
	api.ResponseTime.
		WithLabelValues(endpoint, statusCodeString).
		Set(float64(time.Since(start).Milliseconds()))

	// API Request error handling
	if err != nil {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"API Request to endpoint `%s` failed: \n%s",
			endpoint,
			err.Error(),
		)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"Could not read API response body from endpoint `%s`: \n%s",
			endpoint,
			err.Error(),
		)
	}

	api.ResponseSize.
		WithLabelValues(endpoint, statusCodeString).
		Set(float64(len(body)))

	if response.StatusCode != 200 {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"Received %d response from endpoint `%s`: \n\nResponse body received: \n%s",
			response.StatusCode,
			endpoint,
			body,
		)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		api.Success.WithLabelValues(endpoint).Set(0.0)
		return fmt.Errorf(
			"Could not parse JSON response from endpoint `%s`: \n%s \n\nResponse body received: \n%s",
			endpoint,
			err.Error(),
			body,
		)
	}

	api.Success.WithLabelValues(endpoint).Set(1.0)
	return nil
}

// Provides (meta) metrics about API endpoints
func (api MailcowApiClient) Provide() []prometheus.Collector {
	return []prometheus.Collector{api.ResponseSize, api.ResponseTime, api.Success}
}
