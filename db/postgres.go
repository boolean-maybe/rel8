package db

import (
	"context"
	"database/sql"
)

type Postgres struct {
	DbInstance *sql.DB
}

type PostgresTable struct {
	Name   string
	Type   string
	Engine string
	Rows   string
	Size   string
}

type PostgresDatabase struct {
	Name      string
	Charset   string
	Collation string
}

func (p *Postgres) Db() *sql.DB {
	return p.DbInstance
}

// GetServerInfo returns PostgreSQL mock server information for header display
func (p *Postgres) GetServerInfo(ctx context.Context) map[string]string {
	// Return mock data that resembles PostgreSQL server information
	serverInfo := make(map[string]string)

	serverInfo["version"] = "PostgreSQL 15.4"
	serverInfo["user"] = "postgres"
	serverInfo["database"] = "testdb"
	serverInfo["host"] = "localhost"
	serverInfo["port"] = "5432"
	serverInfo["max_connections"] = "100"
	serverInfo["shared_buffers"] = "128 MB"

	return serverInfo
}

// FetchTableDescr returns mock table description
func (p *Postgres) FetchTableDescr(ctx context.Context, name string) string {
	return "-- Mock PostgreSQL table description for " + name
}

// FetchTableRows returns mock table data
func (p *Postgres) FetchTableRows(ctx context.Context, name string) ([]string, []TableData) {
	headers := []string{"id", "name", "email", "created_at"}
	data := []TableData{
		map[string]string{"id": "1", "name": "John Doe", "email": "john@example.com", "created_at": "2023-01-01"},
		map[string]string{"id": "2", "name": "Jane Smith", "email": "jane@example.com", "created_at": "2023-01-02"},
	}
	return headers, data
}

// FetchSqlRows returns mock SQL query results
func (p *Postgres) FetchSqlRows(ctx context.Context, sqlQuery string) ([]string, []TableData) {
	headers := []string{"result"}
	data := []TableData{
		map[string]string{"result": "Mock PostgreSQL query result"},
	}
	return headers, data
}

// FetchDatabases returns mock database list
func (p *Postgres) FetchDatabases(ctx context.Context) ([]string, []TableData) {
	headers := []string{"NAME", "ENCODING", "COLLATE"}
	databases := []TableData{
		PostgresDatabase{Name: "postgres", Charset: "UTF8", Collation: "en_US.utf8"},
		PostgresDatabase{Name: "testdb", Charset: "UTF8", Collation: "en_US.utf8"},
		PostgresDatabase{Name: "myapp", Charset: "UTF8", Collation: "en_US.utf8"},
	}
	return headers, databases
}

// FetchTables returns mock table list
func (p *Postgres) FetchTables(ctx context.Context) ([]string, []TableData) {
	headers := []string{"NAME", "TYPE", "OWNER", "ROWS", "SIZE"}
	tables := []TableData{
		PostgresTable{Name: "users", Type: "table", Engine: "postgres", Rows: "1250", Size: "2.1MB"},
		PostgresTable{Name: "orders", Type: "table", Engine: "postgres", Rows: "5430", Size: "8.7MB"},
		PostgresTable{Name: "products", Type: "table", Engine: "postgres", Rows: "320", Size: "1.2MB"},
	}
	return headers, tables
}

// FetchViews returns mock view list
func (p *Postgres) FetchViews(ctx context.Context) ([]string, []TableData) {
	headers := []string{"NAME", "TYPE", "OWNER", "DEFINITION"}
	views := []TableData{
		PostgresTable{Name: "user_stats", Type: "view", Engine: "postgres", Rows: "SELECT COUNT(*)", Size: "Statistics view"},
		PostgresTable{Name: "order_summary", Type: "view", Engine: "postgres", Rows: "SELECT SUM(total)", Size: "Order aggregation"},
	}
	return headers, views
}

// FetchProcedures returns mock procedure list
func (p *Postgres) FetchProcedures(ctx context.Context) ([]string, []TableData) {
	headers := []string{"NAME", "TYPE", "OWNER", "LANGUAGE", "RETURNS"}
	procedures := []TableData{
		PostgresTable{Name: "calculate_tax", Type: "function", Engine: "postgres", Rows: "plpgsql", Size: "decimal"},
		PostgresTable{Name: "update_inventory", Type: "function", Engine: "postgres", Rows: "plpgsql", Size: "void"},
	}
	return headers, procedures
}

// FetchFunctions returns mock function list (PostgreSQL treats functions and procedures the same)
func (p *Postgres) FetchFunctions(ctx context.Context) ([]string, []TableData) {
	return p.FetchProcedures(ctx)
}

// FetchTriggers returns mock trigger list
func (p *Postgres) FetchTriggers(ctx context.Context) ([]string, []TableData) {
	headers := []string{"NAME", "EVENT", "TABLE", "TIMING", "FUNCTION"}
	triggers := []TableData{
		PostgresTable{Name: "update_modified_time", Type: "UPDATE", Engine: "users", Rows: "BEFORE", Size: "set_modified_time()"},
		PostgresTable{Name: "audit_changes", Type: "INSERT", Engine: "orders", Rows: "AFTER", Size: "log_audit()"},
	}
	return headers, triggers
}

// Database-specific methods for tree view
func (p *Postgres) FetchTablesForDatabase(ctx context.Context, databaseName string) ([]string, []TableData) {
	return p.FetchTables(ctx) // Return same mock data for any database
}

func (p *Postgres) FetchViewsForDatabase(ctx context.Context, databaseName string) ([]string, []TableData) {
	return p.FetchViews(ctx) // Return same mock data for any database
}

func (p *Postgres) FetchProceduresForDatabase(ctx context.Context, databaseName string) ([]string, []TableData) {
	return p.FetchProcedures(ctx) // Return same mock data for any database
}

func (p *Postgres) FetchFunctionsForDatabase(ctx context.Context, databaseName string) ([]string, []TableData) {
	return p.FetchFunctions(ctx) // Return same mock data for any database
}

func (p *Postgres) FetchTriggersForDatabase(ctx context.Context, databaseName string) ([]string, []TableData) {
	return p.FetchTriggers(ctx) // Return same mock data for any database
}
