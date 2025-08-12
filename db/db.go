package db

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strings"
)

type DatabaseServer interface {
	Db() *sql.DB
	FetchTableDescr(ctx context.Context, name string) string
	FetchTableRows(ctx context.Context, name string) ([]string, []TableData)
	FetchSqlRows(ctx context.Context, SQL string) ([]string, []TableData)
	FetchDatabases(ctx context.Context) ([]string, []TableData)
	FetchTables(ctx context.Context) ([]string, []TableData)
	GetServerInfo(ctx context.Context) map[string]string
}

func Connect(connStr string, useMock bool) DatabaseServer {
	if useMock {
		slog.Info("Using mock database server")
		return &MysqlMock{Mysql{DbInstance: nil}}
	}

	driver := determineDriver(connStr)
	slog.Info("Database driver detected", "driver", driver)

	db, err := sql.Open(driver, connStr)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		slog.Error("Failed to ping database", "error", err)
	} else {
		slog.Info("Successfully connected to database")
	}

	return &Mysql8{Mysql{DbInstance: db}}
}

func determineDriver(connStr string) string {
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		return "pgx"
	}
	if strings.HasPrefix(connStr, "mysql://") || strings.Contains(connStr, "@tcp(") {
		return "mysql"
	}
	if strings.HasSuffix(connStr, ".db") || strings.HasPrefix(connStr, "file:") || strings.HasPrefix(connStr, ":memory:") {
		return "sqlite3"
	}
	return "pgx"
}
