package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"rel8/db"
)

// Detail wraps a TextView with detail-specific functionality
type Detail struct {
	*tview.TextView
}

// NewDetail creates a new detail view with proper configuration
func NewDetail(text string) *Detail {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	// Apply SQL syntax highlighting if the text looks like SQL
	// (typically CREATE TABLE statements from FetchTableDescr)
	highlightedText := db.HighlightSQL(text)

	textView.SetText(highlightedText)
	textView.SetBackgroundColor(tcell.ColorBlack)

	// Add same border styling as table
	textView.SetBorder(true).SetBorderPadding(0, 0, 1, 1).SetBorderColor(tcell.ColorLightSkyBlue)
	textView.SetBorderAttributes(tcell.AttrNone)

	return &Detail{TextView: textView}
}

// NewEmptyDetail creates a new empty detail view
func NewEmptyDetail() *Detail {
	return NewDetail("")
}

// UpdateText updates the detail view with new text
func (d *Detail) UpdateText(text string) {
	// Apply SQL syntax highlighting
	highlightedText := db.HighlightSQL(text)
	d.SetText(highlightedText)
}

// WrapDetail wraps detail with padding (only left/right, NO top/bottom)
func WrapDetail(detail *Detail) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(detail.TextView, 0, 1, true).
		AddItem(nil, 0, 0, false) // Right padding
}
