package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CommandBar wraps a TextArea with command-specific functionality
type CommandBar struct {
	*tview.TextArea
}

// NewCommandBar creates a new command bar with proper configuration
func NewCommandBar() *tview.TextArea {
	textArea := tview.NewTextArea()
	textArea.SetBackgroundColor(Colors.BackgroundDefault)
	textArea.SetBorder(true).SetBorderPadding(0, 0, 0, 0).SetBorderColor(Colors.BorderDefault)

	// Prevent backspace from erasing the "> " prompt
	textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			text := textArea.GetText()
			if len(text) <= 3 {
				// Don't allow backspace if we're at or before the prompt
				return nil
			}
		}
		return event
	})

	textArea.SetText(" > ", true)
	return textArea
}

// GetCommand returns the command text without the "> " prefix
func (cb *CommandBar) GetCommand() string {
	text := cb.GetText()
	if len(text) >= 2 && text[:2] == "> " {
		return strings.TrimSpace(text[2:])
	}
	return text
}

// Clear resets the command bar
func (cb *CommandBar) Clear() {
	cb.SetText("> ", true)
}
