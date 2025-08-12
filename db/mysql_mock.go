package db

import (
	"context"
	"fmt"
	"log/slog"
)

type MysqlMock struct {
	Mysql
}

func (m *MysqlMock) FetchTableDescr(ctx context.Context, name string) string {
	slog.Debug("fetchTableDescr: Getting mock table description", "tableName", name)

	mockDescriptions := map[string]string{
		"users":    "CREATE TABLE `users` (\n  `id` int(11) NOT NULL AUTO_INCREMENT,\n  `username` varchar(50) NOT NULL,\n  `email` varchar(100) NOT NULL,\n  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		"products": "CREATE TABLE `products` (\n  `id` int(11) NOT NULL AUTO_INCREMENT,\n  `name` varchar(100) NOT NULL,\n  `price` decimal(10,2) NOT NULL,\n  `category_id` int(11),\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		"orders":   "CREATE TABLE `orders` (\n  `id` int(11) NOT NULL AUTO_INCREMENT,\n  `user_id` int(11) NOT NULL,\n  `total` decimal(10,2) NOT NULL,\n  `order_date` timestamp DEFAULT CURRENT_TIMESTAMP,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
	}

	if desc, exists := mockDescriptions[name]; exists {
		return desc
	}

	return fmt.Sprintf("CREATE TABLE `%s` (\n  `id` int(11) NOT NULL AUTO_INCREMENT,\n  `data` varchar(255),\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", name)
}

func (m *MysqlMock) FetchTableRows(ctx context.Context, name string) ([]string, []TableData) {
	slog.Debug("fetchTableRows: Starting mock table data fetch", "tableName", name)

	headers := []string{"id", "name", "value", "created_at"}
	var tableData []TableData

	for i := 1; i <= 50; i++ {
		rowData := map[string]string{
			"id":         fmt.Sprintf("%d", i),
			"name":       fmt.Sprintf("Mock_%s_Row_%d", name, i),
			"value":      fmt.Sprintf("Sample data for %s row %d", name, i),
			"created_at": "2024-01-01 12:00:00",
		}
		tableData = append(tableData, rowData)
	}

	return headers, tableData
}

func (m *MysqlMock) FetchDatabases(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchDatabases: Starting mock database fetch")
	headers := []string{"NAME", "CHARSET", "COLLATION"}

	databases := []MysqlDatabase{
		{Name: "production_db", Charset: "utf8mb4", Collation: "utf8mb4_unicode_ci"},
		{Name: "staging_db", Charset: "utf8mb4", Collation: "utf8mb4_unicode_ci"},
		{Name: "test_db", Charset: "utf8", Collation: "utf8_general_ci"},
		{Name: "analytics_db", Charset: "utf8mb4", Collation: "utf8mb4_unicode_ci"},
		{Name: "backup_db", Charset: "utf8mb4", Collation: "utf8mb4_unicode_ci"},
	}

	var databaseData []TableData
	for _, item := range databases {
		databaseData = append(databaseData, item)
	}

	return headers, databaseData
}

func (m *MysqlMock) FetchTables(ctx context.Context) ([]string, []TableData) {
	slog.Debug("fetchTables: Starting mock database table fetch")
	headers := []string{"NAME", "TYPE", "ENGINE", "ROWS", "SIZE"}

	tableNames := []string{
		"users", "products", "orders", "categories", "inventory",
		"payments", "reviews", "addresses", "coupons", "wishlists",
		"cart_items", "shipping", "notifications", "logs", "sessions",
		"permissions", "roles", "settings", "configurations", "audit_trail",
		"analytics", "reports", "metrics", "feedback", "support_tickets",
		"faq", "blog_posts", "comments", "tags", "media_files",
		"email_templates", "job_queue", "cache_entries", "rate_limits", "api_keys",
		"webhooks", "integrations", "backups", "migrations", "health_checks",
		"user_preferences", "themes", "languages", "currencies", "tax_rates",
		"shipping_zones", "product_variants", "stock_movements", "price_history", "search_index",
	}

	var tableData []TableData
	for i, tableName := range tableNames {
		table := MysqlTable{
			Name:   tableName,
			Type:   "BASE TABLE",
			Engine: "InnoDB",
			Rows:   fmt.Sprintf("%d", (i+1)*1000+500),
			Size:   fmt.Sprintf("%.1fMB", float64(i+1)*2.5),
		}
		tableData = append(tableData, table)
	}

	return headers, tableData
}

// FetchSqlRows executes a mock SQL query and returns mock results
func (m *MysqlMock) FetchSqlRows(ctx context.Context, sqlQuery string) ([]string, []TableData) {
	slog.Debug("FetchSqlRows: Executing mock SQL query", "query", sqlQuery)

	// For mock, return some generic columns and data based on the query
	headers := []string{"id", "result", "query_executed"}
	var tableData []TableData

	// Generate some mock rows based on the query
	for i := 1; i <= 10; i++ {
		rowData := map[string]string{
			"id":             fmt.Sprintf("%d", i),
			"result":         fmt.Sprintf("Mock result row %d", i),
			"query_executed": sqlQuery,
		}
		tableData = append(tableData, rowData)
	}

	slog.Debug("FetchSqlRows: Mock processing complete", "query", sqlQuery, "rowsReturned", len(tableData))
	return headers, tableData
}

// FetchProcedures returns mock procedure data
func (m *MysqlMock) FetchProcedures(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchProcedures: Starting mock database procedure fetch")
	headers := []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"}

	procedures := []MysqlTable{
		{Name: "add_customer", Type: "PROCEDURE", Engine: "root@localhost", Rows: "2023-01-15 10:30:00 / 2024-03-22 14:45:00", Size: "DEFINER"},
		{Name: "update_inventory", Type: "PROCEDURE", Engine: "admin@localhost", Rows: "2023-02-22 08:15:00 / 2024-04-10 09:20:00", Size: "INVOKER"},
		{Name: "process_order", Type: "PROCEDURE", Engine: "root@localhost", Rows: "2023-03-10 13:45:00 / 2024-05-05 11:30:00", Size: "DEFINER"},
		{Name: "generate_report", Type: "PROCEDURE", Engine: "admin@localhost", Rows: "2023-04-18 09:00:00 / 2024-06-12 16:15:00", Size: "INVOKER"},
		{Name: "archive_data", Type: "PROCEDURE", Engine: "root@localhost", Rows: "2023-05-20 11:30:00 / 2024-07-01 10:10:00", Size: "DEFINER"},
		{Name: "calculate_metrics", Type: "PROCEDURE", Engine: "reports@localhost", Rows: "2023-06-15 14:20:00 / 2024-07-15 13:45:00", Size: "DEFINER"},
	}

	var procData []TableData
	for _, item := range procedures {
		procData = append(procData, item)
	}

	slog.Debug("FetchProcedures: Mock processing complete", "proceduresFound", len(procedures))
	return headers, procData
}

// FetchFunctions returns mock function data
func (m *MysqlMock) FetchFunctions(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchFunctions: Starting mock database function fetch")
	headers := []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "RETURN_TYPE", "IS_DETERMINISTIC"}

	functions := []MysqlTable{
		{Name: "calc_total_price", Type: "FUNCTION", Engine: "root@localhost", Rows: "2023-01-15 10:30:00 / 2024-03-22 14:45:00", Size: "DECIMAL (YES)"},
		{Name: "get_customer_tier", Type: "FUNCTION", Engine: "admin@localhost", Rows: "2023-02-22 08:15:00 / 2024-04-10 09:20:00", Size: "VARCHAR (NO)"},
		{Name: "apply_discount", Type: "FUNCTION", Engine: "root@localhost", Rows: "2023-03-10 13:45:00 / 2024-05-05 11:30:00", Size: "DECIMAL (YES)"},
		{Name: "format_address", Type: "FUNCTION", Engine: "admin@localhost", Rows: "2023-04-18 09:00:00 / 2024-06-12 16:15:00", Size: "VARCHAR (YES)"},
		{Name: "calculate_shipping", Type: "FUNCTION", Engine: "root@localhost", Rows: "2023-05-20 11:30:00 / 2024-07-01 10:10:00", Size: "DECIMAL (YES)"},
	}

	var fnData []TableData
	for _, item := range functions {
		fnData = append(fnData, item)
	}

	slog.Debug("FetchFunctions: Mock processing complete", "functionsFound", len(functions))
	return headers, fnData
}

// FetchTriggers returns mock trigger data
func (m *MysqlMock) FetchTriggers(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchTriggers: Starting mock database trigger fetch")
	headers := []string{"NAME", "EVENT", "TABLE", "TIMING", "DEFINER", "CREATED"}

	triggers := []MysqlTable{
		{Name: "after_order_insert", Type: "INSERT", Engine: "orders", Rows: "AFTER", Size: "root@localhost (2023-01-15 10:30:00)"},
		{Name: "before_product_update", Type: "UPDATE", Engine: "products", Rows: "BEFORE", Size: "admin@localhost (2023-02-22 08:15:00)"},
		{Name: "after_customer_delete", Type: "DELETE", Engine: "customers", Rows: "AFTER", Size: "root@localhost (2023-03-10 13:45:00)"},
		{Name: "after_payment_insert", Type: "INSERT", Engine: "payments", Rows: "AFTER", Size: "admin@localhost (2023-04-18 09:00:00)"},
		{Name: "before_inventory_update", Type: "UPDATE", Engine: "inventory", Rows: "BEFORE", Size: "root@localhost (2023-05-20 11:30:00)"},
	}

	var triggerData []TableData
	for _, item := range triggers {
		triggerData = append(triggerData, item)
	}

	slog.Debug("FetchTriggers: Mock processing complete", "triggersFound", len(triggers))
	return headers, triggerData
}

// FetchViews returns mock view data
func (m *MysqlMock) FetchViews(ctx context.Context) ([]string, []TableData) {
	slog.Debug("FetchViews: Starting mock database view fetch")
	headers := []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"}

	views := []MysqlTable{
		{Name: "active_customers", Type: "VIEW", Engine: "root@localhost", Rows: "2023-01-15 10:30:00 / 2024-03-22 14:45:00", Size: "DEFINER"},
		{Name: "product_inventory", Type: "VIEW", Engine: "admin@localhost", Rows: "2023-02-22 08:15:00 / 2024-04-10 09:20:00", Size: "INVOKER"},
		{Name: "order_summary", Type: "VIEW", Engine: "root@localhost", Rows: "2023-03-10 13:45:00 / 2024-05-05 11:30:00", Size: "DEFINER"},
		{Name: "sales_by_region", Type: "VIEW", Engine: "admin@localhost", Rows: "2023-04-18 09:00:00 / 2024-06-12 16:15:00", Size: "INVOKER"},
		{Name: "customer_orders", Type: "VIEW", Engine: "root@localhost", Rows: "2023-05-20 11:30:00 / 2024-07-01 10:10:00", Size: "DEFINER"},
		{Name: "monthly_revenue", Type: "VIEW", Engine: "reports@localhost", Rows: "2023-06-15 14:20:00 / 2024-07-15 13:45:00", Size: "DEFINER"},
		{Name: "top_products", Type: "VIEW", Engine: "analyst@localhost", Rows: "2023-07-22 08:45:00 / 2024-07-22 16:30:00", Size: "INVOKER"},
		{Name: "employee_performance", Type: "VIEW", Engine: "hr@localhost", Rows: "2023-08-05 12:15:00 / 2024-08-01 09:00:00", Size: "DEFINER"},
	}

	var viewData []TableData
	for _, item := range views {
		viewData = append(viewData, item)
	}

	slog.Debug("FetchViews: Mock processing complete", "viewsFound", len(views))
	return headers, viewData
}

// GetServerInfo returns mock MySQL server information
func (m *MysqlMock) GetServerInfo(ctx context.Context) map[string]string {
	slog.Debug("GetServerInfo: Returning mock MySQL server information")

	serverInfo := map[string]string{
		"version":                 "MySQL 8.0.30-mock",
		"user":                    "mock_user",
		"host":                    "localhost",
		"server_host":             "mock-server",
		"port":                    "3306",
		"database":                "mock_database",
		"max_connections":         "151",
		"innodb_buffer_pool_size": "128.0 MB",
	}

	return serverInfo
}
