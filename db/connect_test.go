package db

import (
	"context"
	"testing"
	"time"
)

func TestConnectPostgreSQL(t *testing.T) {
	// Test that PostgreSQL connection strings return Postgres server
	postgresConnStr := "postgres://user:password@localhost:5432/testdb"

	server := Connect(postgresConnStr, true) // Use mock mode to avoid actual connection

	// Verify it's a Postgres server by checking the server info
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	serverInfo := server.GetServerInfo(ctx)

	// PostgreSQL should have shared_buffers, not innodb_buffer_pool_size
	if _, hasSharedBuffers := serverInfo["shared_buffers"]; !hasSharedBuffers {
		t.Error("Expected PostgreSQL server to have shared_buffers")
	}

	if _, hasInnodbBuffer := serverInfo["innodb_buffer_pool_size"]; hasInnodbBuffer {
		t.Error("PostgreSQL server should not have innodb_buffer_pool_size")
	}

	// Check expected PostgreSQL version format
	if version, exists := serverInfo["version"]; !exists {
		t.Error("Expected version to be present")
	} else if version != "PostgreSQL 15.4-mock" {
		t.Errorf("Expected PostgreSQL mock version, got: %s", version)
	}
}

func TestConnectMySQL(t *testing.T) {
	// Test that MySQL connection strings return MySQL server
	mysqlConnStr := "user:password@tcp(localhost:3306)/testdb"

	server := Connect(mysqlConnStr, true) // Use mock mode to avoid actual connection

	// Verify it's a MySQL server by checking the server info
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	serverInfo := server.GetServerInfo(ctx)

	// MySQL should NOT have shared_buffers (this comes from the mock, which doesn't have it)
	if _, hasSharedBuffers := serverInfo["shared_buffers"]; hasSharedBuffers {
		t.Error("MySQL server should not have shared_buffers")
	}
}
