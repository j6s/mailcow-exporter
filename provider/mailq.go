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
		Name: "mailcow_mailq",
	}, []string{"queue", "sender"})

	body := make([]queueResponseItem, 0)
	err := api.Get("api/v1/get/mailq/all", &body)
	if err != nil {
		return []prometheus.Collector{}, err
	}

	queue := make(map[string]map[string]float64)
	for _, item := range body {
		if _, ok := queue[item.QueueName]; !ok {
			queue[item.QueueName] = make(map[string]float64)
		}
		if _, ok := queue[item.QueueName][item.Sender]; !ok {
			queue[item.QueueName][item.Sender] = 0
		}

		queue[item.QueueName][item.Sender]++
	}

	for queueName, senders := range queue {
		for sender, count := range senders {
			gauge.WithLabelValues(queueName, sender).Set(count)
		}
	}

	return []prometheus.Collector{gauge}, nil
}
