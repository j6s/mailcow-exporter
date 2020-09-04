package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	apiKey = flag.String("api-key", "", "API Key to use for the requests")
	host   = flag.String("host", "", "The host of the mailcow instance")
	listen = flag.String("listen", ":9099", "Host and port to listen on")
)

type QueueResponseItem struct {
	QueueName string `json:"queue_name"`
	Sender    string
}

type Provider interface {
	GetCollectors() []prometheus.Collector
	Update()
}

func main() {
	flag.Parse()
	if *apiKey == "" || *host == "" {
		log.Fatal("Both --api-key and --host must be specified")
	}

	providers := []Provider{
		NewMailq(),
		NewMailbox(),
		NewQuarantine(),
	}

	for _, provider := range providers {
		for _, collector := range provider.GetCollectors() {
			prometheus.MustRegister(collector)
		}
	}

	handler := promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	)

	// Expose the registered metrics via HTTP.
	http.HandleFunc("/metrics", func(response http.ResponseWriter, request *http.Request) {
		for _, provider := range providers {
			provider.Update()
		}
		handler.ServeHTTP(response, request)
	})

	log.Printf("Starting to listen on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
