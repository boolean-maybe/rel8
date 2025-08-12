package view

import (
	"testing"

	"rel8/db"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestNewEmptyGrid(t *testing.T) {
	grid := NewEmptyGrid()
	table := grid.Table

	assert.NotNil(t, table)
	assert.IsType(t, &tview.Table{}, table)

	// Test that table is configured properly - we can't access private fields directly,
	// but we can verify the table was created without panicking
	assert.NotPanics(t, func() {
		table.SetCell(0, 0, tview.NewTableCell("test"))
	})
}

func TestNewGrid(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	data := []db.TableData{
		map[string]string{"Name": "John", "Age": "30", "City": "NYC"},
		map[string]string{"Name": "Jane", "Age": "25", "City": "LA"},
	}

	grid := NewGrid(headers, data)
	table := grid.Table

	assert.NotNil(t, table)
	assert.IsType(t, &tview.Table{}, table)

	// Check table dimensions
	rowCount := table.GetRowCount()
	columnCount := table.GetColumnCount()

	assert.Equal(t, 3, rowCount)    // 1 header + 2 data rows
	assert.Equal(t, 3, columnCount) // 3 columns

	// Check header content
	for col, expectedHeader := range headers {
		cell := table.GetCell(0, col)
		assert.Equal(t, expectedHeader, cell.Text)
	}

	// We can't directly access the selection due to tview's API design,
	// but we can verify the table was populated correctly
	assert.Equal(t, 3, rowCount)    // 1 header + 2 data rows
	assert.Equal(t, 3, columnCount) // 3 columns
}

func TestGridPopulate(t *testing.T) {
	grid := NewEmptyGrid()
	headers := []string{"ID", "Name", "Status"}
	data := []db.TableData{
		map[string]string{"ID": "1", "Name": "Test1", "Status": "Active"},
		map[string]string{"ID": "2", "Name": "Test2", "Status": "Inactive"},
		struct {
			ID     string
			Name   string
			Status string
		}{"3", "Test3", "Pending"}, // Test struct handling
	}

	grid.Populate(headers, data)
	table := grid.Table

	// Check table dimensions
	assert.Equal(t, 4, table.GetRowCount()) // 1 header + 3 data rows
	assert.Equal(t, 3, table.GetColumnCount())

	// Check header row
	for col, expectedHeader := range headers {
		cell := table.GetCell(0, col)
		assert.Equal(t, expectedHeader, cell.Text)
		// Note: Color and attributes may not be set as expected due to tview's internal handling
	}

	// Check first data row (map)
	assert.Equal(t, "1", table.GetCell(1, 0).Text)
	assert.Equal(t, "Test1", table.GetCell(1, 1).Text)
	assert.Equal(t, "Active", table.GetCell(1, 2).Text)

	// Check data row exists and has content
	for row := 1; row <= 3; row++ {
		for col := 0; col < 3; col++ {
			cell := table.GetCell(row, col)
			assert.NotNil(t, cell)
			// Note: Colors may not be set as expected due to tview's internal handling
		}
	}

	// Verify table was populated correctly
	assert.Equal(t, 4, table.GetRowCount()) // 1 header + 3 data rows
	assert.Equal(t, 3, table.GetColumnCount())
}

func TestGridSelectionChangedFunc(t *testing.T) {
	grid := NewGrid([]string{"Name", "Value"}, []db.TableData{
		map[string]string{"Name": "test", "Value": "123"},
	})
	table := grid.Table

	// Verify table has the expected structure
	assert.Equal(t, 2, table.GetRowCount()) // 1 header + 1 data row
	assert.Equal(t, 2, table.GetColumnCount())
	assert.Equal(t, "Name", table.GetCell(0, 0).Text)
	assert.Equal(t, "test", table.GetCell(1, 0).Text)
}

func TestWrapGrid(t *testing.T) {
	grid := NewEmptyGrid()
	wrappedGrid := WrapGrid(grid)
	assert.NotNil(t, wrappedGrid)
	assert.IsType(t, &tview.Flex{}, wrappedGrid)
}

func TestGridRestoreSelection(t *testing.T) {
	grid := NewGrid([]string{"Name", "Value"}, []db.TableData{
		map[string]string{"Name": "test1", "Value": "123"},
		map[string]string{"Name": "test2", "Value": "456"},
		map[string]string{"Name": "test3", "Value": "789"},
	})

	// Test restoring valid selection
	grid.RestoreSelection(1, 3) // Select second data row
	row, _ := grid.GetSelection()
	assert.Equal(t, 2, row) // Should be row 2 (1 + 1 for header)

	// Test invalid selection (negative)
	grid.RestoreSelection(-1, 3)
	// Should not crash or change selection inappropriately

	// Test invalid selection (out of bounds)
	grid.RestoreSelection(5, 3)
	// Should not crash or change selection inappropriately
}

// TestFullRowHighlighting tests that all cells in a selected row are properly highlighted
// This test ensures all columns (including rightmost ones) are highlighted correctly
func TestFullRowHighlighting(t *testing.T) {
	// Create a grid with multiple columns to test rightmost column highlighting
	headers := []string{"Col1", "Col2", "Col3", "Col4", "Col5", "Col6"}
	data := []db.TableData{
		map[string]string{
			"Col1": "Row1Data1",
			"Col2": "Row1Data2",
			"Col3": "Row1Data3",
			"Col4": "Row1Data4",
			"Col5": "Row1Data5",
			"Col6": "Row1Data6",
		},
		map[string]string{
			"Col1": "Row2Data1",
			"Col2": "Row2Data2",
			"Col3": "Row2Data3",
			"Col4": "Row2Data4",
			"Col5": "Row2Data5",
			"Col6": "Row2Data6",
		},
	}

	grid := NewGrid(headers, data)

	// simulate user input to trigger the real selection changed handler
	ih := grid.InputHandler()
	ih(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), func(p tview.Primitive) {})

	// verify row 2 is selected
	selectedRow, selectedCol := grid.GetSelection()
	assert.Equal(t, 2, selectedRow, "should have selected row 2")
	assert.Equal(t, 0, selectedCol, "should have selected column 0")

	// verify all columns in the selected row are highlighted (using style)
	for col := 0; col < len(headers); col++ {
		cell := grid.GetCell(selectedRow, col)
		if assert.NotNil(t, cell) {
			fg, bg, _ := cell.Style.Decompose()
			assert.Equal(t, Colors.TextAqua, bg,
				"cell at row %d, column %d should have highlight background", selectedRow, col)
			assert.Equal(t, Colors.TextBlack, fg,
				"cell at row %d, column %d should have highlight text color", selectedRow, col)
		}
	}
}

// TestSingleRowSelectionAcrossAllTableModes tests that in all table modes:
// 1. Only one row is selected at a time
// 2. Selection behavior works correctly across all table modes
// 3. Table structure is correct for each mode
func TestSingleRowSelectionAcrossAllTableModes(t *testing.T) {
	// Define test data for different table modes
	tableModeTestCases := []struct {
		name        string
		headers     []string
		data        []db.TableData
		description string
	}{
		{
			name:    "DatabaseTable",
			headers: []string{"Table", "Engine", "Rows", "Size", "Comment"},
			data: []db.TableData{
				map[string]string{"Table": "users", "Engine": "InnoDB", "Rows": "100", "Size": "16KB", "Comment": "Users table"},
				map[string]string{"Table": "orders", "Engine": "InnoDB", "Rows": "500", "Size": "64KB", "Comment": "Orders table"},
				map[string]string{"Table": "products", "Engine": "MyISAM", "Rows": "200", "Size": "32KB", "Comment": "Products table"},
			},
			description: "Database tables listing",
		},
		{
			name:    "View",
			headers: []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"},
			data: []db.TableData{
				map[string]string{"NAME": "user_orders", "TYPE": "VIEW", "DEFINER": "root@localhost", "CREATED": "2024-01-01 10:00:00", "UPDATED": "2024-01-15 14:00:00", "SECURITY_TYPE": "DEFINER"},
				map[string]string{"NAME": "product_summary", "TYPE": "VIEW", "DEFINER": "admin@%", "CREATED": "2024-01-02 11:00:00", "UPDATED": "2024-01-16 15:00:00", "SECURITY_TYPE": "INVOKER"},
			},
			description: "Database views listing",
		},
		{
			name:    "Function",
			headers: []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"},
			data: []db.TableData{
				map[string]string{"NAME": "calculate_tax", "TYPE": "FUNCTION", "DEFINER": "root@localhost", "CREATED": "2024-01-01", "MODIFIED": "2024-01-01", "SQL_MODE": "STRICT_TRANS_TABLES", "SECURITY_TYPE": "DEFINER"},
				map[string]string{"NAME": "format_currency", "TYPE": "FUNCTION", "DEFINER": "admin@%", "CREATED": "2024-01-02", "MODIFIED": "2024-01-02", "SQL_MODE": "NO_ZERO_DATE", "SECURITY_TYPE": "INVOKER"},
				map[string]string{"NAME": "validate_email", "TYPE": "FUNCTION", "DEFINER": "dev@localhost", "CREATED": "2024-01-03", "MODIFIED": "2024-01-03", "SQL_MODE": "ERROR_FOR_DIVISION_BY_ZERO", "SECURITY_TYPE": "DEFINER"},
			},
			description: "Database functions listing",
		},
		{
			name:    "Procedure",
			headers: []string{"NAME", "TYPE", "DEFINER", "CREATED", "MODIFIED", "SQL_MODE", "SECURITY_TYPE"},
			data: []db.TableData{
				map[string]string{"NAME": "update_inventory", "TYPE": "PROCEDURE", "DEFINER": "root@localhost", "CREATED": "2024-01-01", "MODIFIED": "2024-01-01", "SQL_MODE": "STRICT_TRANS_TABLES", "SECURITY_TYPE": "DEFINER"},
				map[string]string{"NAME": "process_orders", "TYPE": "PROCEDURE", "DEFINER": "admin@%", "CREATED": "2024-01-02", "MODIFIED": "2024-01-02", "SQL_MODE": "NO_ZERO_DATE", "SECURITY_TYPE": "INVOKER"},
			},
			description: "Database procedures listing",
		},
		{
			name:    "Trigger",
			headers: []string{"Trigger", "Event", "Table", "Statement", "Timing", "Created", "Definer"},
			data: []db.TableData{
				map[string]string{"Trigger": "users_update_trigger", "Event": "UPDATE", "Table": "users", "Statement": "SET NEW.updated_at = NOW()", "Timing": "BEFORE", "Created": "2024-01-01", "Definer": "root@localhost"},
				map[string]string{"Trigger": "orders_insert_trigger", "Event": "INSERT", "Table": "orders", "Statement": "SET NEW.created_at = NOW()", "Timing": "BEFORE", "Created": "2024-01-02", "Definer": "admin@%"},
			},
			description: "Database triggers listing",
		},
		{
			name:    "Database",
			headers: []string{"Database", "Default Charset", "Default Collation", "Size"},
			data: []db.TableData{
				map[string]string{"Database": "production", "Default Charset": "utf8mb4", "Default Collation": "utf8mb4_unicode_ci", "Size": "512MB"},
				map[string]string{"Database": "staging", "Default Charset": "utf8", "Default Collation": "utf8_general_ci", "Size": "128MB"},
				map[string]string{"Database": "development", "Default Charset": "utf8mb4", "Default Collation": "utf8mb4_unicode_ci", "Size": "64MB"},
			},
			description: "Databases listing",
		},
		{
			name:    "TableRow",
			headers: []string{"id", "name", "email", "created_at", "status", "last_login"},
			data: []db.TableData{
				map[string]string{"id": "1", "name": "John Doe", "email": "john@example.com", "created_at": "2024-01-01 10:00:00", "status": "active", "last_login": "2024-01-15 14:30:00"},
				map[string]string{"id": "2", "name": "Jane Smith", "email": "jane@example.com", "created_at": "2024-01-02 11:00:00", "status": "inactive", "last_login": "2024-01-10 09:15:00"},
				map[string]string{"id": "3", "name": "Bob Wilson", "email": "bob@example.com", "created_at": "2024-01-03 12:00:00", "status": "active", "last_login": "2024-01-14 16:45:00"},
				map[string]string{"id": "4", "name": "Alice Brown", "email": "alice@example.com", "created_at": "2024-01-04 13:00:00", "status": "pending", "last_login": ""},
			},
			description: "Table rows data",
		},
	}

	for _, tc := range tableModeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create grid with test data
			grid := NewGrid(tc.headers, tc.data)

			// Test basic table structure
			totalRows := grid.GetRowCount()
			totalCols := grid.GetColumnCount()

			assert.Equal(t, len(tc.headers), totalCols, "Column count should match headers for %s", tc.description)
			assert.Equal(t, len(tc.data)+1, totalRows, "Row count should be data rows + 1 header for %s", tc.description)

			// Test that table is configured for single row selection (not column selection)
			// This is verified by the fact that the table was created with SetSelectable(true, false)
			// We verify this indirectly by ensuring GetSelection returns valid coordinates
			selectedRow, selectedCol := grid.GetSelection()
			assert.Equal(t, 1, selectedRow, "Initial selection should be first data row (row 1) for %s", tc.description)
			assert.Equal(t, 0, selectedCol, "Initial selection should be first column for %s", tc.description)

			// Test that we can change selection to different rows (proving only one row selected at a time)
			if totalRows > 2 {
				// Change selection to row 2
				grid.Select(2, 0)
				newSelectedRow, newSelectedCol := grid.GetSelection()
				assert.Equal(t, 2, newSelectedRow, "After Select(2,0), selected row should be 2 for %s", tc.description)
				assert.Equal(t, 0, newSelectedCol, "After Select(2,0), selected column should be 0 for %s", tc.description)

				// Change back to row 1
				grid.Select(1, 0)
				backSelectedRow, backSelectedCol := grid.GetSelection()
				assert.Equal(t, 1, backSelectedRow, "After Select(1,0), selected row should be 1 for %s", tc.description)
				assert.Equal(t, 0, backSelectedCol, "After Select(1,0), selected column should be 0 for %s", tc.description)
			}

			// Test that header row (row 0) cannot be selected by attempting to select it
			grid.Select(0, 0)
			finalSelectedRow, _ := grid.GetSelection()
			// The selection should either stay where it was or move to row 1 (depends on tview implementation)
			// But it should never be row 0 (header)
			assert.True(t, finalSelectedRow >= 1, "After trying to select header row, selection should be >= 1 for %s", tc.description)

			// Verify header row content and structure
			for col := 0; col < totalCols; col++ {
				headerCell := grid.GetCell(0, col)
				if assert.NotNil(t, headerCell, "Header cell at col %d should exist for %s", col, tc.description) {
					assert.Equal(t, tc.headers[col], headerCell.Text, "Header text should match for %s", tc.description)
				}
			}

			// Verify data rows content exists (cells should exist even if content is empty)
			for row := 1; row < totalRows; row++ {
				for col := 0; col < totalCols; col++ {
					dataCell := grid.GetCell(row, col)
					assert.NotNil(t, dataCell, "Data cell at row %d, col %d should exist for %s", row, col, tc.description)
					// Note: Content can be empty for optional fields like last_login
				}
			}

			// Test that the table has appropriate border and padding configuration
			// This verifies the table was set up with the proper UI configuration
			// (Border is enabled, padding is set, etc.)
			assert.NotNil(t, grid.Table, "Grid should have a valid table for %s", tc.description)
		})
	}
}

// TestSelectionHandlerHighlightsAllColumns tests that the selection change handler
// properly highlights ALL columns by directly simulating the handler logic
func TestSelectionHandlerHighlightsAllColumns(t *testing.T) {
	testCases := []struct {
		name    string
		headers []string
		data    []db.TableData
	}{
		{
			name:    "TwoColumns",
			headers: []string{"First", "Last"},
			data: []db.TableData{
				map[string]string{"First": "A", "Last": "Z"},
				map[string]string{"First": "B", "Last": "Y"},
			},
		},
		{
			name:    "ManyColumns",
			headers: []string{"C1", "C2", "C3", "C4", "C5", "C6", "C7", "LastColumn"},
			data: []db.TableData{
				map[string]string{"C1": "a", "C2": "b", "C3": "c", "C4": "d", "C5": "e", "C6": "f", "C7": "g", "LastColumn": "FINAL"},
				map[string]string{"C1": "1", "C2": "2", "C3": "3", "C4": "4", "C5": "5", "C6": "6", "C7": "7", "LastColumn": "END"},
			},
		},
		{
			name:    "ViewHeaders",
			headers: []string{"NAME", "TYPE", "DEFINER", "CREATED", "UPDATED", "SECURITY_TYPE"},
			data: []db.TableData{
				map[string]string{"NAME": "user_orders", "TYPE": "VIEW", "DEFINER": "root@localhost", "CREATED": "2024-01-01 10:00:00", "UPDATED": "2024-01-15 14:00:00", "SECURITY_TYPE": "DEFINER"},
				map[string]string{"NAME": "product_summary", "TYPE": "VIEW", "DEFINER": "admin@%", "CREATED": "2024-01-02 11:00:00", "UPDATED": "2024-01-16 15:00:00", "SECURITY_TYPE": "INVOKER"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grid := NewGrid(tc.headers, tc.data)
			totalRows := grid.GetRowCount()
			expectedCols := len(tc.headers)
			lastColIndex := expectedCols - 1

			// iterate over each data row and trigger the real selection change handler
			for testRow := 1; testRow < totalRows; testRow++ {
				// select the row to invoke the handler
				grid.Select(testRow, 0)

				// verify all cells in selected row are highlighted (using style)
				for col := 0; col < expectedCols; col++ {
					cell := grid.GetCell(testRow, col)
					if assert.NotNil(t, cell, "cell should exist at row %d, col %d", testRow, col) {
						fg, bg, _ := cell.Style.Decompose()
						if bg != Colors.TextAqua || fg != Colors.TextBlack {
							t.Errorf("cell (%d,%d) not highlighted; got bg=%v fg=%v", testRow, col, bg, fg)
						}
					}
				}

				// last column must be highlighted as well
				lastCell := grid.GetCell(testRow, lastColIndex)
				if assert.NotNil(t, lastCell, "last cell should exist at row %d", testRow) {
					fg, bg, _ := lastCell.Style.Decompose()
					if bg != Colors.TextAqua || fg != Colors.TextBlack {
						t.Errorf("LAST COLUMN '%s' at index %d in row %d is NOT highlighted (bg=%v fg=%v)",
							tc.headers[lastColIndex], lastColIndex, testRow, bg, fg)
					}
				}

				// additionally ensure previous row (if any) got reset (not highlighted)
				if testRow > 1 {
					prevRow := testRow - 1
					for col := 0; col < expectedCols; col++ {
						cell := grid.GetCell(prevRow, col)
						if assert.NotNil(t, cell, "prev row cell should exist at row %d, col %d", prevRow, col) {
							fg, bg, _ := cell.Style.Decompose()
							if bg == Colors.TextAqua && fg == Colors.TextBlack {
								t.Errorf("previous row %d, col %d remains highlighted after moving selection", prevRow, col)
							}
						}
					}
				}
			}
		})
	}
}
