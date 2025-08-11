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
func NewCommandBar() *CommandBar {
	// Create command bar (initially hidden)
	textArea := tview.NewTextArea()
	textArea.SetBackgroundColor(tcell.ColorBlack)
	textArea.SetBorder(true).SetBorderPadding(0, 0, 0, 0).SetBorderColor(tcell.ColorLightSkyBlue)

	// Prevent backspace from erasing the "> " prompt
	textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			text := textArea.GetText()
			if len(text) <= 2 {
				// Don't allow backspace if we're at or before the prompt
				return nil
			}
		}
		return event
	})

	return &CommandBar{TextArea: textArea}
}

// Show initializes the command bar for display
func (cb *CommandBar) Show() {
	cb.SetText("> ", true)
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

// WrapCommandBar wraps command bar with same padding to align borders
func WrapCommandBar(commandBar *CommandBar) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(commandBar.TextArea, 0, 1, true).
		AddItem(nil, 0, 0, false) // Right padding
}
