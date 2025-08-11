package view

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestNewDetail(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		expectedText string
	}{
		{
			name:         "create empty detail",
			text:         "",
			expectedText: "",
		},
		{
			name:         "create text detail",
			text:         "Sample details information",
			expectedText: "Sample details information", // Non-SQL text should pass through unchanged
		},
		{
			name: "create CREATE TABLE detail with SQL highlighting",
			text: `CREATE TABLE users (
  id INT PRIMARY KEY,
  name VARCHAR(255)
)`,
			expectedText: `[lightblue]CREATE[-] [lightblue]TABLE[-] users (
  id [lightgreen]INT[-] [lightblue]PRIMARY[-] [lightblue]KEY[-],
  name [lightgreen]VARCHAR[-]([magenta]255[-])
)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := NewDetail(tt.text)

			assert.NotNil(t, detail)
			assert.IsType(t, &Detail{}, detail)
			assert.NotNil(t, detail.TextView)

			text := detail.GetText(false)
			assert.Equal(t, tt.expectedText, text)

			// Verify styling properties
			assert.Equal(t, tcell.ColorBlack, detail.GetBackgroundColor())
		})
	}
}

func TestNewEmptyDetail(t *testing.T) {
	detail := NewEmptyDetail()

	assert.NotNil(t, detail)
	assert.IsType(t, &Detail{}, detail)
	assert.NotNil(t, detail.TextView)

	text := detail.GetText(false)
	assert.Equal(t, "", text)
}

func TestDetailUpdateText(t *testing.T) {
	detail := NewEmptyDetail()

	// Update with plain text (should not be highlighted)
	detail.UpdateText("Simple message")
	text := detail.GetText(false)
	assert.Equal(t, "Simple message", text)

	// Update with SQL text
	sqlText := "SELECT * FROM users"
	detail.UpdateText(sqlText)
	text = detail.GetText(false)
	expected := "[lightblue]SELECT[-] * [lightblue]FROM[-] users"
	assert.Equal(t, expected, text)

	// Update with empty text
	detail.UpdateText("")
	text = detail.GetText(false)
	assert.Equal(t, "", text)
}

func TestWrapDetail(t *testing.T) {
	detail := NewDetail("test content")
	wrappedDetail := WrapDetail(detail)

	assert.NotNil(t, wrappedDetail)
	assert.IsType(t, &tview.Flex{}, wrappedDetail)
}

func TestDetailStyling(t *testing.T) {
	detail := NewDetail("test")

	// Verify border and styling
	assert.Equal(t, tcell.ColorBlack, detail.GetBackgroundColor())

	// Verify the detail has dynamic colors and proper wrap settings
	// (We can't easily test these private properties, but we can verify creation doesn't panic)
	assert.NotNil(t, detail.TextView)
}
