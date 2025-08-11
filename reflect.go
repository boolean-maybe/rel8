package main

import "reflect"

// extract fields from unknown structure
func getFields(item interface{}) []string {
	var fields []string
	v := reflect.ValueOf(item)

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fields = append(fields, v.Field(i).String())
		}
	}
	return fields
}

// extract fields from map using headers to maintain order
func getFieldsWithHeaders(item interface{}, headers []string) []string {
	var fields []string

	// Try map first
	if mapData, ok := item.(map[string]string); ok {
		for _, header := range headers {
			if value, exists := mapData[header]; exists {
				fields = append(fields, value)
			} else {
				fields = append(fields, "")
			}
		}
		return fields
	}

	// Fall back to struct handling
	return getFields(item)
}
