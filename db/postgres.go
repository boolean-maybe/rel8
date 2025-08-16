package db

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
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

// GetServerInfo retrieves PostgreSQL server information using SQL queries
func (p *Postgres) GetServerInfo(ctx context.Context) map[string]string {
	slog.Debug("GetServerInfo: Retrieving PostgreSQL server information")
	serverInfo := make(map[string]string)

	if p.DbInstance == nil {
		// Return mock data when no database connection
		slog.Debug("GetServerInfo: No database connection, returning mock data")
		serverInfo["version"] = "PostgreSQL 15.4-mock"
		serverInfo["user"] = "postgres"
		serverInfo["database"] = "testdb"
		serverInfo["host"] = "localhost"
		serverInfo["port"] = "5432"
		serverInfo["max_connections"] = "100"
		serverInfo["shared_buffers"] = "128 MB"
		return serverInfo
	}

	// Get server version
	versionQuery := "SELECT version()"
	versionRows, err := p.DbInstance.QueryContext(ctx, versionQuery)
	if err == nil {
		defer versionRows.Close()
		if versionRows.Next() {
			var version string
			if err := versionRows.Scan(&version); err == nil {
				// Extract PostgreSQL version from full version string
				// Example: "PostgreSQL 15.4 on x86_64-pc-linux-gnu, compiled by gcc..."
				if strings.HasPrefix(version, "PostgreSQL ") {
					parts := strings.Split(version, " ")
					if len(parts) >= 2 {
						serverInfo["version"] = "PostgreSQL " + parts[1]
					} else {
						serverInfo["version"] = version
					}
				} else {
					serverInfo["version"] = version
				}
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get version", "error", err)
		serverInfo["version"] = "Unknown"
	}

	// Get current user
	userQuery := "SELECT current_user"
	userRows, err := p.DbInstance.QueryContext(ctx, userQuery)
	if err == nil {
		defer userRows.Close()
		if userRows.Next() {
			var user string
			if err := userRows.Scan(&user); err == nil {
				serverInfo["user"] = user
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get current user", "error", err)
		serverInfo["user"] = "Unknown"
	}

	// Get current database
	dbQuery := "SELECT current_database()"
	dbRows, err := p.DbInstance.QueryContext(ctx, dbQuery)
	if err == nil {
		defer dbRows.Close()
		if dbRows.Next() {
			var database string
			if err := dbRows.Scan(&database); err == nil {
				serverInfo["database"] = database
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get current database", "error", err)
		serverInfo["database"] = "Unknown"
	}

	// Get server host (may be NULL for Unix domain socket connections)
	hostQuery := "SELECT inet_server_addr()"
	hostRows, err := p.DbInstance.QueryContext(ctx, hostQuery)
	if err == nil {
		defer hostRows.Close()
		if hostRows.Next() {
			var host sql.NullString
			if err := hostRows.Scan(&host); err == nil {
				if host.Valid {
					serverInfo["host"] = host.String
				} else {
					serverInfo["host"] = "localhost" // Unix domain socket
				}
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get server host", "error", err)
		serverInfo["host"] = "Unknown"
	}

	// Get server port
	portQuery := "SHOW port"
	portRows, err := p.DbInstance.QueryContext(ctx, portQuery)
	if err == nil {
		defer portRows.Close()
		if portRows.Next() {
			var port string
			if err := portRows.Scan(&port); err == nil {
				serverInfo["port"] = port
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get server port", "error", err)
		serverInfo["port"] = "Unknown"
	}

	// Get max_connections
	maxConnQuery := "SHOW max_connections"
	maxConnRows, err := p.DbInstance.QueryContext(ctx, maxConnQuery)
	if err == nil {
		defer maxConnRows.Close()
		if maxConnRows.Next() {
			var maxConn string
			if err := maxConnRows.Scan(&maxConn); err == nil {
				serverInfo["max_connections"] = maxConn
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get max_connections", "error", err)
		serverInfo["max_connections"] = "Unknown"
	}

	// Get shared_buffers (PostgreSQL equivalent of MySQL's innodb_buffer_pool_size)
	sharedBuffersQuery := "SHOW shared_buffers"
	sharedBuffersRows, err := p.DbInstance.QueryContext(ctx, sharedBuffersQuery)
	if err == nil {
		defer sharedBuffersRows.Close()
		if sharedBuffersRows.Next() {
			var sharedBuffers string
			if err := sharedBuffersRows.Scan(&sharedBuffers); err == nil {
				serverInfo["shared_buffers"] = sharedBuffers
			}
		}
	} else {
		slog.Error("GetServerInfo: Failed to get shared_buffers", "error", err)
		serverInfo["shared_buffers"] = "Unknown"
	}

	slog.Debug("GetServerInfo: Retrieved PostgreSQL server information", "info", serverInfo)
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
