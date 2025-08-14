package db

import (
	"context"
	"fmt"
	"log/slog"
)

type PostgresMock struct {
	Postgres
}

func (p *PostgresMock) FetchTableDescr(ctx context.Context, name string) string {
	slog.Debug("fetchTableDescr: Getting mock PostgreSQL table description", "tableName", name)

	mockDescriptions := map[string]string{
		"users":    "CREATE TABLE users (\n  id serial PRIMARY KEY,\n  username varchar(50) NOT NULL,\n  email varchar(100) NOT NULL,\n  created_at timestamp DEFAULT now()\n);",
		"products": "CREATE TABLE products (\n  id serial PRIMARY KEY,\n  name varchar(100) NOT NULL,\n  price decimal(10,2) NOT NULL,\n  category_id integer\n);",
		"orders":   "CREATE TABLE orders (\n  id serial PRIMARY KEY,\n  user_id integer NOT NULL,\n  total decimal(10,2) NOT NULL,\n  order_date timestamp DEFAULT now()\n);",
	}

	if desc, exists := mockDescriptions[name]; exists {
		return desc
	}

	return fmt.Sprintf("CREATE TABLE %s (\n  id serial PRIMARY KEY,\n  data varchar(255)\n);", name)
}

func (p *PostgresMock) FetchTableRows(ctx context.Context, name string) ([]string, []TableData) {
	slog.Debug("fetchTableRows: Starting mock PostgreSQL table data fetch", "tableName", name)

	headers := []string{"id", "name", "value", "created_at"}
	var tableData []TableData

	for i := 1; i <= 50; i++ {
		rowData := map[string]string{
			"id":         fmt.Sprintf("%d", i),
			"name":       fmt.Sprintf("Mock_%s_Row_%d", name, i),
			"value":      fmt.Sprintf("Sample PostgreSQL data for %s row %d", name, i),
			"created_at": "2024-01-01 12:00:00",
		}
		tableData = append(tableData, rowData)
	}

	return headers, tableData
}

func (p *PostgresMock) FetchDatabases(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchDatabases: Starting mock PostgreSQL database fetch")
	headers := []string{"NAME", "ENCODING", "COLLATE"}

	databases := []PostgresDatabase{
		{Name: "postgres", Charset: "UTF8", Collation: "en_US.utf8"},
		{Name: "testdb", Charset: "UTF8", Collation: "en_US.utf8"},
		{Name: "myapp", Charset: "UTF8", Collation: "en_US.utf8"},
		{Name: "analytics", Charset: "UTF8", Collation: "en_US.utf8"},
		{Name: "staging", Charset: "UTF8", Collation: "en_US.utf8"},
	}

	var databaseData []TableData
	for _, item := range databases {
		databaseData = append(databaseData, item)
	}

	return headers, databaseData
}

func (p *PostgresMock) FetchTables(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchTables: Starting mock PostgreSQL table fetch")
	headers := []string{"NAME", "TYPE", "OWNER", "ROWS", "SIZE"}

	tableNames := []string{
		"users", "products", "orders", "categories", "inventory",
		"payments", "reviews", "addresses", "coupons", "wishlists",
		"cart_items", "shipping", "notifications", "logs", "sessions",
		"permissions", "roles", "settings", "configurations", "audit_trail",
		"analytics", "reports", "metrics", "feedback", "support_tickets",
	}

	var tableData []TableData
	for i, tableName := range tableNames {
		table := PostgresTable{
			Name:   tableName,
			Type:   "table",
			Engine: "postgres",
			Rows:   fmt.Sprintf("%d", (i+1)*1000+500),
			Size:   fmt.Sprintf("%.1fMB", float64(i+1)*2.5),
		}
		tableData = append(tableData, table)
	}

	return headers, tableData
}

// FetchSqlRows executes a mock PostgreSQL SQL query and returns mock results
func (p *PostgresMock) FetchSqlRows(ctx context.Context, sqlQuery string) ([]string, []TableData) {
	slog.Debug("FetchSqlRows: Executing mock PostgreSQL SQL query", "query", sqlQuery)

	headers := []string{"id", "result", "query_executed"}
	var tableData []TableData

	for i := 1; i <= 10; i++ {
		rowData := map[string]string{
			"id":             fmt.Sprintf("%d", i),
			"result":         fmt.Sprintf("Mock PostgreSQL result row %d", i),
			"query_executed": sqlQuery,
		}
		tableData = append(tableData, rowData)
	}

	slog.Debug("FetchSqlRows: Mock PostgreSQL processing complete", "query", sqlQuery, "rowsReturned", len(tableData))
	return headers, tableData
}

// FetchProcedures returns mock PostgreSQL function data (PostgreSQL calls them functions)
func (p *PostgresMock) FetchProcedures(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchProcedures: Starting mock PostgreSQL function fetch")
	headers := []string{"NAME", "TYPE", "OWNER", "LANGUAGE", "RETURNS"}

	procedures := []PostgresTable{
		{Name: "calculate_tax", Type: "function", Engine: "postgres", Rows: "plpgsql", Size: "decimal"},
		{Name: "update_inventory", Type: "function", Engine: "postgres", Rows: "plpgsql", Size: "void"},
		{Name: "process_payment", Type: "function", Engine: "admin", Rows: "plpgsql", Size: "boolean"},
		{Name: "generate_report", Type: "function", Engine: "postgres", Rows: "sql", Size: "setof record"},
	}

	var procData []TableData
	for _, item := range procedures {
		procData = append(procData, item)
	}

	slog.Debug("FetchProcedures: Mock PostgreSQL processing complete", "functionsFound", len(procedures))
	return headers, procData
}

// FetchFunctions returns same as procedures for PostgreSQL
func (p *PostgresMock) FetchFunctions(ctx context.Context) ([]string, []TableData) {
	return p.FetchProcedures(ctx)
}

// FetchTriggers returns mock PostgreSQL trigger data
func (p *PostgresMock) FetchTriggers(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchTriggers: Starting mock PostgreSQL trigger fetch")
	headers := []string{"NAME", "EVENT", "TABLE", "TIMING", "FUNCTION"}

	triggers := []PostgresTable{
		{Name: "update_modified_time", Type: "UPDATE", Engine: "users", Rows: "BEFORE", Size: "set_modified_time()"},
		{Name: "audit_changes", Type: "INSERT", Engine: "orders", Rows: "AFTER", Size: "log_audit()"},
		{Name: "validate_email", Type: "INSERT", Engine: "users", Rows: "BEFORE", Size: "validate_email_format()"},
		{Name: "sync_inventory", Type: "UPDATE", Engine: "products", Rows: "AFTER", Size: "sync_inventory_count()"},
	}

	var triggerData []TableData
	for _, item := range triggers {
		triggerData = append(triggerData, item)
	}

	slog.Debug("FetchTriggers: Mock PostgreSQL processing complete", "triggersFound", len(triggers))
	return headers, triggerData
}

// FetchViews returns mock PostgreSQL view data
func (p *PostgresMock) FetchViews(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchViews: Starting mock PostgreSQL view fetch")
	headers := []string{"NAME", "TYPE", "OWNER", "DEFINITION"}

	views := []PostgresTable{
		{Name: "active_users", Type: "view", Engine: "postgres", Rows: "SELECT * FROM users WHERE active", Size: "User activity view"},
		{Name: "order_summary", Type: "view", Engine: "postgres", Rows: "SELECT COUNT(*), SUM(total)", Size: "Order aggregation"},
		{Name: "product_stats", Type: "view", Engine: "admin", Rows: "SELECT category, COUNT(*)", Size: "Product statistics"},
		{Name: "monthly_sales", Type: "view", Engine: "reports", Rows: "SELECT date_trunc('month')", Size: "Monthly sales data"},
	}

	var viewData []TableData
	for _, item := range views {
		viewData = append(viewData, item)
	}

	slog.Debug("FetchViews: Mock PostgreSQL processing complete", "viewsFound", len(views))
	return headers, viewData
}

// GetServerInfo returns mock PostgreSQL server information for header display
func (p *PostgresMock) GetServerInfo(ctx context.Context) map[string]string {
	slog.Debug("GetServerInfo: Returning mock PostgreSQL server information")

	serverInfo := map[string]string{
		"version":         "PostgreSQL 15.4-mock",
		"user":            "postgres",
		"host":            "localhost",
		"server_host":     "mock-pg-server",
		"port":            "5432",
		"database":        "testdb",
		"max_connections": "100",
		"shared_buffers":  "128 MB",
	}

	return serverInfo
}
