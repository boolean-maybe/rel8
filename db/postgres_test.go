package db

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresGetServerInfo(t *testing.T) {
	postgres := &Postgres{DbInstance: nil}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	serverInfo := postgres.GetServerInfo(ctx)

	// Verify that PostgreSQL mock data is returned (when no DB connection)
	expectedValues := map[string]string{
		"version":         "PostgreSQL 15.4-mock",
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

func TestPostgresGetServerInfoWithConnection(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	postgres := &Postgres{DbInstance: db}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Mock all the expected queries
	mock.ExpectQuery("SELECT version\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 15.4 on x86_64-pc-linux-gnu, compiled by gcc..."))

	mock.ExpectQuery("SELECT current_user").
		WillReturnRows(sqlmock.NewRows([]string{"current_user"}).
			AddRow("testuser"))

	mock.ExpectQuery("SELECT current_database\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).
			AddRow("testdb"))

	mock.ExpectQuery("SELECT inet_server_addr\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"inet_server_addr"}).
			AddRow("192.168.1.100"))

	mock.ExpectQuery("SHOW port").
		WillReturnRows(sqlmock.NewRows([]string{"port"}).
			AddRow("5432"))

	mock.ExpectQuery("SHOW max_connections").
		WillReturnRows(sqlmock.NewRows([]string{"max_connections"}).
			AddRow("200"))

	mock.ExpectQuery("SHOW shared_buffers").
		WillReturnRows(sqlmock.NewRows([]string{"shared_buffers"}).
			AddRow("256MB"))

	// Execute the function
	serverInfo := postgres.GetServerInfo(ctx)

	// Verify the results
	expectedValues := map[string]string{
		"version":         "PostgreSQL 15.4",
		"user":            "testuser",
		"database":        "testdb",
		"host":            "192.168.1.100",
		"port":            "5432",
		"max_connections": "200",
		"shared_buffers":  "256MB",
	}

	for key, expectedValue := range expectedValues {
		if actualValue, exists := serverInfo[key]; !exists {
			t.Errorf("Expected key %s not found in server info", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s to be %s, but got %s", key, expectedValue, actualValue)
		}
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestPostgresGetServerInfoUnixSocket(t *testing.T) {
	// Test the case where inet_server_addr() returns NULL (Unix socket connection)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	postgres := &Postgres{DbInstance: db}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Mock queries with NULL host (Unix socket)
	mock.ExpectQuery("SELECT version\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("PostgreSQL 15.4 on x86_64-pc-linux-gnu, compiled by gcc..."))

	mock.ExpectQuery("SELECT current_user").
		WillReturnRows(sqlmock.NewRows([]string{"current_user"}).
			AddRow("postgres"))

	mock.ExpectQuery("SELECT current_database\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).
			AddRow("mydb"))

	mock.ExpectQuery("SELECT inet_server_addr\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"inet_server_addr"}).
			AddRow(nil)) // NULL for Unix socket

	mock.ExpectQuery("SHOW port").
		WillReturnRows(sqlmock.NewRows([]string{"port"}).
			AddRow("5432"))

	mock.ExpectQuery("SHOW max_connections").
		WillReturnRows(sqlmock.NewRows([]string{"max_connections"}).
			AddRow("100"))

	mock.ExpectQuery("SHOW shared_buffers").
		WillReturnRows(sqlmock.NewRows([]string{"shared_buffers"}).
			AddRow("128MB"))

	// Execute the function
	serverInfo := postgres.GetServerInfo(ctx)

	// For Unix socket, host should default to "localhost"
	if host, exists := serverInfo["host"]; !exists {
		t.Error("Expected host key not found in server info")
	} else if host != "localhost" {
		t.Errorf("Expected host to be 'localhost' for Unix socket, but got %s", host)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
