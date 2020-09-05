package provider

import (
	"strconv"

	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Mailbox Provider. This provider uses the `/api/v1/get/mailbox/all`
// endpoint in order to gather metrics.
type Mailbox struct{}

type mailboxItem struct {
	Username      string `json:"username"`
	LastImapLogin string `json:"last_imap_login"`
	Quota         int    `json:"quota"`
	QuotaUsed     int    `json:"quota_used"`
	Messages      int    `json:"messages"`
}

// All mailbox gauges have the same options anyways.
func mailboxGauge(name string, description string, host string) prometheus.GaugeVec {
	return *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        name,
		Help:        description,
		ConstLabels: map[string]string{"host": host},
	}, []string{"mailbox"})
}

func (mailbox Mailbox) Provide(api mailcowApi.MailcowApiClient) ([]prometheus.Collector, error) {
	lastLogin := mailboxGauge("mailcow_mailbox_last_login", "Timestamp of the last IMAP login for this mailbox", api.Host)
	quotaAllowed := mailboxGauge("mailcow_mailbox_quota_allowed", "Quota maximum for the mailbox in bytes", api.Host)
	quotaUsed := mailboxGauge("mailcow_mailbox_quota_used", "Current syze of the mailbox in bytes", api.Host)
	messages := mailboxGauge("mailcow_mailbox_messages", "Number of messages in the mailbox", api.Host)

	body := make([]mailboxItem, 0)
	err := api.Get("api/v1/get/mailbox/all", &body)
	if err != nil {
		return []prometheus.Collector{}, err
	}

	for _, m := range body {
		lastLoginTimestamp, _ := strconv.ParseFloat(m.LastImapLogin, 64)
		lastLogin.WithLabelValues(m.Username).Set(lastLoginTimestamp)
		quotaAllowed.WithLabelValues(m.Username).Set(float64(m.Quota))
		quotaUsed.WithLabelValues(m.Username).Set(float64(m.QuotaUsed))
		messages.WithLabelValues(m.Username).Set(float64(m.Messages))
	}

	return []prometheus.Collector{lastLogin, quotaAllowed, quotaUsed, messages}, nil
}
