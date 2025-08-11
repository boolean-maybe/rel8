package db

import (
	"regexp"
	"strings"
)

// SQL keywords and their categories for syntax highlighting
var (
	sqlKeywords = []string{
		"SELECT", "FROM", "WHERE", "JOIN", "INNER", "LEFT", "RIGHT", "OUTER", "FULL",
		"ON", "AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS", "NULL",
		"ORDER", "BY", "GROUP", "HAVING", "LIMIT", "OFFSET", "DISTINCT", "ALL",
		"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE", "CREATE", "ALTER",
		"DROP", "TABLE", "INDEX", "VIEW", "DATABASE", "SCHEMA", "PRIMARY", "KEY",
		"FOREIGN", "REFERENCES", "CONSTRAINT", "UNIQUE", "CHECK", "DEFAULT",
		"AUTO_INCREMENT", "NOT", "NULL", "UNSIGNED", "ZEROFILL", "BINARY",
		"VARBINARY",
		"ENGINE", "CHARSET", "COLLATE", "COMMENT", "TEMPORARY", "IF",
		"BEGIN", "COMMIT", "ROLLBACK", "TRANSACTION", "START", "WORK",
		"SAVEPOINT", "RELEASE", "LOCK", "UNLOCK", "TABLES", "READ", "WRITE",
		"SHOW", "DESCRIBE", "DESC", "EXPLAIN", "ANALYZE", "USE", "CALL",
		"PROCEDURE", "FUNCTION", "TRIGGER", "EVENT", "GRANT", "REVOKE",
		"PRIVILEGES", "USER", "IDENTIFIED", "PASSWORD", "FLUSH", "RESET",
		"OPTIMIZE", "REPAIR", "BACKUP", "RESTORE", "DUMP", "LOAD", "DATA",
		"INFILE", "OUTFILE", "FIELDS", "TERMINATED", "ENCLOSED", "ESCAPED",
		"LINES", "STARTING", "IGNORE", "REPLACE", "DUPLICATE", "LOW_PRIORITY",
		"HIGH_PRIORITY", "DELAYED", "QUICK", "EXTENDED", "FULL", "MEDIUM",
		"PARTIAL", "FAST", "CHANGED", "USE_FRM", "FOR", "UPGRADE", "UNION",
	}

	sqlDataTypes = []string{
		"TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT",
		"DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL", "BIT", "BOOLEAN", "BOOL",
		"SERIAL", "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR",
		"CHAR", "VARCHAR", "BINARY", "VARBINARY", "TINYBLOB", "BLOB",
		"MEDIUMBLOB", "LONGBLOB", "TINYTEXT", "TEXT", "MEDIUMTEXT", "LONGTEXT",
		"ENUM", "SET", "GEOMETRY", "POINT", "LINESTRING", "POLYGON",
		"MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION",
		"JSON",
	}

	sqlFunctions = []string{
		"COUNT", "SUM", "AVG", "MIN", "MAX", "CONCAT", "LENGTH", "UPPER", "LOWER",
		"TRIM", "LTRIM", "RTRIM", "SUBSTRING", "SUBSTR", "LEFT", "RIGHT", "REPLACE",
		"REGEXP", "LIKE", "NOW", "CURDATE", "CURTIME", "DATE", "TIME", "YEAR",
		"MONTH", "DAY", "HOUR", "MINUTE", "SECOND", "DAYOFWEEK", "DAYOFYEAR",
		"WEEKDAY", "YEARWEEK", "DATE_FORMAT", "STR_TO_DATE", "UNIX_TIMESTAMP",
		"FROM_UNIXTIME", "COALESCE", "ISNULL", "NULLIF", "IF", "CASE", "WHEN",
		"THEN", "ELSE", "END", "CAST", "CONVERT", "FORMAT", "ROUND", "CEIL",
		"FLOOR", "ABS", "SIGN", "MOD", "POWER", "SQRT", "RAND", "PI",
	}
)

// HighlightSQL applies syntax highlighting to SQL text for tview TextView
// Uses tview color tags: [color]text[-] format
func HighlightSQL(sql string) string {
	if sql == "" {
		return sql
	}

	result := sql
	protectedAreas := make(map[string]string)
	placeholderIndex := 0

	// Function to create a placeholder for protected content
	createPlaceholder := func(content string) string {
		placeholder := "<<<PLACEHOLDER" + strings.Repeat("_", placeholderIndex) + ">>>"
		protectedAreas[placeholder] = content
		placeholderIndex++
		return placeholder
	}

	// Function to restore all placeholders
	restorePlaceholders := func(text string) string {
		for placeholder, original := range protectedAreas {
			text = strings.ReplaceAll(text, placeholder, original)
		}
		return text
	}

	// First, protect comments from other highlighting
	// Single-line comments starting with --
	singleCommentRe := regexp.MustCompile(`--.*`)
	result = singleCommentRe.ReplaceAllStringFunc(result, func(match string) string {
		highlighted := "[gray]" + match + "[-]"
		return createPlaceholder(highlighted)
	})

	// Multi-line comments /* ... */
	multiCommentRe := regexp.MustCompile(`(?s)/\*.*?\*/`)
	result = multiCommentRe.ReplaceAllStringFunc(result, func(match string) string {
		highlighted := "[gray]" + match + "[-]"
		return createPlaceholder(highlighted)
	})

	// Protect string literals from keyword highlighting
	// Single-quoted strings
	singleQuoteRe := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	result = singleQuoteRe.ReplaceAllStringFunc(result, func(match string) string {
		highlighted := "[red]" + match + "[-]"
		return createPlaceholder(highlighted)
	})

	// Double-quoted strings
	doubleQuoteRe := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	result = doubleQuoteRe.ReplaceAllStringFunc(result, func(match string) string {
		highlighted := "[red]" + match + "[-]"
		return createPlaceholder(highlighted)
	})

	// Protect backtick identifiers
	backtickRe := regexp.MustCompile("`([^`]+)`")
	result = backtickRe.ReplaceAllStringFunc(result, func(match string) string {
		highlighted := "[cyan]" + match + "[-]"
		return createPlaceholder(highlighted)
	})

	// Highlight functions in yellow
	for _, function := range sqlFunctions {
		// Match function calls (function name followed by opening parenthesis)
		pattern := `(?i)\b` + regexp.QuoteMeta(function) + `\s*\(`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// Extract just the function name part and highlight it
			funcPattern := `(?i)\b` + regexp.QuoteMeta(function)
			funcRe := regexp.MustCompile(funcPattern)
			return funcRe.ReplaceAllString(match, "[yellow]"+function+"[-]")
		})
	}

	// Highlight data types in light green
	for _, dataType := range sqlDataTypes {
		pattern := `(?i)\b` + regexp.QuoteMeta(dataType) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return "[lightgreen]" + match + "[-]"
		})
	}

	// Highlight SQL keywords in light blue
	for _, keyword := range sqlKeywords {
		// Use word boundaries to match whole words only, case-insensitive
		pattern := `(?i)\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return "[lightblue]" + match + "[-]"
		})
	}

	// Highlight numeric literals in magenta
	numberRe := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	result = numberRe.ReplaceAllStringFunc(result, func(match string) string {
		return "[magenta]" + match + "[-]"
	})

	// Restore all protected content
	result = restorePlaceholders(result)

	return result
}

// FormatSQL formats SQL with proper indentation and highlighting
// This is a simple formatter that adds basic indentation
func FormatSQL(sql string) string {
	if sql == "" {
		return sql
	}

	// Clean up the SQL first
	sql = strings.TrimSpace(sql)
	
	// Add basic formatting
	// This is a simple approach - for production use, consider a proper SQL formatter
	lines := strings.Split(sql, "\n")
	var formattedLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			formattedLines = append(formattedLines, line)
		}
	}
	
	formatted := strings.Join(formattedLines, "\n")
	
	// Apply syntax highlighting
	return HighlightSQL(formatted)
}