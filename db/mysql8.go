package db

import (
	"context"
	"fmt"
	"log/slog"
)

type Mysql8 struct {
	Mysql
}

// fetchTableDescr queries table description by table name using SHOW CREATE TABLE
func (m *Mysql8) FetchTableDescr(ctx context.Context, name string) string {
	slog.Debug("fetchTableDescr: Getting table description", "tableName", name)

	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", name)
	slog.Debug("fetchTableDescr: Executing query", "query", query)

	row := m.Db().QueryRowContext(ctx, query)

	var tableName, createTable string
	err := row.Scan(&tableName, &createTable)
	if err != nil {
		slog.Error("fetchTableDescr: Failed to get table description", "error", err, "tableName", name)
		return ""
	}

	slog.Debug("fetchTableDescr: Successfully retrieved table description", "tableName", name)
	return createTable
}

// fetchTableRows queries table rows by table name
func (m *Mysql8) FetchTableRows(ctx context.Context, name string) ([]string, []TableData) {
	slog.Debug("fetchTableRows: Starting table data fetch", "tableName", name)

	// First, get column information to build headers
	columnQuery := `
		SELECT COLUMN_NAME 
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	slog.Debug("fetchTableRows: Getting column info", "query", columnQuery, "tableName", name)
	columnRows, err := m.Db().QueryContext(ctx, columnQuery, name)
	if err != nil {
		slog.Error("fetchTableRows: Column query failed", "error", err, "tableName", name)
		return []string{}, []TableData{}
	}
	defer columnRows.Close()

	var headers []string
	for columnRows.Next() {
		var columnName string
		if err := columnRows.Scan(&columnName); err != nil {
			slog.Warn("fetchTableRows: Failed to scan column name", "error", err)
			continue
		}
		headers = append(headers, columnName)
	}

	if len(headers) == 0 {
		slog.Error("fetchTableRows: No columns found", "tableName", name)
		return []string{}, []TableData{}
	}

	slog.Debug("fetchTableRows: Found columns", "count", len(headers), "headers", headers)

	// Query table data with LIMIT to avoid overwhelming memory
	dataQuery := fmt.Sprintf("SELECT * FROM `%s` LIMIT 1000", name)
	slog.Debug("fetchTableRows: Executing data query", "query", dataQuery)

	dataRows, err := m.Db().QueryContext(ctx, dataQuery)
	if err != nil {
		slog.Error("fetchTableRows: Data query failed", "error", err, "tableName", name)
		return headers, []TableData{}
	}
	defer dataRows.Close()

	var tableData []TableData
	rowCount := 0

	for dataRows.Next() {
		// Create slice to hold all column values for this row
		values := make([]interface{}, len(headers))
		valuePtrs := make([]interface{}, len(headers))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			slog.Warn("fetchTableRows: Failed to scan row", "error", err, "rowNum", rowCount)
			continue
		}

		// Convert values to strings and create a map-like structure
		rowData := make(map[string]string)
		for i, header := range headers {
			if values[i] != nil {
				// Handle byte slices (BLOB, TEXT, etc.) properly
				if byteData, ok := values[i].([]byte); ok {
					rowData[header] = string(byteData)
				} else {
					rowData[header] = fmt.Sprintf("%v", values[i])
				}
			} else {
				rowData[header] = "NULL"
			}
		}

		tableData = append(tableData, rowData)
		rowCount++
	}

	slog.Debug("fetchTableRows: Processing complete", "tableName", name, "rowsFound", rowCount)

	return headers, tableData
}

// fetchDatabases queries the database for database information
func (m *Mysql8) FetchDatabases(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchDatabases: Starting database fetch")
	headers := []string{"NAME", "CHARSET", "COLLATION"}
	databases := []MysqlDatabase{}

	// Query to get database information from information_schema
	query := `
		SELECT 
			SCHEMA_NAME,
			DEFAULT_CHARACTER_SET_NAME,
			DEFAULT_COLLATION_NAME
		FROM information_schema.SCHEMATA 
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		ORDER BY SCHEMA_NAME
	`

	slog.Debug("fetchDatabases: Executing query", "query", query)
	rows, err := m.Db().QueryContext(ctx, query)
	if err != nil {
		slog.Error("fetchDatabases: Query failed, returning mock data", "error", err)
		// If query fails, return mock data
		return []string{}, []TableData{}
	}
	defer rows.Close()

	slog.Debug("fetchDatabases: Query successful, processing rows")
	rowCount := 0
	for rows.Next() {
		var database MysqlDatabase

		err := rows.Scan(&database.Name, &database.Charset, &database.Collation)
		if err != nil {
			slog.Warn("fetchDatabases: Failed to scan row, skipping", "error", err)
			continue
		}

		slog.Debug("fetchDatabases: Processed database", "name", database.Name, "charset", database.Charset, "collation", database.Collation)
		databases = append(databases, database)
		rowCount++
	}

	slog.Debug("fetchDatabases: Processing complete", "databasesFound", rowCount)

	// If no databases found, return mock data
	if len(databases) == 0 {
		slog.Info("fetchDatabases: No databases found, returning mock data")
		return []string{}, []TableData{}
	}

	slog.Debug("fetchDatabases: Returning real data", "databaseCount", len(databases))
	var databaseData []TableData
	for _, item := range databases {
		databaseData = append(databaseData, item)
	}

	return headers, databaseData
}

// fetchTables queries the database for table information
func (m *Mysql8) FetchTables(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchTables: Starting database table fetch")
	headers := []string{"NAME", "TYPE", "ENGINE", "ROWS", "SIZE"}
	tables := []MysqlTable{}

	// Query to get table information from information_schema
	query := `
		SELECT 
			TABLE_NAME,
			TABLE_TYPE,
			IFNULL(ENGINE, 'N/A') as ENGINE,
			IFNULL(TABLE_ROWS, 0) as TABLE_ROWS,
			IFNULL(ROUND(((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024), 2), 0) as SIZE_MB
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME
	`

	slog.Debug("fetchTables: Executing query", "query", query)
	rows, err := m.Db().QueryContext(ctx, query)
	if err != nil {
		slog.Error("fetchTables: Query failed, returning mock data", "error", err)
		// If query fails, return mock data
		return []string{}, []TableData{}
	}
	defer rows.Close()

	slog.Debug("fetchTables: Query successful, processing rows")
	rowCount := 0
	for rows.Next() {
		var table MysqlTable
		var sizeFloat float64
		var tableRowCount int64

		err := rows.Scan(&table.Name, &table.Type, &table.Engine, &tableRowCount, &sizeFloat)
		if err != nil {
			slog.Warn("fetchTables: Failed to scan row, skipping", "error", err)
			continue
		}

		// Format the data nicely
		table.Rows = fmt.Sprintf("%d", tableRowCount)
		table.Size = fmt.Sprintf("%.1fMB", sizeFloat)

		slog.Debug("fetchTables: Processed table", "name", table.Name, "type", table.Type, "engine", table.Engine, "rows", table.Rows, "size", table.Size)
		tables = append(tables, table)
		rowCount++
	}

	slog.Debug("fetchTables: Processing complete", "tablesFound", rowCount)

	// If no tables found, return mock data
	if len(tables) == 0 {
		slog.Info("fetchTables: No tables found, returning mock data")
		return []string{}, []TableData{}
	}

	slog.Debug("fetchTables: Returning real data", "tableCount", len(tables))
	var tableData []TableData
	for _, item := range tables {
		tableData = append(tableData, item)
	}

	return headers, tableData
}

// FetchSqlRows executes a SQL query and returns the results similar to FetchTableRows
func (m *Mysql8) FetchSqlRows(ctx context.Context, sqlQuery string) ([]string, []TableData) {
	slog.Debug("FetchSqlRows: Executing SQL query", "query", sqlQuery)

	rows, err := m.Db().QueryContext(ctx, sqlQuery)
	if err != nil {
		slog.Error("FetchSqlRows: Query failed", "error", err, "query", sqlQuery)
		return []string{"Error"}, []TableData{map[string]string{"Error": err.Error()}}
	}
	defer rows.Close()

	// Get column names from the result set
	columnNames, err := rows.Columns()
	if err != nil {
		slog.Error("FetchSqlRows: Failed to get column names", "error", err)
		return []string{"Error"}, []TableData{map[string]string{"Error": err.Error()}}
	}

	slog.Debug("FetchSqlRows: Found columns", "count", len(columnNames), "headers", columnNames)

	var tableData []TableData
	rowCount := 0

	for rows.Next() {
		// Create slice to hold all column values for this row
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			slog.Warn("FetchSqlRows: Failed to scan row", "error", err, "rowNum", rowCount)
			continue
		}

		// Convert values to strings and create a map-like structure
		rowData := make(map[string]string)
		for i, columnName := range columnNames {
			if values[i] != nil {
				// Handle byte slices (BLOB, TEXT, etc.) properly
				if byteData, ok := values[i].([]byte); ok {
					rowData[columnName] = string(byteData)
				} else {
					rowData[columnName] = fmt.Sprintf("%v", values[i])
				}
			} else {
				rowData[columnName] = "NULL"
			}
		}

		tableData = append(tableData, rowData)
		rowCount++

		// Limit results to avoid overwhelming memory (same as FetchTableRows)
		if rowCount >= 1000 {
			slog.Debug("FetchSqlRows: Limiting results to 1000 rows")
			break
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("FetchSqlRows: Error during row iteration", "error", err)
		return []string{"Error"}, []TableData{map[string]string{"Error": err.Error()}}
	}

	slog.Debug("FetchSqlRows: Processing complete", "query", sqlQuery, "rowsFound", rowCount)
	return columnNames, tableData
}
