package db

import (
	"context"
	"testing"
	"time"
)

func TestPostgresGetServerInfo(t *testing.T) {
	postgres := &Postgres{DbInstance: nil}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	serverInfo := postgres.GetServerInfo(ctx)

	// Verify that PostgreSQL mock data is returned
	expectedValues := map[string]string{
		"version":         "PostgreSQL 15.4",
		"user":            "postgres",
		"database":        "testdb",
		"host":            "localhost",
		"port":            "5432",
		"max_connections": "100",
		"shared_buffers":  "128 MB",
	}

	for key, expectedValue := range expectedValues {
		if actualValue, exists := serverInfo[key]; !exists {
			t.Errorf("Expected key %s not found in server info", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s to be %s, but got %s", key, expectedValue, actualValue)
		}
	}

	// Verify that MySQL-specific keys are not present
	mysqlSpecificKeys := []string{"innodb_buffer_pool_size"}
	for _, key := range mysqlSpecificKeys {
		if _, exists := serverInfo[key]; exists {
			t.Errorf("MySQL-specific key %s should not be present in PostgreSQL server info", key)
		}
	}
}
