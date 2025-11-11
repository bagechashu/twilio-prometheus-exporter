package main

import (
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// TwilioCollector provides a minimal, low-cardinality set of metrics from Twilio.
type TwilioCollector struct {
	client *TwilioClient

	// Gauges (reset each scrape) representing counts/amounts observed in the requested window
	balanceGauge        *prometheus.GaugeVec // labels: currency
	usageGauge          *prometheus.GaugeVec // labels: category, usage_unit
	callsWindowGauge    *prometheus.GaugeVec // labels: status
	messagesWindowGauge *prometheus.GaugeVec // labels: status
	messagesFailedGauge *prometheus.GaugeVec // labels: error_code

	// Persistent counter for API errors
	apiErrorCounter *prometheus.CounterVec // labels: operation

	config Config

	// mutex to protect internal operations
	mu sync.Mutex
}

// NewTwilioCollector builds a collector with a small, well-defined label set to avoid cardinality explosions.
func NewTwilioCollector(config Config) *TwilioCollector {
	client := NewTwilioClient(config.TwilioAccountSID, config.TwilioAuthToken, config)

	tc := &TwilioCollector{
		client: client,
		config: config,
		balanceGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_account_balance",
			Help: "Current balance of Twilio account (numeric).",
		}, []string{"currency"}),
		usageGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_usage_amount",
			Help: "Amount used for usage records in the requested window.",
		}, []string{"category", "usage_unit"}),
		callsWindowGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_calls_window_count",
			Help: "Number of calls observed in the requested window, grouped by status.",
		}, []string{"status"}),
		messagesWindowGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_messages_window_count",
			Help: "Number of messages observed in the requested window, grouped by status.",
		}, []string{"status"}),
		messagesFailedGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_messages_window_failed_count",
			Help: "Number of failed/undelivered messages in the requested window, grouped by error code.",
		}, []string{"error_code"}),
		apiErrorCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "twilio_api_errors_total",
			Help: "Total number of Twilio API errors by operation.",
		}, []string{"operation"}),
	}

	return tc
}

func (tc *TwilioCollector) Describe(ch chan<- *prometheus.Desc) {
	// Let the individual metric descriptors be handled by their collectors
}

func (tc *TwilioCollector) Collect(ch chan<- prometheus.Metric) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Balance
	balanceObjs, err := tc.client.FetchBalance()
	if err != nil {
		logrus.WithError(err).Error("FetchBalance failed")
		tc.apiErrorCounter.WithLabelValues("FetchBalance").Inc()
	} else {
		// reset and set latest
		tc.balanceGauge.Reset()
		for _, b := range balanceObjs {
			var val float64 = 0
			if b.Balance != nil {
				if parsed, err := strconv.ParseFloat(*b.Balance, 64); err == nil {
					val = parsed
				}
			}
			currency := ""
			if b.Currency != nil {
				currency = *b.Currency
			}
			tc.balanceGauge.WithLabelValues(currency).Set(val)
		}
	}

	// Usage records (sum by category + usage_unit)
	usageRecords, err := tc.client.FetchUsageRecordsToday()
	if err != nil {
		logrus.WithError(err).Error("FetchUsageRecordsToday failed")
		tc.apiErrorCounter.WithLabelValues("FetchUsageRecordsToday").Inc()
	} else {
		tc.usageGauge.Reset()
		sums := make(map[string]float64) // key: category\xffusage_unit
		for _, r := range usageRecords {
			cat := ""
			if r.Category != nil {
				cat = *r.Category
			}
			unit := ""
			if r.UsageUnit != nil {
				unit = *r.UsageUnit
			}
			key := cat + "\xff" + unit
			var v float64 = 0
			if r.Usage != nil {
				if parsed, err := strconv.ParseFloat(*r.Usage, 64); err == nil {
					v = parsed
				}
			}
			sums[key] += v
		}
		for k, v := range sums {
			parts := strings.SplitN(k, "\xff", 2)
			cat := ""
			unit := ""
			if len(parts) >= 1 {
				cat = parts[0]
			}
			if len(parts) == 2 {
				unit = parts[1]
			}
			tc.usageGauge.WithLabelValues(cat, unit).Set(v)
		}
	}

	// Calls: count by status
	calls, err := tc.client.FetchCalls()
	if err != nil {
		logrus.WithError(err).Error("FetchCalls failed")
		tc.apiErrorCounter.WithLabelValues("FetchCalls").Inc()
	} else {
		tc.callsWindowGauge.Reset()
		counts := make(map[string]int)
		for _, c := range calls {
			status := "unknown"
			if c.Status != nil && *c.Status != "" {
				status = *c.Status
			}
			counts[status]++
		}
		for status, cnt := range counts {
			tc.callsWindowGauge.WithLabelValues(status).Set(float64(cnt))
		}
	}

	// Messages: count by status, and failed by error code
	messages, err := tc.client.FetchMessages()
	if err != nil {
		logrus.WithError(err).Error("FetchMessages failed")
		tc.apiErrorCounter.WithLabelValues("FetchMessages").Inc()
	} else {
		tc.messagesWindowGauge.Reset()
		tc.messagesFailedGauge.Reset()
		statusCounts := make(map[string]int)
		failedCounts := make(map[string]int)
		for _, m := range messages {
			status := "unknown"
			if m.Status != nil && *m.Status != "" {
				status = *m.Status
			}
			statusCounts[status]++

			if m.ErrorCode != nil {
				code := strconv.Itoa(*m.ErrorCode)
				failedCounts[code]++
			}
		}
		for s, cnt := range statusCounts {
			tc.messagesWindowGauge.WithLabelValues(s).Set(float64(cnt))
		}
		for code, cnt := range failedCounts {
			tc.messagesFailedGauge.WithLabelValues(code).Set(float64(cnt))
		}
	}

	// Collect all metrics onto the channel
	tc.balanceGauge.Collect(ch)
	tc.usageGauge.Collect(ch)
	tc.callsWindowGauge.Collect(ch)
	tc.messagesWindowGauge.Collect(ch)
	tc.messagesFailedGauge.Collect(ch)
	tc.apiErrorCounter.Collect(ch)
}
