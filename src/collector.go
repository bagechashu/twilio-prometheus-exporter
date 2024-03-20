package main

import (
	"reflect"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioCollector struct {
	client        *TwilioClient
	balanceMetric *prometheus.GaugeVec
	usageMetric   *prometheus.GaugeVec
	callMetric    *prometheus.GaugeVec
	messageMetric *prometheus.GaugeVec
	config        Config
}

func NewTwilioCollector(config Config) *TwilioCollector {
	client := NewTwilioClient(config.TwilioAccountSID, config.TwilioAuthToken, config)

	// Create Twilio collector
	tc := &TwilioCollector{
		client: client,
		config: config,
		balanceMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_account_balance",
			Help: "Current balance of Twilio account.",
		}, addLabelsFromStruct(reflect.TypeOf(openapi.ApiV2010Balance{}))),
		usageMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_usage",
			Help: "The amount used to bill usage and measured in usage_units.",
		}, addLabelsFromStruct(reflect.TypeOf(openapi.ApiV2010UsageRecordToday{}))),
		callMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_calls",
			Help: "Number of calls made or received.",
		}, addLabelsFromStruct(reflect.TypeOf(openapi.ApiV2010Call{}))),
		messageMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "twilio_messages",
			Help: "Number of messages sent or received.",
		}, addLabelsFromStruct(reflect.TypeOf(openapi.ApiV2010Message{}))),
	}

	return tc
}

func (tc *TwilioCollector) Describe(ch chan<- *prometheus.Desc) {
	// We don't need to describe individual gauges, as they are self-describing.
}

func (tc *TwilioCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch account balance
	balance, err := tc.client.FetchBalance()
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch balance")
		return
	}
	tc.UpdateMetric(tc.balanceMetric, balance, "Balance")

	// Fetch and update metrics
	usageRecordsToday, err := tc.client.FetchUsageRecordsToday()
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch usage records")
		return
	}
	tc.UpdateMetric(tc.usageMetric, usageRecordsToday, "Usage")

	// Fetch and update call metrics
	callRecords, err := tc.client.FetchCalls()
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch call records")
		return
	}
	tc.UpdateMetric(tc.callMetric, callRecords, "")

	// Fetch and update message metrics
	messageRecords, err := tc.client.FetchMessages()
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch message records")
		return
	}
	tc.UpdateMetric(tc.messageMetric, messageRecords, "")

	// Export metrics to channel
	tc.balanceMetric.Collect(ch)
	tc.usageMetric.Collect(ch)
	tc.callMetric.Collect(ch)
	tc.messageMetric.Collect(ch)
}

func (tc *TwilioCollector) UpdateMetric(metric prometheus.Collector, records interface{}, metricField string) {
	logrus.WithFields(logrus.Fields{
		"metric":  metric,
		"records": records,
	}).Debug("Processed UpdateMetric")
	metric.(*prometheus.GaugeVec).Reset()
	v := reflect.ValueOf(records)

	if v.Kind() != reflect.Slice {
		logrus.Error("Invalid records type, expecting a slice")
		return
	}

	// Iterate over slice elements
	for i := 0; i < v.Len(); i++ {
		record := v.Index(i)

		// Ensure record is a struct
		if record.Kind() != reflect.Struct {
			logrus.Error("Invalid record type, expecting a struct")
			continue
		}

		var labelValues []string    // Initialize empty labelValues slice
		var metricValue float64 = 1 // Initialize empty metric value

		logrus.WithFields(logrus.Fields{
			"record": record,
		}).Debug("Processed record")

		// Iterate over struct fields
		for j := 0; j < record.NumField(); j++ {
			field := record.Type().Field(j)
			logrus.WithFields(logrus.Fields{
				"field": field,
			}).Debug("Processed field")

			value := record.Field(j)
			logrus.WithFields(logrus.Fields{
				"field value": value,
			}).Debug("Processed field value")

			fieldName := fieldNameToLabel(field.Name)
			logrus.WithFields(logrus.Fields{
				"fieldName": fieldName,
			}).Debug("Processed field name")

			fieldValue := ""

			// Convert pointers to strings and integers to values
			if value.Kind() == reflect.Ptr && !value.IsNil() {
				switch value.Elem().Kind() {
				case reflect.String:
					fieldValue = *value.Interface().(*string)
				case reflect.Int:
					fieldValue = strconv.Itoa(int(value.Elem().Int()))
				default:
					fieldValue = ""
				}
			}
			logrus.WithFields(logrus.Fields{
				"fieldName": fieldName,
			}).Debug("Processed field name")
			labelValues = append(labelValues, fieldValue)

			// Check if the current field is the metricField
			if field.Name == metricField {
				metricValue, _ = strconv.ParseFloat(fieldValue, 64)
			}
		}

		logrus.WithFields(logrus.Fields{
			"labelValues": labelValues,
			"metricValue": metricValue,
		}).Debug("Twilio metrics")
		// If skipMissing is true and metricValue is 0, skip the metric
		if tc.config.SkipMissing && metricValue == 0 {
			continue
		}

		if gaugeVec, ok := metric.(*prometheus.GaugeVec); ok {
			gaugeVec.WithLabelValues(labelValues...).Set(metricValue)
		} else {
			logrus.Error("Invalid metric type, expecting GaugeVec")
			return
		}
	}
}
