package main

import (
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// fieldNameToLabel converts a field name to a label format by replacing spaces with underscores and converting to lowercase.
func fieldNameToLabel(fieldName string) string {
	return strings.ReplaceAll(strings.ToLower(fieldName), " ", "_")
}

// addLabelsFromStruct extracts field names from a struct type and converts them to labels.
func addLabelsFromStruct(recordType reflect.Type) []string {
	var labels []string
	for i := 0; i < recordType.NumField(); i++ {
		label := fieldNameToLabel(recordType.Field(i).Name)
		labels = append(labels, label)
	}
	return labels
}

func RemoveMetricWithLabels(metric prometheus.Collector, labels prometheus.Labels) {
	if gaugeVec, ok := metric.(*prometheus.GaugeVec); ok {
		if gaugeVec.Delete(labels) {
			logrus.Infof("Removed metric with labels: %+v", labels)
		} else {
			logrus.Warnf("Metric with labels: %+v not found", labels)
		}
	} else {
		logrus.Error("Invalid metric type, expecting GaugeVec")
	}
}
