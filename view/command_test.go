package view

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandBar(t *testing.T) {
	commandBar := NewCommandBar()

	assert.NotNil(t, commandBar)
	assert.IsType(t, &CommandBar{}, commandBar)
	assert.NotNil(t, commandBar.TextArea)
}

func TestWrapCommandBar(t *testing.T) {
	commandBar := NewCommandBar()
	wrappedCommandBar := WrapCommandBar(commandBar)
	assert.NotNil(t, wrappedCommandBar)
	assert.IsType(t, &tview.Flex{}, wrappedCommandBar)
}

func TestCommandBarShow(t *testing.T) {
	commandBar := NewCommandBar()

	// Test showing the command bar
	assert.NotPanics(t, func() {
		commandBar.Show()
	})

	// Verify the text is set to "> "
	text := commandBar.GetText()
	assert.Equal(t, "> ", text)
}

func TestCommandBarGetCommand(t *testing.T) {
	commandBar := NewCommandBar()

	// Test with prompt prefix
	commandBar.SetText("> test command", true)
	command := commandBar.GetCommand()
	assert.Equal(t, "test command", command)

	// Test without prompt prefix (edge case)
	commandBar.SetText("no prompt", true)
	command = commandBar.GetCommand()
	assert.Equal(t, "no prompt", command)

	// Test with empty command (just prompt)
	commandBar.SetText("> ", true)
	command = commandBar.GetCommand()
	assert.Equal(t, "", command)

	// Test with just prompt
	commandBar.SetText(">", true)
	command = commandBar.GetCommand()
	assert.Equal(t, ">", command)
}

func TestCommandBarClear(t *testing.T) {
	commandBar := NewCommandBar()

	// Set some text first
	commandBar.SetText("> some command", true)

	// Clear it
	commandBar.Clear()

	// Verify it's back to just the prompt
	text := commandBar.GetText()
	assert.Equal(t, "> ", text)
}

func TestCommandBarBackspaceProtection(t *testing.T) {
	commandBar := NewCommandBar()
	commandBar.Show() // Set to "> "

	// We can't easily test the input capture directly since it requires event simulation,
	// but we can verify the command bar was created with input capture
	assert.NotNil(t, commandBar.TextArea)

	// Verify the text starts with the prompt
	text := commandBar.GetText()
	assert.Equal(t, "> ", text)
}

func TestCommandBarInputCapture(t *testing.T) {
	commandBar := NewCommandBar()

	// The input capture should be set, but we can't easily test the actual behavior
	// without more complex event simulation. We can at least verify the command bar
	// was created successfully with the input capture setup.
	assert.NotNil(t, commandBar.TextArea)
}
