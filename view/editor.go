package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"rel8/db"
)

// Editor wraps a TextArea with editor-specific functionality
type Editor struct {
	*tview.TextArea
}

// NewEditor creates a new editor view with proper configuration
func NewEditor(text string) *Editor {
	textArea := tview.NewTextArea().
		SetWrap(false)

	// Apply SQL syntax highlighting if the text looks like SQL
	// (typically CREATE TABLE statements from FetchTableDescr)
	highlightedText := db.HighlightSQL(text)

	textArea.SetText(highlightedText, false)
	textArea.SetBackgroundColor(Colors.BackgroundDefault)

	// Add same border styling as table
	textArea.SetBorder(true).SetBorderPadding(0, 0, 1, 1).SetBorderColor(Colors.BorderDefault)
	textArea.SetBorderAttributes(tcell.AttrNone)

	return &Editor{TextArea: textArea}
}

// NewEmptyEditor creates a new empty editor view
func NewEmptyEditor() *Editor {
	return NewEditor("")
}

// UpdateText updates the editor view with new text
func (e *Editor) UpdateText(text string) {
	// Apply SQL syntax highlighting
	highlightedText := db.HighlightSQL(text)
	e.SetText(highlightedText, false)
}

// WrapEditor wraps editor with padding (only left/right, NO top/bottom)
func WrapEditor(editor *Editor) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(editor.TextArea, 0, 1, true).
		AddItem(nil, 0, 0, false) // Right padding
}
