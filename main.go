package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/j6s/mailcow-exporter/provider"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultHost   string
	defaultApiKey string
	listen        string
	scheme        string 
	defaultScheme string
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
		provider.Rspamd{},
		provider.Domain{},
	}
)

func parseFlagsAndEnv() {
	envHost, _ := os.LookupEnv("MAILCOW_EXPORTER_HOST")
	envApiKey, _ := os.LookupEnv("MAILCOW_EXPORTER_API_KEY")
	defaultListen, _ := os.LookupEnv("MAILCOW_EXPORTER_LISTEN")
	if defaultListen == "" {
		defaultListen = ":9099"
	}


	scheme, _ := os.LookupEnv("MAILCOW_EXPORTER_SCHEME")
	if defaultScheme == "" {
		defaultScheme = "https"
	}
	

	flag.StringVar(&defaultHost, "defaultHost", envHost, "The defaultHost to connect to. Defaults to the MAILCOW_EXPORTER_HOST environment variable")
	flag.StringVar(&defaultApiKey, "apikey", envApiKey, "The API key to use for connection. Defaults to the MAILCOW_EXPORTER_API_KEY environment variable")
	flag.StringVar(&listen, "listen", defaultListen, "Host and port to listen on. Defaults to the MAILCOW_EXPORTER_LISTEN environment variable or ':9099' otherwise")
	flag.StringVar(&scheme, "scheme", defaultScheme, "Default connection scheme. Defaults to the MAILCOW_EXPORTER_SCHEME environment variable or 'https' otherwise")
	
	flag.Parse()


}

func collectMetrics(scheme string, host string, apiKey string) *prometheus.Registry {
	apiClient := mailcowApi.NewMailcowApiClient(scheme, host, apiKey)

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
	parseFlagsAndEnv()

	http.HandleFunc("/metrics", func(response http.ResponseWriter, request *http.Request) {
		
		host := request.URL.Query().Get("host")
		apiKey := request.URL.Query().Get("apiKey")
		scheme := request.URL.Query().Get("scheme")
		
		if host == "" {
			host = defaultHost
		}
		if apiKey == "" {
			apiKey = defaultApiKey
		}
		if scheme == "" {
			scheme = defaultScheme 
		}

		if host == "" {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("Query parameter `host` is required, since it is not defined by flags or environment"))
			return
		}
		if apiKey == "" {
			response.WriteHeader(http.StatusUnauthorized)
			response.Write([]byte("Query parameter `apiKey` is required, since it is not defined by flags or environment"))
			return
		}




		registry := collectMetrics(scheme, host, apiKey)

		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{},
		).ServeHTTP(response, request)
	})

	log.Printf("Starting to listen on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
