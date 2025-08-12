package view

import (
	"reflect"
	"rel8/db"
	"unsafe"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Grid wraps a Table with grid-specific functionality
type Grid struct {
	*tview.Table
	headerCount int
}

// NewGrid creates a new grid with proper configuration
func NewGrid(headers []string, data []db.TableData) *Grid {
	table := configureTable()
	grid := &Grid{Table: table, headerCount: len(headers)}
	attachSelectionHandler(grid)
	grid.Populate(headers, data)
	return grid
}

// NewEmptyGrid creates a new empty grid with proper configuration
func NewEmptyGrid() *Grid {
	table := configureTable()
	grid := &Grid{Table: table}
	attachSelectionHandler(grid)
	return grid
}

// Populate fills the grid with headers and data
func (g *Grid) Populate(headers []string, data []db.TableData) {
	g.Clear()
	g.headerCount = len(headers)

	// add headers
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(Colors.TextWhite).
			SetBackgroundColor(Colors.BackgroundDefault).
			SetAttributes(tcell.AttrBold)

		// Set column expansion (uniform for predictable layout)
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

	// ensure the last visible column absorbs any leftover width so the highlight
	// visually spans to the right edge even when content is narrow
	g.StretchLastColumn()

	// Start selection at first data row, not header, and scroll to top
	g.Select(1, 0)
	g.ScrollToBeginning()
	// ensure initial selection row is visually highlighted
	g.RefreshSelectionHighlight()
}

// RestoreSelection restores the selected row if valid
func (g *Grid) RestoreSelection(selectedIndex int, dataLen int) {
	if selectedIndex >= 0 && selectedIndex < dataLen {
		g.Select(selectedIndex+1, 0) // +1 because table has header row
		g.RefreshSelectionHighlight()
	}
}

// RefreshSelectionHighlight reapplies styles so that the entire currently-selected row
// is highlighted and all other data rows are reset to default styling
func (g *Grid) RefreshSelectionHighlight() {
	if g.Table == nil {
		return
	}
	selectedRow, _ := g.GetSelection()
	// styles
	selectionStyle := tcell.StyleDefault.
		Background(Colors.TextAqua).
		Foreground(Colors.TextBlack)
	defaultDataRowStyle := tcell.StyleDefault.
		Background(Colors.BackgroundDefault).
		Foreground(Colors.TextLightSkyBlue)

	// ensure selected row has cells up to headerCount so highlight applies to the rightmost column
	if g.headerCount > 0 {
		selRow, _ := g.GetSelection()
		if selRow > 0 && selRow < g.GetRowCount() {
			for col := 0; col < g.headerCount; col++ {
				if g.GetCell(selRow, col) == nil {
					g.SetCell(selRow, col, tview.NewTableCell("").SetExpansion(1))
				}
			}
		}
	}

	// reset all data rows (exclude header row 0)
	totalRows := g.GetRowCount()
	// cover up to the max of headerCount and current table columns (to include any runtime filler)
	totalCols := g.GetColumnCount()
	if g.headerCount > totalCols {
		totalCols = g.headerCount
	}
	for row := 1; row < totalRows; row++ {
		for col := 0; col < totalCols; col++ {
			cell := g.GetCell(row, col)
			// do not synthesize cells: that breaks layout; only style existing cells
			if cell == nil {
				continue
			}
			if row == selectedRow {
				cell.SetStyle(selectionStyle)
			} else {
				cell.SetStyle(defaultDataRowStyle)
			}
		}
	}
}

// StretchLastColumn increases the expansion weight on the last existing column
// to absorb leftover width so the selected row's background appears to span to the edge.
// this mimics a row-level band without introducing synthetic columns.
func (g *Grid) StretchLastColumn() {
	if g.Table == nil {
		return
	}
	cols := g.GetColumnCount()
	if cols == 0 {
		return
	}
	last := cols - 1
	rows := g.GetRowCount()
	for r := 0; r < rows; r++ {
		if cell := g.GetCell(r, last); cell != nil {
			cell.SetExpansion(10)
		}
	}
}

// getRowOffsetUnsafe reads tview.Table's private rowOffset field to align the band with the visible row
// drawing a cursor band at the actual on-screen row position
func (g *Grid) getRowOffsetUnsafe() int {
	if g.Table == nil {
		return 0
	}
	v := reflect.ValueOf(g.Table).Elem()
	f := v.FieldByName("rowOffset")
	if !f.IsValid() {
		return 0
	}
	// access unexported field via unsafe pointer
	ptr := unsafe.Pointer(f.UnsafeAddr())
	off := *(*int)(ptr)
	if off < 0 {
		return 0
	}
	return off
}

// DrawSelectionBand paints a full-width band on the selected row across the table's inner rect
// cursor band while keeping normal drawing intact (use via Application.SetAfterDrawFunc)
func (g *Grid) DrawSelectionBand(screen tcell.Screen) {
	if g.Table == nil {
		return
	}
	ix, iy, iw, ih := g.Table.GetInnerRect()
	if iw <= 0 || ih <= 0 {
		return
	}
	selRow, _ := g.Table.GetSelection()
	if selRow <= 0 || selRow >= g.Table.GetRowCount() {
		return
	}
	rowOffset := g.getRowOffsetUnsafe()
	rowY := iy + (selRow - rowOffset)
	if rowY < iy || rowY >= iy+ih {
		return
	}
	for cx := 0; cx < iw; cx++ {
		mainc, combc, st, _ := screen.GetContent(ix+cx, rowY)
		// preserve existing rune and foreground; only change background to the selection band color
		st = st.Background(Colors.SelectionBandBg)
		screen.SetContent(ix+cx, rowY, mainc, combc, st)
	}
}

// AttachSelectionBand registers an after-draw hook on the app that paints
// a full-width selection band when shouldDraw returns true. It preserves
// any existing after-draw handler by chaining it first.
func (g *Grid) AttachSelectionBand(app *tview.Application, shouldDraw func() bool) {
	if app == nil || g == nil {
		return
	}
	prev := app.GetAfterDrawFunc()
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if prev != nil {
			prev(screen)
		}
		if shouldDraw != nil && shouldDraw() {
			g.DrawSelectionBand(screen)
		}
	})
}

// Select overrides tview.Table.Select to also refresh full-row highlighting
func (g *Grid) Select(row, column int) *tview.Table {
	if g.Table == nil {
		return nil
	}
	g.Table.Select(row, column)
	g.RefreshSelectionHighlight()
	return g.Table
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
	// Define selection style for the cells
	selectionStyle := tcell.StyleDefault.
		Background(Colors.TextAqua).
		Foreground(Colors.TextBlack)
	table.SetSelectedStyle(selectionStyle)

	return table
}

// attachSelectionHandler wires selection changes to full-row highlighting using the Grid instance
func attachSelectionHandler(g *Grid) {
	if g == nil || g.Table == nil {
		return
	}
	g.Table.SetSelectionChangedFunc(func(row, column int) {
		// prevent selecting header row
		if row == 0 && g.Table.GetRowCount() > 1 {
			g.Table.Select(1, column)
			return
		}
		g.RefreshSelectionHighlight()
	})
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
