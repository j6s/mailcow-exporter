package provider

import (
	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Rspamd provider, fetching data from `/api/v1/get/logs/rspamd-stats`
type Rspamd struct{}

type RspamdResponse struct {
	Scanned int `json:"scanned"`
	Learned int `json:"learned"`

	Actions map[string]int `json:"actions"`

	SpamCount          int `json:"spam_count"`
	HamCount           int `json:"ham_count"`
	Connections        int `json:"connections"`
	ControlConnections int `json:"control_connections"`
	PoolsAllocated     int `json:"pools_allocated"`
	PoolsFreed         int `json:"pools_freed"`

	BytesAllocated        int `json:"bytes_allocated"`
	ChunksAllocated       int `json:"chunks_allocated"`
	SharedChunksAllocated int `json:"shared_chunks_allocated"`
	ChunksFreed           int `json:"chunks_freed"`
	ChunksOversized       int `json:"chunks_oversized"`
	Fragmented            int `json:"fragmented"`
	TotalLearns           int `json:"total_learns"`

	FuzzyHashes map[string]int `json:"fuzzy_hashes"`
}

func (rspamd Rspamd) simpleGauge(
	host string,
	name string,
	value int,
) prometheus.Collector {
	return prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        name,
		ConstLabels: map[string]string{"host": host},
	}, func() float64 {
		return float64(value)
	})
}

func (rspamd Rspamd) extractActions(host string, stats RspamdResponse) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_rspamd_action",
		Help:        "Number of items for which a certain action has been taken",
		ConstLabels: map[string]string{"host": host},
	}, []string{"action"})

	for action, number := range stats.Actions {
		gauge.WithLabelValues(action).Set(float64(number))
	}

	return gauge
}

func (rspamd Rspamd) extractFuzzyHashes(host string, stats RspamdResponse) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_rspamd_fuzzy_hashes",
		ConstLabels: map[string]string{"host": host},
	}, []string{"action"})

	for action, number := range stats.FuzzyHashes {
		gauge.WithLabelValues(action).Set(float64(number))
	}

	return gauge
}

func (rspamd Rspamd) extractClassification(host string, stats RspamdResponse) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_rspamd_classification",
		ConstLabels: map[string]string{"host": host},
	}, []string{"classification"})

	gauge.WithLabelValues("spam").Set(float64(stats.SpamCount))
	gauge.WithLabelValues("ham").Set(float64(stats.HamCount))

	return gauge
}

func (rspamd Rspamd) extractPools(host string, stats RspamdResponse) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_rspamd_pools",
		ConstLabels: map[string]string{"host": host},
	}, []string{"state"})

	gauge.WithLabelValues("allocated").Set(float64(stats.PoolsAllocated))
	gauge.WithLabelValues("freed").Set(float64(stats.PoolsFreed))

	return gauge
}

func (rspamd Rspamd) extractChunks(host string, stats RspamdResponse) prometheus.Collector {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "mailcow_rspamd_chunks",
		ConstLabels: map[string]string{"host": host},
	}, []string{"state"})

	gauge.WithLabelValues("allocated").Set(float64(stats.ChunksAllocated))
	gauge.WithLabelValues("freed").Set(float64(stats.ChunksFreed))
	gauge.WithLabelValues("oversized").Set(float64(stats.ChunksOversized))
	gauge.WithLabelValues("shared").Set(float64(stats.SharedChunksAllocated))

	return gauge
}

func (rspamd Rspamd) Provide(api mailcowApi.MailcowApiClient) ([]prometheus.Collector, error) {
	body := RspamdResponse{}
	err := api.Get("api/v1/get/logs/rspamd-stats", &body)

	collectors := []prometheus.Collector{
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_scanned", body.Scanned),
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_learned", body.Learned),
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_connections", body.Connections),
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_control_connections", body.ControlConnections),
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_bytes_allocated", body.BytesAllocated),
		rspamd.simpleGauge(api.Host, "mailcow_rspamd_fragmented", body.Fragmented),
		rspamd.extractChunks(api.Host, body),
		rspamd.extractPools(api.Host, body),
		rspamd.extractClassification(api.Host, body),
		rspamd.extractActions(api.Host, body),
		rspamd.extractFuzzyHashes(api.Host, body),
	}

	return collectors, err
}
