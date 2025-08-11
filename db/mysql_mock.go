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
