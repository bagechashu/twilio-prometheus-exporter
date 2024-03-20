package main

import (
	"reflect"
	"strings"
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
