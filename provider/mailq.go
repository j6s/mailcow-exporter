package provider

import (
	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Mailq provider.
// This provider uses the `/api/v1/get/mailq/all` endpoint
// in order to gather metrics.
type Mailq struct{}

type queueResponseItem struct {
	QueueName string `json:"queue_name"`
	Sender    string
}

func (mailq Mailq) Provide(api mailcowApi.MailcowApiClient) ([]prometheus.Collector, error) {
	gauge := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_mailq",
		ConstLabels: map[string]string{"host": api.Host},
	}, []string{"queue", "sender"})

	body := make([]queueResponseItem, 0)
	err := api.Get("api/v1/get/mailq/all", &body)
	if err != nil {
		return []prometheus.Collector{}, err
	}

	for _, item := range body {
		gauge.WithLabelValues(item.QueueName, item.Sender).Inc()
	}

	return []prometheus.Collector{gauge}, nil
}
