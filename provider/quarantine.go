package provider

import (
	"time"

	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Quarantine Provider. Use `NewQuarantine` to initialize this struct.
// This provider uses the `/api/v1/get/quarantine/all` endpoint
// in order to gather metrics about quarantined mails.
type Quarantine struct {
	count prometheus.GaugeVec
	score prometheus.HistogramVec
	age   prometheus.HistogramVec
}

type quarantineItem struct {
	VirusFlag int     `json:"virus_flag"`
	Score     float64 `json:"score"`
	Recipient string  `json:"rcpt"`
	Created   int64   `json:"created"`
}

func NewQuarantine() Quarantine {
	return Quarantine{
		count: *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_quarantine_count"}, []string{"recipient", "is_virus"}),
		score: *prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "mailcow_quarantine_score",
			Buckets: []float64{0, 10, 20, 40, 60, 80, 100},
		}, []string{"recipient"}),
		age: *prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "mailcow_quarantine_age",
			Help: "Age of quarantined items in seconds",
			Buckets: []float64{
				(60 * 60 * 3),       // 3 hours
				(60 * 60 * 12),      // 12 hours
				(60 * 60 * 24),      // 1 day
				(3 * 60 * 60 * 24),  // 3 days
				(7 * 60 * 60 * 24),  // 7 days
				(14 * 60 * 60 * 24), // 14 days
				(30 * 60 * 60 * 24), // 30 days
			},
		}, []string{"recipient"}),
	}
}

func (quarantine Quarantine) GetCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		quarantine.count,
		quarantine.score,
		quarantine.age,
	}
}

func (quarantine Quarantine) Update(api mailcowApi.MailcowApiClient) {
	quarantine.age.Reset()
	quarantine.score.Reset()

	body := make([]quarantineItem, 0)
	api.Get("api/v1/get/quarantine/all", &body)

	virus := make(map[string]int)
	notVirus := make(map[string]int)
	for _, q := range body {
		if _, ok := virus[q.Recipient]; !ok {
			virus[q.Recipient] = 0
		}
		if _, ok := notVirus[q.Recipient]; !ok {
			notVirus[q.Recipient] = 0
		}

		if q.VirusFlag == 1 {
			virus[q.Recipient]++
		} else {
			notVirus[q.Recipient]++
		}

		age := time.Now().Unix() - q.Created
		quarantine.age.WithLabelValues(q.Recipient).Observe(float64(age))
		quarantine.score.WithLabelValues(q.Recipient).Observe(float64(q.Score))
	}

	for recipient, count := range virus {
		quarantine.count.WithLabelValues(recipient, "1").Set(float64(count))
	}
	for recipient, count := range notVirus {
		quarantine.count.WithLabelValues(recipient, "0").Set(float64(count))
	}
}
