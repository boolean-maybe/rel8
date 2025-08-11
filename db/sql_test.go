package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHighlightSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Expected color tags to be present
	}{
		{
			name:  "simple SELECT query",
			input: "SELECT id, name FROM users WHERE age > 25",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[lightblue]WHERE[-]",
				"[magenta]25[-]",
			},
		},
		{
			name:  "query with string literals",
			input: "SELECT * FROM users WHERE name = 'John Doe'",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[lightblue]WHERE[-]",
				"[red]'John Doe'[-]",
			},
		},
		{
			name:  "query with functions",
			input: "SELECT COUNT(*), MAX(age) FROM users",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[yellow]COUNT[-]",
				"[yellow]MAX[-]",
			},
		},
		{
			name:  "CREATE TABLE with data types",
			input: "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255), created_at TIMESTAMP)",
			expected: []string{
				"[lightblue]CREATE[-]",
				"[lightblue]TABLE[-]",
				"[lightblue]PRIMARY[-]",
				"[lightblue]KEY[-]",
				"[lightgreen]INT[-]",
				"[lightgreen]VARCHAR[-]",
				"[lightgreen]TIMESTAMP[-]",
				"[magenta]255[-]",
			},
		},
		{
			name:  "query with backtick identifiers",
			input: "SELECT `user_id`, `full_name` FROM `user_table`",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[cyan]`user_id`[-]",
				"[cyan]`full_name`[-]",
				"[cyan]`user_table`[-]",
			},
		},
		{
			name:  "query with comments",
			input: "SELECT * FROM users -- This is a comment\n/* Multi-line\n   comment */",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[gray]-- This is a comment[-]",
				"[gray]/* Multi-line\n   comment */[-]",
			},
		},
		{
			name:  "case insensitive keywords",
			input: "select id from users where name like '%john%'",
			expected: []string{
				"[lightblue]select[-]",
				"[lightblue]from[-]",
				"[lightblue]where[-]",
				"[lightblue]like[-]",
				"[red]'%john%'[-]",
			},
		},
		{
			name:  "complex query with JOIN",
			input: "SELECT u.name, p.title FROM users u LEFT JOIN posts p ON u.id = p.user_id",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[lightblue]LEFT[-]",
				"[lightblue]JOIN[-]",
				"[lightblue]ON[-]",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:  "numbers and decimals",
			input: "SELECT price FROM products WHERE price > 19.99 AND quantity = 10",
			expected: []string{
				"[lightblue]SELECT[-]",
				"[lightblue]FROM[-]",
				"[lightblue]WHERE[-]",
				"[lightblue]AND[-]",
				"[magenta]19.99[-]",
				"[magenta]10[-]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightSQL(tt.input)

			// Check that all expected color tags are present
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected color tag not found in result")
			}

			// For empty input, result should be empty
			if tt.input == "" {
				assert.Equal(t, "", result)
			}
		})
	}
}

func TestFormatSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple query",
			input:    "SELECT * FROM users",
			expected: "[lightblue]SELECT[-] * [lightblue]FROM[-] users",
		},
		{
			name:     "query with extra whitespace",
			input:    "   SELECT   *   FROM   users   ",
			expected: "[lightblue]SELECT[-]   *   [lightblue]FROM[-]   users",
		},
		{
			name:     "multiline query",
			input:    "SELECT *\nFROM users\nWHERE age > 25",
			expected: "[lightblue]SELECT[-] *\n[lightblue]FROM[-] users\n[lightblue]WHERE[-] age > [magenta]25[-]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSQL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHighlightSQLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, result string)
	}{
		{
			name:  "keywords within identifiers should not be highlighted",
			input: "SELECT user_select, from_date FROM user_table",
			validate: func(t *testing.T, result string) {
				// SELECT should be highlighted, but user_select should not have SELECT highlighted
				assert.Contains(t, result, "[lightblue]SELECT[-]")
				assert.Contains(t, result, "[lightblue]FROM[-]")
				// user_select should appear as-is (no partial highlighting)
				assert.Contains(t, result, "user_select")
				assert.Contains(t, result, "from_date")
			},
		},
		{
			name:  "escaped quotes in strings",
			input: "SELECT 'It\\'s a test' FROM users",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "[lightblue]SELECT[-]")
				assert.Contains(t, result, "[lightblue]FROM[-]")
				assert.Contains(t, result, "[red]'It\\'s a test'[-]")
			},
		},
		{
			name:  "functions without parentheses should not be highlighted as functions",
			input: "SELECT COUNT, MAX FROM function_names",
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "[lightblue]SELECT[-]")
				assert.Contains(t, result, "[lightblue]FROM[-]")
				// COUNT and MAX without parentheses should not be highlighted as functions
				assert.NotContains(t, result, "[yellow]COUNT[-]")
				assert.NotContains(t, result, "[yellow]MAX[-]")
			},
		},
		{
			name:  "nested quotes",
			input: `SELECT "John's \"nickname\"" FROM users`,
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "[lightblue]SELECT[-]")
				assert.Contains(t, result, "[lightblue]FROM[-]")
				assert.Contains(t, result, `[red]"John's \"nickname\""[-]`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightSQL(tt.input)
			tt.validate(t, result)
		})
	}
}

func TestSQLKeywordsCoverage(t *testing.T) {
	// Test that we have good coverage of common SQL keywords
	commonKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER",
		"WHERE", "ORDER", "GROUP", "HAVING", "JOIN", "UNION", "INDEX",
	}

	for _, keyword := range commonKeywords {
		found := false
		for _, sqlKeyword := range sqlKeywords {
			if strings.EqualFold(keyword, sqlKeyword) {
				found = true
				break
			}
		}
		assert.True(t, found, "Common SQL keyword '%s' not found in sqlKeywords list", keyword)
	}
}