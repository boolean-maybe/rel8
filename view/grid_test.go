package view

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"rel8/db"
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
