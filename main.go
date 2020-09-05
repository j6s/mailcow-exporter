package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/j6s/mailcow-exporter/provider"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	apiKey = flag.String("api-key", "", "API Key to use for the requests")
	host   = flag.String("host", "", "The host of the mailcow instance")
	listen = flag.String("listen", ":9099", "Host and port to listen on")
)

// A Provider is the common abstraction over collection of metrics in this
// exporter. It can provide one or more prometheus collectors (e.g. gauges,
// histograms, ...) that are updated every time the `Update` method is called.
// Be sure to keep a copy of the collectors returned by `GetCollectors`
// in your provider in order to update that same instance.
type Provider interface {
	GetCollectors() []prometheus.Collector
	Update(mailcowApi.MailcowApiClient)
}

// Provider setup. Every provider in this array will be used for gathering metrics.
var (
	providers = []Provider{
		provider.NewMailq(),
		provider.NewMailbox(),
		provider.NewQuarantine(),
	}
)

func init() {
	// Command line argument parsing
	flag.Parse()
	if *apiKey == "" || *host == "" {
		log.Fatal("Both --api-key and --host must be specified")
	}

	// Registrationg of collectors
	for _, provider := range providers {
		for _, collector := range provider.GetCollectors() {
			prometheus.MustRegister(collector)
		}
	}
}

func main() {
	apiClient := mailcowApi.MailcowApiClient{
		Host:   *host,
		ApiKey: *apiKey,
	}

	handler := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	)

	http.HandleFunc("/metrics", func(response http.ResponseWriter, request *http.Request) {
		for _, provider := range providers {
			provider.Update(apiClient)
		}
		handler.ServeHTTP(response, request)
	})

	log.Printf("Starting to listen on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
