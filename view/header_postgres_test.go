package view

import (
	"rel8/db"
	"strings"
	"testing"
)

func TestHeaderWithPostgreSQL(t *testing.T) {
	// Create a PostgreSQL mock server
	server := &db.PostgresMock{db.Postgres{DbInstance: nil}}

	// Create a header and set the server
	header := NewHeader()
	header.SetServer(server)

	// Get the header text
	headerText := header.serverInfoHeader.GetText(false)

	// Check that PostgreSQL-specific information is displayed
	if !strings.Contains(headerText, "PostgreSQL 15.4-mock") {
		t.Error("Expected PostgreSQL version in header")
	}

	if !strings.Contains(headerText, "postgres") {
		t.Error("Expected postgres user in header")
	}

	if !strings.Contains(headerText, "testdb") {
		t.Error("Expected testdb database in header")
	}

	if !strings.Contains(headerText, "5432") {
		t.Error("Expected port 5432 in header")
	}

	if !strings.Contains(headerText, "Shared Buffers") {
		t.Error("Expected 'Shared Buffers' label for PostgreSQL")
	}

	if !strings.Contains(headerText, "128 MB") {
		t.Error("Expected shared buffers value in header")
	}

	// Make sure MySQL-specific terms are not present
	if strings.Contains(headerText, "Buffer Pool") {
		t.Error("Should not contain MySQL 'Buffer Pool' label for PostgreSQL")
	}
}

func TestHeaderWithMySQL(t *testing.T) {
	// Create a MySQL mock server
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}

	// Create a header and set the server
	header := NewHeader()
	header.SetServer(server)

	// Get the header text
	headerText := header.serverInfoHeader.GetText(false)

	// Check that MySQL-specific information is displayed
	if !strings.Contains(headerText, "MySQL 8.0.30-mock") {
		t.Error("Expected MySQL version in header")
	}

	if !strings.Contains(headerText, "Buffer Pool") {
		t.Error("Expected 'Buffer Pool' label for MySQL")
	}

	// Make sure PostgreSQL-specific terms are not present
	if strings.Contains(headerText, "Shared Buffers") {
		t.Error("Should not contain PostgreSQL 'Shared Buffers' label for MySQL")
	}
}
