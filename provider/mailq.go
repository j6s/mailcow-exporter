package provider

import (
	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Mailq Provider. Use `NewMailq` to initialize this struct.
// This provider uses the `/api/v1/get/mailq/all` endpoint
// in order to gather metrics.
type Mailq struct {
	Gauge prometheus.GaugeVec
}

type queueResponseItem struct {
	QueueName string `json:"queue_name"`
	Sender    string
}

func NewMailq() Mailq {
	return Mailq{
		Gauge: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "mailcow_mailq",
		}, []string{"queue", "sender"}),
	}
}

func (mailq Mailq) GetCollectors() []prometheus.Collector {
	return []prometheus.Collector{mailq.Gauge}
}

func (mailq Mailq) Update(api mailcowApi.MailcowApiClient) {
	body := make([]queueResponseItem, 0)
	api.Get("api/v1/get/mailq/all", &body)

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
			mailq.Gauge.WithLabelValues(queueName, sender).Set(count)
		}
	}
}
