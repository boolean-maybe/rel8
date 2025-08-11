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
