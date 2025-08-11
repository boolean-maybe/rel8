package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFields(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name: "struct with string fields",
			input: struct {
				Name string
				Age  string
			}{
				Name: "John",
				Age:  "30",
			},
			expected: []string{"John", "30"},
		},
		{
			name: "empty struct",
			input: struct {
			}{},
			expected: nil,
		},
		{
			name:     "non-struct type",
			input:    "not a struct",
			expected: nil,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFieldsWithHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		headers  []string
		expected []string
	}{
		{
			name: "map with all headers present",
			input: map[string]string{
				"name": "John",
				"age":  "30",
				"city": "NYC",
			},
			headers:  []string{"name", "age", "city"},
			expected: []string{"John", "30", "NYC"},
		},
		{
			name: "map with missing headers",
			input: map[string]string{
				"name": "John",
				"city": "NYC",
			},
			headers:  []string{"name", "age", "city"},
			expected: []string{"John", "", "NYC"},
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			headers:  []string{"name", "age"},
			expected: []string{"", ""},
		},
		{
			name: "map with extra fields not in headers",
			input: map[string]string{
				"name":    "John",
				"age":     "30",
				"ignored": "value",
			},
			headers:  []string{"name", "age"},
			expected: []string{"John", "30"},
		},
		{
			name: "struct fallback",
			input: struct {
				Name string
				Age  string
			}{
				Name: "John",
				Age:  "30",
			},
			headers:  []string{"NAME", "AGE"},
			expected: []string{"John", "30"},
		},
		{
			name:     "nil input",
			input:    nil,
			headers:  []string{"name", "age"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldsWithHeaders(tt.input, tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}
