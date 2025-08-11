package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugConnectionString(t *testing.T) {
	tests := []struct {
		name        string
		connStr     string
		shouldParse bool
	}{
		{
			name:        "PostgreSQL URL",
			connStr:     "postgres://user:password@localhost:5432/mydb?sslmode=disable",
			shouldParse: true,
		},
		{
			name:        "MySQL URL",
			connStr:     "mysql://user:password@localhost:3306/mydb",
			shouldParse: true,
		},
		{
			name:        "SQLite file path",
			connStr:     "file:/path/to/database.db",
			shouldParse: true,
		},
		{
			name:        "Invalid URL",
			connStr:     "not://a:valid:url:format",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function primarily logs, so we just ensure it doesn't panic
			assert.NotPanics(t, func() {
				debugConnectionString(tt.connStr)
			})

			// Verify the URL can be parsed as expected
			u, err := url.Parse(tt.connStr)
			if tt.shouldParse {
				assert.NoError(t, err)
				assert.NotNil(t, u)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Note: Skipping configure function tests due to global state and flag redefinition issues
// The configure function is better tested through integration tests

// Environment variable testing also skipped due to global state issues
