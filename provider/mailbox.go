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

func (mailbox Mailbox) Provide(api mailcowApi.MailcowApiClient) ([]prometheus.Collector, error) {
	lastLogin := *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_last_login"}, []string{"host", "mailbox"})
	quotaAllowed := *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_quota_allowed"}, []string{"host", "mailbox"})
	quotaUsed := *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_quota_used"}, []string{"host", "mailbox"})
	messages := *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_messages"}, []string{"host", "mailbox"})

	body := make([]mailboxItem, 0)
	err := api.Get("api/v1/get/mailbox/all", &body)
	if err != nil {
		return []prometheus.Collector{}, err
	}

	for _, m := range body {
		lastLoginTimestamp, _ := strconv.ParseFloat(m.LastImapLogin, 64)
		lastLogin.WithLabelValues(api.Host, m.Username).Set(lastLoginTimestamp)
		quotaAllowed.WithLabelValues(api.Host, m.Username).Set(float64(m.Quota))
		quotaUsed.WithLabelValues(api.Host, m.Username).Set(float64(m.QuotaUsed))
		messages.WithLabelValues(api.Host, m.Username).Set(float64(m.Messages))
	}

	return []prometheus.Collector{lastLogin, quotaAllowed, quotaUsed, messages}, nil
}
