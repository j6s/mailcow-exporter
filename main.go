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
	listen = flag.String("listen", ":9099", "Host and port to listen on")
)

// A Provider is the common abstraction over collection of metrics in this
// exporter. It can provide one or more prometheus collectors (e.g. gauges,
// histograms, ...) that are updated every time the `Update` method is called.
// Be sure to keep a copy of the collectors returned by `GetCollectors`
// in your provider in order to update that same instance.
type Provider interface {
	Provide(mailcowApi.MailcowApiClient) ([]prometheus.Collector, error)
}

// Provider setup. Every provider in this array will be used for gathering metrics.
var (
	providers = []Provider{
		provider.Mailq{},
		provider.Mailbox{},
		provider.Quarantine{},
	}
)

func collectMetrics(host string, apiKey string) (*prometheus.Registry, error) {
	apiClient := mailcowApi.NewMailcowApiClient(host, apiKey)

	registry := prometheus.NewRegistry()
	for _, provider := range providers {
		collectors, err := provider.Provide(apiClient)
		if err != nil {
			return registry, err
		}

		for _, collector := range collectors {
			err = registry.Register(collector)
			if err != nil {
				return registry, err
			}
		}
	}

	for _, collector := range apiClient.Provide() {
		registry.Register(collector)
	}

	return registry, nil
}

func main() {
	flag.Parse()

	http.HandleFunc("/metrics", func(response http.ResponseWriter, request *http.Request) {
		host := request.URL.Query().Get("host")
		apiKey := request.URL.Query().Get("apiKey")
		if host == "" || apiKey == "" {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("Query parameters `host` & `apiKey` are required"))
			return
		}

		registry, err := collectMetrics(host, apiKey)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(err.Error()))
			return
		}

		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{},
		).ServeHTTP(response, request)
	})

	log.Printf("Starting to listen on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
