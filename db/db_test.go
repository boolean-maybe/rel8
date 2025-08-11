package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineDriver(t *testing.T) {
	tests := []struct {
		name           string
		connStr        string
		expectedDriver string
	}{
		{
			name:           "PostgreSQL with postgres:// prefix",
			connStr:        "postgres://user:pass@localhost:5432/dbname",
			expectedDriver: "pgx",
		},
		{
			name:           "PostgreSQL with postgresql:// prefix",
			connStr:        "postgresql://user:pass@localhost:5432/dbname",
			expectedDriver: "pgx",
		},
		{
			name:           "MySQL with mysql:// prefix",
			connStr:        "mysql://user:pass@localhost:3306/dbname",
			expectedDriver: "mysql",
		},
		{
			name:           "MySQL with @tcp( pattern",
			connStr:        "user:pass@tcp(localhost:3306)/dbname",
			expectedDriver: "mysql",
		},
		{
			name:           "SQLite with .db suffix",
			connStr:        "/path/to/database.db",
			expectedDriver: "sqlite3",
		},
		{
			name:           "SQLite with file: prefix",
			connStr:        "file:/path/to/database.db",
			expectedDriver: "sqlite3",
		},
		{
			name:           "SQLite in-memory",
			connStr:        ":memory:",
			expectedDriver: "sqlite3",
		},
		{
			name:           "Unknown connection string defaults to pgx",
			connStr:        "unknown://connection/string",
			expectedDriver: "pgx",
		},
		{
			name:           "Empty connection string defaults to pgx",
			connStr:        "",
			expectedDriver: "pgx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineDriver(tt.connStr)
			assert.Equal(t, tt.expectedDriver, result)
		})
	}
}
