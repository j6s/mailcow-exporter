package main

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// Mailbox Provider. Use `NewMailbox` to initialize this struct.
// This provider uses the `/api/v1/get/mailbox/all` endpoint
// in order to gather metrics.
type Mailbox struct {
	lastLogin    prometheus.GaugeVec
	quotaAllowed prometheus.GaugeVec
	quotaUsed    prometheus.GaugeVec
	messages     prometheus.GaugeVec
}

type mailboxItem struct {
	Username      string `json:"username"`
	LastImapLogin string `json:"last_imap_login"`
	Quota         int    `json:"quota"`
	QuotaUsed     int    `json:"quota_used"`
	Messages      int    `json:"messages"`
}

func NewMailbox() Mailbox {
	return Mailbox{
		lastLogin:    *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_last_login"}, []string{"mailbox"}),
		quotaAllowed: *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_quota_allowed"}, []string{"mailbox"}),
		quotaUsed:    *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_quota_used"}, []string{"mailbox"}),
		messages:     *prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mailcow_mailbox_messages"}, []string{"mailbox"}),
	}
}

func (mailbox Mailbox) GetCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		mailbox.lastLogin,
		mailbox.quotaAllowed,
		mailbox.quotaUsed,
		mailbox.messages,
	}
}

func (mailbox Mailbox) Update() {
	body := make([]mailboxItem, 0)
	apiRequest("api/v1/get/mailbox/all", &body)

	for _, m := range body {
		lastLogin, _ := strconv.ParseFloat(m.LastImapLogin, 64)
		mailbox.lastLogin.WithLabelValues(m.Username).Set(lastLogin)
		mailbox.quotaAllowed.WithLabelValues(m.Username).Set(float64(m.Quota))
		mailbox.quotaUsed.WithLabelValues(m.Username).Set(float64(m.QuotaUsed))
		mailbox.messages.WithLabelValues(m.Username).Set(float64(m.Messages))
	}
}
