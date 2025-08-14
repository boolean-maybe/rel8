package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFetchDatabases(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful database fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SCHEMA_NAME", "DEFAULT_CHARACTER_SET_NAME", "DEFAULT_COLLATION_NAME"}).
					AddRow("testdb", "utf8mb4", "utf8mb4_general_ci").
					AddRow("myapp", "utf8", "utf8_general_ci")
				mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "CHARSET", "COLLATION"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name: "no databases found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"SCHEMA_NAME", "DEFAULT_CHARACTER_SET_NAME", "DEFAULT_COLLATION_NAME"})
				mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchDatabases(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first database data structure
				if mysqlDB, ok := data[0].(MysqlDatabase); ok {
					assert.NotEmpty(t, mysqlDB.Name)
					assert.NotEmpty(t, mysqlDB.Charset)
					assert.NotEmpty(t, mysqlDB.Collation)
				} else {
					t.Error("Expected MysqlDatabase type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchTables(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful tables fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "ENGINE", "TABLE_ROWS", "SIZE_MB"}).
					AddRow("users", "BASE TABLE", "InnoDB", 100, 1.5).
					AddRow("orders", "BASE TABLE", "InnoDB", 500, 5.2)
				mock.ExpectQuery("SELECT TABLE_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "ENGINE", "ROWS", "SIZE"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT TABLE_NAME").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name: "no tables found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "ENGINE", "TABLE_ROWS", "SIZE_MB"})
				mock.ExpectQuery("SELECT TABLE_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchTables(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first table data structure
				if mysqlTable, ok := data[0].(MysqlTable); ok {
					assert.NotEmpty(t, mysqlTable.Name)
					assert.NotEmpty(t, mysqlTable.Type)
					assert.NotEmpty(t, mysqlTable.Engine)
					assert.NotEmpty(t, mysqlTable.Rows)
					assert.NotEmpty(t, mysqlTable.Size)
				} else {
					t.Error("Expected MysqlTable type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchTableRows(t *testing.T) {
	tests := []struct {
		name            string
		tableName       string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name:      "successful table rows fetch",
			tableName: "users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock column information query
				columnRows := sqlmock.NewRows([]string{"COLUMN_NAME"}).
					AddRow("id").
					AddRow("name").
					AddRow("email")
				mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.COLUMNS").
					WithArgs("users").
					WillReturnRows(columnRows)

				// Mock data query
				dataRows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john@example.com").
					AddRow(2, "Jane Smith", "jane@example.com")
				mock.ExpectQuery("SELECT \\* FROM `users` LIMIT 1000").
					WillReturnRows(dataRows)
			},
			expectedHeaders: []string{"id", "name", "email"},
			expectedCount:   2,
		},
		{
			name:      "column query fails",
			tableName: "users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.COLUMNS").
					WithArgs("users").
					WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name:      "no columns found",
			tableName: "users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				columnRows := sqlmock.NewRows([]string{"COLUMN_NAME"})
				mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.COLUMNS").
					WithArgs("users").
					WillReturnRows(columnRows)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name:      "data query fails but returns headers",
			tableName: "users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				columnRows := sqlmock.NewRows([]string{"COLUMN_NAME"}).
					AddRow("id").
					AddRow("name")
				mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.COLUMNS").
					WithArgs("users").
					WillReturnRows(columnRows)

				mock.ExpectQuery("SELECT \\* FROM `users` LIMIT 1000").
					WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{"id", "name"},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchTableRows(ctx, tt.tableName)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the data structure is map[string]string
				if rowData, ok := data[0].(map[string]string); ok {
					for _, header := range tt.expectedHeaders {
						_, exists := rowData[header]
						assert.True(t, exists, "Expected header %s to exist in row data", header)
					}
				} else {
					t.Error("Expected map[string]string type for row data")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchTableDescr(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		mockSetup      func(sqlmock.Sqlmock)
		expectedResult string
	}{
		{
			name:      "successful table description fetch",
			tableName: "users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				createTableSQL := `CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
				rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
					AddRow("users", createTableSQL)
				mock.ExpectQuery("SHOW CREATE TABLE `users`").
					WillReturnRows(rows)
			},
			expectedResult: `CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		},
		{
			name:      "simple table description",
			tableName: "orders",
			mockSetup: func(mock sqlmock.Sqlmock) {
				createTableSQL := `CREATE TABLE orders (id INT, user_id INT)`
				rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
					AddRow("orders", createTableSQL)
				mock.ExpectQuery("SHOW CREATE TABLE `orders`").
					WillReturnRows(rows)
			},
			expectedResult: `CREATE TABLE orders (id INT, user_id INT)`,
		},
		{
			name:      "query fails returns empty string",
			tableName: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SHOW CREATE TABLE `nonexistent`").
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: "",
		},
		{
			name:      "scan error returns empty string",
			tableName: "badtable",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
					AddRow("badtable", nil) // This will cause scan error
				mock.ExpectQuery("SHOW CREATE TABLE `badtable`").
					WillReturnRows(rows)
			},
			expectedResult: "",
		},
		{
			name:      "table with complex structure",
			tableName: "products",
			mockSetup: func(mock sqlmock.Sqlmock) {
				createTableSQL := `CREATE TABLE products (
  id INT AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  price DECIMAL(10,2) DEFAULT 0.00,
  description TEXT,
  category_id INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (category_id) REFERENCES categories(id),
  INDEX idx_category (category_id),
  INDEX idx_name (name)
)`
				rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
					AddRow("products", createTableSQL)
				mock.ExpectQuery("SHOW CREATE TABLE `products`").
					WillReturnRows(rows)
			},
			expectedResult: `CREATE TABLE products (
  id INT AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  price DECIMAL(10,2) DEFAULT 0.00,
  description TEXT,
  category_id INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (category_id) REFERENCES categories(id),
  INDEX idx_category (category_id),
  INDEX idx_name (name)
)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			result := mysql.FetchTableDescr(ctx, tt.tableName)

			assert.Equal(t, tt.expectedResult, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchSqlRows(t *testing.T) {
	tests := []struct {
		name            string
		sqlQuery        string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name:     "successful SQL query",
			sqlQuery: "SELECT id, name, email FROM users WHERE active = 1",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john@example.com").
					AddRow(2, "Jane Smith", "jane@example.com")
				mock.ExpectQuery("SELECT id, name, email FROM users WHERE active = 1").
					WillReturnRows(rows)
			},
			expectedHeaders: []string{"id", "name", "email"},
			expectedCount:   2,
		},
		{
			name:     "SQL query with joins",
			sqlQuery: "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "total"}).
					AddRow("John", "99.99").
					AddRow("Jane", "149.50")
				mock.ExpectQuery("SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id").
					WillReturnRows(rows)
			},
			expectedHeaders: []string{"name", "total"},
			expectedCount:   2,
		},
		{
			name:     "query returns error",
			sqlQuery: "SELECT * FROM nonexistent_table",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM nonexistent_table").
					WillReturnError(sql.ErrNoRows)
			},
			expectedHeaders: []string{"Error"},
			expectedCount:   1, // Should return one error row
		},
		{
			name:     "empty result set",
			sqlQuery: "SELECT id, name FROM users WHERE id = 999",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("SELECT id, name FROM users WHERE id = 999").
					WillReturnRows(rows)
			},
			expectedHeaders: []string{"id", "name"},
			expectedCount:   0,
		},
		{
			name:     "query with aggregation",
			sqlQuery: "SELECT COUNT(*) as count, AVG(price) as avg_price FROM products",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count", "avg_price"}).
					AddRow(150, 29.99)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) as count, AVG\\(price\\) as avg_price FROM products").
					WillReturnRows(rows)
			},
			expectedHeaders: []string{"count", "avg_price"},
			expectedCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchSqlRows(ctx, tt.sqlQuery)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the data structure is map[string]string
				if rowData, ok := data[0].(map[string]string); ok {
					for _, header := range tt.expectedHeaders {
						_, exists := rowData[header]
						assert.True(t, exists, "Expected header %s to exist in row data", header)
					}
				} else {
					t.Error("Expected map[string]string type for row data")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMockFetchSqlRows(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	testSQL := "SELECT * FROM users WHERE active = 1"
	headers, data := mock.FetchSqlRows(ctx, testSQL)

	// Verify headers
	assert.Len(t, headers, 3)
	assert.Equal(t, "id", headers[0])
	assert.Equal(t, "result", headers[1])
	assert.Equal(t, "query_executed", headers[2])

	// Verify data
	assert.Len(t, data, 10) // Mock returns 10 rows

	// Check first row
	firstRow := data[0].(map[string]string)
	assert.Equal(t, "1", firstRow["id"])
	assert.Equal(t, "Mock result row 1", firstRow["result"])
	assert.Equal(t, testSQL, firstRow["query_executed"])

	// Check last row
	lastRow := data[9].(map[string]string)
	assert.Equal(t, "10", lastRow["id"])
	assert.Equal(t, "Mock result row 10", lastRow["result"])
	assert.Equal(t, testSQL, lastRow["query_executed"])
}

func TestMockGetServerInfo(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	serverInfo := mock.GetServerInfo(ctx)

	// Verify expected keys are present and have values
	assert.Equal(t, "MySQL 8.0.30-mock", serverInfo["version"])
	assert.Equal(t, "mock_user", serverInfo["user"])
	assert.Equal(t, "localhost", serverInfo["host"])
	assert.Equal(t, "mock-server", serverInfo["server_host"])
	assert.Equal(t, "3306", serverInfo["port"])
	assert.Equal(t, "mock_database", serverInfo["database"])
	assert.Equal(t, "151", serverInfo["max_connections"])
	assert.Equal(t, "128.0 MB", serverInfo["innodb_buffer_pool_size"])
}

func TestFetchViews(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful views fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"}).
					AddRow("active_users", "VIEW", "root@localhost", "2023-01-15 10:30:00", "2024-03-22 14:45:00", "DEFINER").
					AddRow("sales_report", "VIEW", "admin@localhost", "2023-02-22 08:15:00", "2024-04-10 09:20:00", "INVOKER")
				mock.ExpectQuery("(?s)SELECT\\s+t\\.TABLE_NAME,\\s*'VIEW'\\s+as\\s+TABLE_TYPE,\\s*v\\.DEFINER,.*FROM\\s+information_schema\\.TABLES\\s+t\\s+JOIN\\s+information_schema\\.VIEWS").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("(?s)SELECT\\s+t\\.TABLE_NAME,\\s*'VIEW'\\s+as\\s+TABLE_TYPE,\\s*v\\.DEFINER,.*FROM\\s+information_schema\\.TABLES\\s+t\\s+JOIN\\s+information_schema\\.VIEWS").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name: "no views found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"})
				mock.ExpectQuery("(?s)SELECT\\s+t\\.TABLE_NAME,\\s*'VIEW'\\s+as\\s+TABLE_TYPE,\\s*v\\.DEFINER,.*FROM\\s+information_schema\\.TABLES\\s+t\\s+JOIN\\s+information_schema\\.VIEWS").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchViews(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first view data structure
				if mysqlTable, ok := data[0].(MysqlTable); ok {
					assert.NotEmpty(t, mysqlTable.Name)
					assert.NotEmpty(t, mysqlTable.Type)
					assert.NotEmpty(t, mysqlTable.Engine)
					assert.NotEmpty(t, mysqlTable.Rows)
					assert.NotEmpty(t, mysqlTable.Size)
				} else {
					t.Error("Expected MysqlTable type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMockFetchViews(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	headers, data := mock.FetchViews(ctx)

	// Verify headers
	assert.Equal(t, []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"}, headers)

	// Verify we have some views
	assert.True(t, len(data) > 0)

	// Check first view
	firstView := data[0].(MysqlTable)
	assert.Equal(t, "active_customers", firstView.Name)
	assert.Equal(t, "VIEW", firstView.Type)
}

func TestFetchProcedures(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful procedures fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"ROUTINE_NAME", "ROUTINE_TYPE", "DEFINER", "CREATED", "LAST_ALTERED", "SQL_MODE", "SECURITY_TYPE"}).
					AddRow("add_customer", "PROCEDURE", "root@localhost", "2023-01-15 10:30:00", "2024-03-22 14:45:00", "STRICT_TRANS_TABLES", "DEFINER").
					AddRow("update_inventory", "PROCEDURE", "admin@localhost", "2023-02-22 08:15:00", "2024-04-10 09:20:00", "STRICT_TRANS_TABLES", "INVOKER")
				mock.ExpectQuery("SELECT ROUTINE_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT ROUTINE_NAME").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
		{
			name: "no procedures found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"ROUTINE_NAME", "ROUTINE_TYPE", "DEFINER", "CREATED", "LAST_ALTERED", "SQL_MODE", "SECURITY_TYPE"})
				mock.ExpectQuery("SELECT ROUTINE_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchProcedures(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first procedure data structure
				if mysqlTable, ok := data[0].(MysqlTable); ok {
					assert.NotEmpty(t, mysqlTable.Name)
					assert.NotEmpty(t, mysqlTable.Type)
					assert.NotEmpty(t, mysqlTable.Engine)
					assert.NotEmpty(t, mysqlTable.Rows)
				} else {
					t.Error("Expected MysqlTable type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMockFetchProcedures(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	headers, data := mock.FetchProcedures(ctx)

	// Verify headers
	assert.Equal(t, []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"}, headers)

	// Verify we have some procedures
	assert.True(t, len(data) > 0)

	// Check first procedure
	firstProc := data[0].(MysqlTable)
	assert.Equal(t, "add_customer", firstProc.Name)
	assert.Equal(t, "PROCEDURE", firstProc.Type)
}

func TestFetchFunctions(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful functions fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"ROUTINE_NAME", "ROUTINE_TYPE", "DEFINER", "CREATED", "LAST_ALTERED", "DATA_TYPE", "IS_DETERMINISTIC"}).
					AddRow("calc_total", "FUNCTION", "root@localhost", "2023-01-15 10:30:00", "2024-03-22 14:45:00", "DECIMAL", "YES").
					AddRow("get_customer", "FUNCTION", "admin@localhost", "2023-02-22 08:15:00", "2024-04-10 09:20:00", "VARCHAR", "NO")
				mock.ExpectQuery("SELECT ROUTINE_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "RETURN_TYPE", "IS_DETERMINISTIC"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT ROUTINE_NAME").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchFunctions(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first function data structure
				if mysqlTable, ok := data[0].(MysqlTable); ok {
					assert.NotEmpty(t, mysqlTable.Name)
					assert.NotEmpty(t, mysqlTable.Type)
					assert.NotEmpty(t, mysqlTable.Engine)
					assert.NotEmpty(t, mysqlTable.Rows)
				} else {
					t.Error("Expected MysqlTable type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMockFetchFunctions(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	headers, data := mock.FetchFunctions(ctx)

	// Verify headers
	assert.Equal(t, []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "RETURN_TYPE", "IS_DETERMINISTIC"}, headers)

	// Verify we have some functions
	assert.True(t, len(data) > 0)

	// Check first function
	firstFunc := data[0].(MysqlTable)
	assert.Equal(t, "calc_total_price", firstFunc.Name)
	assert.Equal(t, "FUNCTION", firstFunc.Type)
}

func TestFetchTriggers(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(sqlmock.Sqlmock)
		expectedHeaders []string
		expectedCount   int
	}{
		{
			name: "successful triggers fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"TRIGGER_NAME", "EVENT_MANIPULATION", "EVENT_OBJECT_TABLE", "ACTION_TIMING", "DEFINER", "CREATED"}).
					AddRow("after_insert", "INSERT", "orders", "AFTER", "root@localhost", "2023-01-15 10:30:00").
					AddRow("before_update", "UPDATE", "products", "BEFORE", "admin@localhost", "2023-02-22 08:15:00")
				mock.ExpectQuery("SELECT TRIGGER_NAME").WillReturnRows(rows)
			},
			expectedHeaders: []string{"NAME", "EVENT", "TABLE", "TIMING", "DEFINER", "CREATED"},
			expectedCount:   2,
		},
		{
			name: "query error returns empty result",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT TRIGGER_NAME").WillReturnError(sql.ErrConnDone)
			},
			expectedHeaders: []string{},
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			headers, data := mysql.FetchTriggers(ctx)

			assert.Equal(t, tt.expectedHeaders, headers)
			assert.Len(t, data, tt.expectedCount)

			if tt.expectedCount > 0 {
				// Verify the first trigger data structure
				if mysqlTable, ok := data[0].(MysqlTable); ok {
					assert.NotEmpty(t, mysqlTable.Name)
					assert.NotEmpty(t, mysqlTable.Type)
					assert.NotEmpty(t, mysqlTable.Engine)
					assert.NotEmpty(t, mysqlTable.Rows)
				} else {
					t.Error("Expected MysqlTable type")
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestMockFetchTriggers(t *testing.T) {
	mock := &MysqlMock{}
	ctx := context.Background()

	headers, data := mock.FetchTriggers(ctx)

	// Verify headers
	assert.Equal(t, []string{"NAME", "EVENT", "TABLE", "TIMING", "DEFINER", "CREATED"}, headers)

	// Verify we have some triggers
	assert.True(t, len(data) > 0)

	// Check first trigger
	firstTrigger := data[0].(MysqlTable)
	assert.Equal(t, "after_order_insert", firstTrigger.Name)
	assert.Equal(t, "INSERT", firstTrigger.Type)
}

func TestGetServerInfo(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock)
		expectedValues map[string]string
	}{
		{
			name: "successful server info fetch",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock version query
				versionRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("version", "8.0.32")
				mock.ExpectQuery("SHOW VARIABLES LIKE 'version'").
					WillReturnRows(versionRows)

				// Mock connection info query
				connRows := sqlmock.NewRows([]string{"USER()", "@@hostname", "@@port", "DATABASE()"}).
					AddRow("testuser@localhost", "dbserver", "3306", "testdb")
				mock.ExpectQuery("SELECT USER\\(\\), @@hostname, @@port, DATABASE\\(\\)").
					WillReturnRows(connRows)

				// Mock max_connections query
				maxConnRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("max_connections", "500")
				mock.ExpectQuery("SHOW VARIABLES LIKE 'max_connections'").
					WillReturnRows(maxConnRows)

				// Mock innodb_buffer_pool_size query
				bufferPoolRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("innodb_buffer_pool_size", "134217728") // 128MB in bytes
				mock.ExpectQuery("SHOW VARIABLES LIKE 'innodb_buffer_pool_size'").
					WillReturnRows(bufferPoolRows)
			},
			expectedValues: map[string]string{
				"version":                 "MySQL 8.0.32",
				"user":                    "testuser",
				"host":                    "localhost",
				"server_host":             "dbserver",
				"port":                    "3306",
				"database":                "testdb",
				"max_connections":         "500",
				"innodb_buffer_pool_size": "128.0 MB",
			},
		},
		{
			name: "version query fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock version query with error
				mock.ExpectQuery("SHOW VARIABLES LIKE 'version'").
					WillReturnError(sql.ErrConnDone)

				// Mock connection info query
				connRows := sqlmock.NewRows([]string{"USER()", "@@hostname", "@@port", "DATABASE()"}).
					AddRow("testuser@localhost", "dbserver", "3306", "testdb")
				mock.ExpectQuery("SELECT USER\\(\\), @@hostname, @@port, DATABASE\\(\\)").
					WillReturnRows(connRows)

				// Mock max_connections query
				maxConnRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("max_connections", "500")
				mock.ExpectQuery("SHOW VARIABLES LIKE 'max_connections'").
					WillReturnRows(maxConnRows)

				// Mock innodb_buffer_pool_size query
				bufferPoolRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("innodb_buffer_pool_size", "134217728") // 128MB in bytes
				mock.ExpectQuery("SHOW VARIABLES LIKE 'innodb_buffer_pool_size'").
					WillReturnRows(bufferPoolRows)
			},
			expectedValues: map[string]string{
				"version": "Unknown",
			},
		},
		{
			name: "connection info query fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock version query
				versionRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("version", "8.0.32")
				mock.ExpectQuery("SHOW VARIABLES LIKE 'version'").
					WillReturnRows(versionRows)

				// Mock connection info query with error
				mock.ExpectQuery("SELECT USER\\(\\), @@hostname, @@port, DATABASE\\(\\)").
					WillReturnError(sql.ErrConnDone)

				// Mock max_connections query
				maxConnRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("max_connections", "500")
				mock.ExpectQuery("SHOW VARIABLES LIKE 'max_connections'").
					WillReturnRows(maxConnRows)

				// Mock innodb_buffer_pool_size query
				bufferPoolRows := sqlmock.NewRows([]string{"Variable_name", "Value"}).
					AddRow("innodb_buffer_pool_size", "134217728") // 128MB in bytes
				mock.ExpectQuery("SHOW VARIABLES LIKE 'innodb_buffer_pool_size'").
					WillReturnRows(bufferPoolRows)
			},
			expectedValues: map[string]string{
				"user":     "Unknown",
				"host":     "Unknown",
				"database": "Unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			tt.mockSetup(mock)

			mysql := &Mysql8{Mysql{DbInstance: mockDB}}
			ctx := context.Background()
			serverInfo := mysql.GetServerInfo(ctx)

			// Verify expected keys have the correct values
			for key, expectedValue := range tt.expectedValues {
				assert.Equal(t, expectedValue, serverInfo[key], "Value mismatch for key %s", key)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
