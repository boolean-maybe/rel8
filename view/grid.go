package view

import (
	"reflect"
	"rel8/db"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Grid wraps a Table with grid-specific functionality
type Grid struct {
	*tview.Table
}

// NewGrid creates a new grid with proper configuration
func NewGrid(headers []string, data []db.TableData) *Grid {
	table := configureTable()
	grid := &Grid{Table: table}
	grid.Populate(headers, data)
	return grid
}

// NewEmptyGrid creates a new empty grid with proper configuration
func NewEmptyGrid() *Grid {
	table := configureTable()
	return &Grid{Table: table}
}

// Populate fills the grid with headers and data
func (g *Grid) Populate(headers []string, data []db.TableData) {
	g.Clear()

	// add headers
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(Colors.TextWhite).
			SetBackgroundColor(Colors.BackgroundDefault).
			SetAttributes(tcell.AttrBold)

		// Set expansion for header
		setExpansion(col, cell)

		g.SetCell(0, col, cell)
	}

	// Add table data
	for row, item := range data {
		// extract fields of db.TableData runtime type, using headers for map data
		fields := getFieldsWithHeaders(item, headers)
		for col, field := range fields {
			cell := tview.NewTableCell(field).SetTextColor(Colors.TextLightSkyBlue)
			setExpansion(col, cell)
			g.SetCell(row+1, col, cell)
		}
	}

	// Start selection at first data row, not header, and scroll to top
	g.Select(1, 0)
	g.ScrollToBeginning()
}

// RestoreSelection restores the selected row if valid
func (g *Grid) RestoreSelection(selectedIndex int, dataLen int) {
	if selectedIndex >= 0 && selectedIndex < dataLen {
		g.Select(selectedIndex+1, 0) // +1 because table has header row
	}
}

// configureTable creates and configures a new table
func configureTable() *tview.Table {
	// Force single line borders by setting tview.Borders to use light characters for focus
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	// Add single line border around the entire table
	table.SetBorder(true).SetBorderPadding(0, 0, 1, 1).SetBorderColor(Colors.BorderDefault)
	table.SetBorderAttributes(tcell.AttrNone)

	// Fix header row and make table scrollable
	table.SetFixed(1, 0)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(Colors.TextAqua).
		Foreground(Colors.TextBlack))

	// Set selection changed handler to prevent selecting header row
	table.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 {
			// If trying to select header row, move to first data row
			table.Select(1, column)
		}
	})

	return table
}

// setExpansion sets cell expansion properties
func setExpansion(col int, cell *tview.TableCell) {
	// Set expansion for first column only ?
	//if col == 0 {
	cell.SetExpansion(1)
	//}
}

// WrapGrid wraps grid with padding (only left/right, NO top/bottom)
func WrapGrid(grid *Grid) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(grid.Table, 0, 1, true).
		AddItem(nil, 0, 0, false) // Right padding
}

// getFields extracts fields from a struct
func getFields(item interface{}) []string {
	var fields []string
	v := reflect.ValueOf(item)

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fields = append(fields, v.Field(i).String())
		}
	}
	return fields
}

// getFieldsWithHeaders extracts fields from map using headers to maintain order
func getFieldsWithHeaders(item interface{}, headers []string) []string {
	var fields []string

	// Try map first
	if mapData, ok := item.(map[string]string); ok {
		for _, header := range headers {
			if value, exists := mapData[header]; exists {
				fields = append(fields, value)
			} else {
				fields = append(fields, "")
			}
		}
		return fields
	}

	// Fall back to struct handling
	return getFields(item)
}
