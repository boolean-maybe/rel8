package view

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// KeyExplanationPair represents a key binding and its explanation
type KeyExplanationPair struct {
	Key         string
	Explanation string
}

// KeyHelpHeader wraps a Flex with keyHelpHeader-specific functionality
type KeyHelpHeader struct {
	*tview.Flex
	column1 *tview.TextView
	column2 *tview.TextView
	column3 *tview.TextView
}

// NewKeys creates a new keyHelpHeader view with proper configuration
func NewKeys() *KeyHelpHeader {
	return NewKeysWithPairs(getDefaultKeyPairs())
}

// NewKeysWithPairs creates a new keyHelpHeader view with the provided key/explanation pairs
func NewKeysWithPairs(pairs []KeyExplanationPair) *KeyHelpHeader {
	// Create three columns for perfect alignment
	column1 := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	column1.SetBackgroundColor(Colors.BackgroundDefault)

	column2 := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	column2.SetBackgroundColor(Colors.BackgroundDefault)

	column3 := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	column3.SetBackgroundColor(Colors.BackgroundDefault)

	// Create flex layout with three columns with spacing
	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(column1, 0, 2, false).
		AddItem(nil, 1, 0, false). // Small spacer
		AddItem(column2, 0, 2, false).
		AddItem(nil, 1, 0, false). // Small spacer
		AddItem(column3, 0, 2, false)

	keys := &KeyHelpHeader{
		Flex:    flex,
		column1: column1,
		column2: column2,
		column3: column3,
	}

	// Set the key pairs
	keys.SetKeyPairs(pairs)

	return keys
}

// getDefaultKeyPairs returns default mock key pairs for backward compatibility
func getDefaultKeyPairs() []KeyExplanationPair {
	return []KeyExplanationPair{
		{"<1>", "all"},
		{"<1>", "patching"},
		{"<2>", "default"},
		{"<shift-j>", "Jump Owner"},
		{"<ctrl-d>", "Delete"},
		{"<e>", "Describe"},
		{"<o>", "Edit"},
		{"<2>", "Help"},
		{"<r>", "Show Ports"},
		{"<y>", "YAML"},
		{"<p>", "Logs Prev"},
		{"<shift-f>", "Port Forward"},
		{"<x>", "Restart"},
		{"<s>", "Scale"},
		{"<v>", "View Logs"},
	}
}

// formatKeyDesc formats a key-description pair with fixed width alignment
func formatKeyDesc(key, desc string) string {
	return fmt.Sprintf("[%s]%-12s[%s] %s", Colors.KeyColor, key, Colors.TextDefault, desc)
}

// formatKeysColumn formats a list of key-explanation pairs into a column text
func formatKeysColumn(pairs []KeyExplanationPair) string {
	var lines []string
	for _, pair := range pairs {
		lines = append(lines, formatKeyDesc(pair.Key, pair.Explanation))
	}
	return strings.Join(lines, "\n")
}

// distributeKeyPairs distributes key pairs across three columns targeting 6 rows per column
func distributeKeyPairs(pairs []KeyExplanationPair) ([]KeyExplanationPair, []KeyExplanationPair, []KeyExplanationPair) {
	if len(pairs) == 0 {
		return nil, nil, nil
	}

	const targetRowsPerColumn = 6
	maxItems := targetRowsPerColumn * 3 // 18 total items max

	// Limit to maximum displayable items
	displayPairs := pairs
	if len(pairs) > maxItems {
		displayPairs = pairs[:maxItems]
	}

	// Distribute items targeting 6 rows per column
	var col1, col2, col3 []KeyExplanationPair

	idx := 0

	// Fill column 1 (up to 6 items)
	for i := 0; i < targetRowsPerColumn && idx < len(displayPairs); i++ {
		col1 = append(col1, displayPairs[idx])
		idx++
	}

	// Fill column 2 (up to 6 items)
	for i := 0; i < targetRowsPerColumn && idx < len(displayPairs); i++ {
		col2 = append(col2, displayPairs[idx])
		idx++
	}

	// Fill column 3 (up to 6 items)
	for i := 0; i < targetRowsPerColumn && idx < len(displayPairs); i++ {
		col3 = append(col3, displayPairs[idx])
		idx++
	}

	return col1, col2, col3
}

// SetKeyPairs updates the display with new key/explanation pairs
func (k *KeyHelpHeader) SetKeyPairs(pairs []KeyExplanationPair) {
	col1, col2, col3 := distributeKeyPairs(pairs)

	col1Text := formatKeysColumn(col1)
	col2Text := formatKeysColumn(col2)
	col3Text := formatKeysColumn(col3)

	k.column1.SetText(col1Text)
	k.column2.SetText(col2Text)
	k.column3.SetText(col3Text)
}

// UpdateKeys updates the keyHelpHeader display (legacy method for backward compatibility)
func (k *KeyHelpHeader) UpdateKeys(text string) {
	// For backward compatibility, just update column1
	k.column1.SetText(text)
}

// UpdateColumns updates each column separately (legacy method)
func (k *KeyHelpHeader) UpdateColumns(col1, col2, col3 string) {
	k.column1.SetText(col1)
	k.column2.SetText(col2)
	k.column3.SetText(col3)
}
