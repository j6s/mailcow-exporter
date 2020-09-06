package main

import (
	"flag"
	"fmt"
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
		provider.Container{},
	}
)

func collectMetrics(host string, apiKey string) *prometheus.Registry {
	apiClient := mailcowApi.NewMailcowApiClient(host, apiKey)

	success := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_exporter_success",
		ConstLabels: map[string]string{"host": host},
	}, []string{"provider"})

	registry := prometheus.NewRegistry()
	registry.Register(success)

	for _, provider := range providers {
		providerSuccess := true
		collectors, err := provider.Provide(apiClient)
		if err != nil {
			providerSuccess = false
			log.Printf(
				"Error while updating metrics of %T:\n%s",
				provider,
				err.Error(),
			)
		}

		for _, collector := range collectors {
			err = registry.Register(collector)
			if err != nil {
				providerSuccess = false
				log.Printf(
					"Error while updating metrics of %T:\n%s",
					provider,
					err.Error(),
				)
			}
		}

		if providerSuccess {
			success.WithLabelValues(fmt.Sprintf("%T", provider)).Set(1.0)
		} else {
			success.WithLabelValues(fmt.Sprintf("%T", provider)).Set(0.0)
		}
	}

	for _, collector := range apiClient.Provide() {
		registry.Register(collector)
	}

	return registry
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

		registry := collectMetrics(host, apiKey)

		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{},
		).ServeHTTP(response, request)
	})

	log.Printf("Starting to listen on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
