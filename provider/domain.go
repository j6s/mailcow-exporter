package provider

import (
	"encoding/json"

	"github.com/j6s/mailcow-exporter/mailcowApi"
	"github.com/prometheus/client_golang/prometheus"
)

// Domain Provider. This provider uses the `/api/v1/get/domain/all`
// endpoint in order to gather metrics.
type Domain struct{}

type domainItem struct {
	Domain        string      `json:"domain_name"`
	Active        json.Number `json:"active"`
	Mailboxes     json.Number `json:"mboxes_in_domain"`
	MaxMailboxes  json.Number `json:"max_num_mboxes_for_domain"`
	Aliases       json.Number `json:"aliases_in_domain"`
	MaxAliases    json.Number `json:"max_num_aliases_for_domain"`
	Quota         json.Number `json:"max_quota_for_domain"`
	QuotaUsed     json.Number `json:"bytes_total"`
	Messages      json.Number `json:"msgs_total"`
}

// All domain gauges have the same options anyways.
func domainGauge(name string, description string, host string) prometheus.GaugeVec {
	return *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        name,
		Help:        description,
		ConstLabels: map[string]string{"host": host},
	}, []string{"domain"})
}

func (domain Domain) Provide(api mailcowApi.MailcowApiClient) ([]prometheus.Collector, error) {
	active := domainGauge("mailcow_domain_active", "Active flag for this domain", api.Host)
	mailboxes := domainGauge("mailcow_domain_mailboxes", "Current mailboxes count for the domain", api.Host)
	maxMailboxes := domainGauge("mailcow_domain_max_mailboxes", "Maximum amount of mailboxes for the domain", api.Host)
	aliases := domainGauge("mailcow_domain_aliases", "Current aliases count for the domain", api.Host)
	maxAliases := domainGauge("mailcow_domain_max_aliases", "Maximum amount of aliases for the domain", api.Host)
	quotaAllowed := domainGauge("mailcow_domain_quota_allowed", "Aggregate quota maximum for the domain in bytes", api.Host)
	quotaUsed := domainGauge("mailcow_domain_quota_used", "Current size of the domain in bytes", api.Host)
	messages := domainGauge("mailcow_domain_messages", "Number of messages in for the domain mailboxes", api.Host)
	collectors := []prometheus.Collector{active, mailboxes, maxMailboxes, aliases, maxAliases, quotaAllowed, quotaUsed, messages}

	body := make([]domainItem, 0)
	err := api.Get("api/v1/get/domain/all", &body)
	if err != nil {
		return collectors, err
	}

	for _, d := range body {
		valueActive, err := d.Active.Float64()
		if err != nil {
			return collectors, err
		}

		valueMailboxes, err := d.Mailboxes.Float64()
		if err != nil {
			return collectors, err
		}

		valueMaxMailboxes, err := d.MaxMailboxes.Float64()
		if err != nil {
			return collectors, err
		}

		valueAliases, err := d.Aliases.Float64()
		if err != nil {
			return collectors, err
		}

		valueMaxAliases, err := d.MaxAliases.Float64()
		if err != nil {
			return collectors, err
		}

		valueQuota, err := d.Quota.Float64()
		if err != nil {
			return collectors, err
		}

		valueQuotaUsed, err := d.QuotaUsed.Float64()
		if err != nil {
			return collectors, err
		}

		valueMessages, err := d.Messages.Float64()
		if err != nil {
			return collectors, err
		}

		active.WithLabelValues(d.Domain).Set(valueActive)
		mailboxes.WithLabelValues(d.Domain).Set(valueMailboxes)
		maxMailboxes.WithLabelValues(d.Domain).Set(valueMaxMailboxes)
		aliases.WithLabelValues(d.Domain).Set(valueAliases)
		maxAliases.WithLabelValues(d.Domain).Set(valueMaxAliases)
		quotaAllowed.WithLabelValues(d.Domain).Set(valueQuota)
		quotaUsed.WithLabelValues(d.Domain).Set(valueQuotaUsed)
		messages.WithLabelValues(d.Domain).Set(valueMessages)
	}

	return collectors, nil
}
