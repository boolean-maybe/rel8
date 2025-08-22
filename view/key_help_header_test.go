package view

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// Test helper functions that use the actual implementation

// stripColorTags removes tview color tags from text
func stripColorTags(text string) string {
	// Handle both named colors [colorname] and hex colors [#RRGGBB]
	re := regexp.MustCompile(`\[(?:[a-zA-Z-]+|#[0-9A-Fa-f]{6})\]`)
	return re.ReplaceAllString(text, "")
}

// TestKeyExplanationAlignment verifies that keyHelpHeader and explanations are properly aligned
func TestKeyExplanationAlignment(t *testing.T) {
	testCases := []struct {
		name  string
		pairs []KeyExplanationPair
	}{
		{
			name: "short keyHelpHeader and explanations",
			pairs: []KeyExplanationPair{
				{"<1>", "all"},
				{"<e>", "edit"},
				{"<q>", "quit"},
			},
		},
		{
			name: "mixed length keyHelpHeader",
			pairs: []KeyExplanationPair{
				{"<1>", "short"},
				{"<ctrl-d>", "delete"},
				{"<shift-f5>", "long key"},
				{"<x>", "restart"},
			},
		},
		{
			name: "long explanations",
			pairs: []KeyExplanationPair{
				{"<p>", "logs previous entry"},
				{"<shift-f>", "port forward"},
				{"<ctrl-alt-d>", "deep delete action"},
			},
		},
		{
			name: "maximum items",
			pairs: func() []KeyExplanationPair {
				var pairs []KeyExplanationPair
				for i := 1; i <= 20; i++ {
					pairs = append(pairs, KeyExplanationPair{
						Key:         fmt.Sprintf("<%d>", i),
						Explanation: fmt.Sprintf("action %d", i),
					})
				}
				return pairs
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Format the column
			formatted := formatKeysColumn(tc.pairs)
			lines := strings.Split(formatted, "\n")

			// Verify we have the expected number of lines
			if len(lines) != len(tc.pairs) {
				t.Errorf("Expected %d lines, got %d", len(tc.pairs), len(lines))
			}

			// Extract keyHelpHeader and explanations from each line
			var extractedKeys []string
			var extractedExplanations []string

			for i, line := range lines {
				// Strip color tags for easier parsing
				cleanLine := stripColorTags(line)

				// The format should be: key (padded to 12 chars) + space + explanation
				if len(cleanLine) < 13 {
					t.Errorf("Line %d too short: %q", i, cleanLine)
					continue
				}

				key := strings.TrimSpace(cleanLine[:12])
				explanation := strings.TrimSpace(cleanLine[13:])

				extractedKeys = append(extractedKeys, key)
				extractedExplanations = append(extractedExplanations, explanation)

				// Verify the key matches expected
				if key != tc.pairs[i].Key {
					t.Errorf("Line %d: expected key %q, got %q", i, tc.pairs[i].Key, key)
				}

				// Verify the explanation matches expected
				if explanation != tc.pairs[i].Explanation {
					t.Errorf("Line %d: expected explanation %q, got %q", i, tc.pairs[i].Explanation, explanation)
				}
			}

			// Verify alignment by checking that all keyHelpHeader start at the same position (0)
			// and all explanations start at the same position (13)
			for i, line := range lines {
				cleanLine := stripColorTags(line)

				// Find where the key ends (should be at position 12 or earlier if shorter)
				keyPart := cleanLine[:12]
				if !strings.HasPrefix(keyPart, tc.pairs[i].Key) {
					t.Errorf("Line %d: key not properly positioned at start: %q", i, keyPart)
				}

				// Verify there's a space at position 12 (after the key field)
				if len(cleanLine) > 12 && cleanLine[12] != ' ' {
					t.Errorf("Line %d: missing space separator at position 12: %q", i, cleanLine)
				}

				// Verify explanation starts at position 13
				if len(cleanLine) > 13 {
					explanationStart := strings.TrimLeft(cleanLine[13:], " ")
					if !strings.HasPrefix(explanationStart, tc.pairs[i].Explanation) {
						t.Errorf("Line %d: explanation not properly positioned: expected %q, got %q",
							i, tc.pairs[i].Explanation, explanationStart)
					}
				}
			}
		})
	}
}

// TestKeyExplanationAlignmentConsistency verifies consistent alignment across multiple columns
func TestKeyExplanationAlignmentConsistency(t *testing.T) {
	// Test data for three columns
	column1 := []KeyExplanationPair{
		{"<1>", "all"},
		{"<2>", "patching"},
		{"<3>", "default"},
	}

	column2 := []KeyExplanationPair{
		{"<ctrl-d>", "delete"},
		{"<e>", "describe"},
		{"<o>", "edit"},
	}

	column3 := []KeyExplanationPair{
		{"<p>", "logs previous"},
		{"<shift-f>", "port forward"},
		{"<x>", "restart"},
	}

	// Format all columns
	col1Text := formatKeysColumn(column1)
	col2Text := formatKeysColumn(column2)
	col3Text := formatKeysColumn(column3)

	// Split into lines
	col1Lines := strings.Split(col1Text, "\n")
	col2Lines := strings.Split(col2Text, "\n")
	col3Lines := strings.Split(col3Text, "\n")

	// Verify all columns have same number of lines
	if len(col1Lines) != len(col2Lines) || len(col2Lines) != len(col3Lines) {
		t.Errorf("Column lengths don't match: col1=%d, col2=%d, col3=%d",
			len(col1Lines), len(col2Lines), len(col3Lines))
	}

	// Verify alignment consistency across all columns
	for i := 0; i < len(col1Lines); i++ {
		// Check that explanation start position is consistent across columns
		for _, line := range []string{col1Lines[i], col2Lines[i], col3Lines[i]} {
			cleanLine := stripColorTags(line)
			if len(cleanLine) > 12 && cleanLine[12] != ' ' {
				t.Errorf("Line %d inconsistent separator: %q", i, cleanLine)
			}
		}
	}
}

// TestFormatKeyDescEdgeCases tests edge cases for the formatting function
func TestFormatKeyDescEdgeCases(t *testing.T) {
	testCases := []struct {
		key         string
		explanation string
		expectError bool
	}{
		{"", "empty key", false},
		{"<very-long-key-name>", "long key", false},
		{"<k>", "", false},
		{"<k>", "very long explanation text", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("key=%s,exp=%s", tc.key, tc.explanation), func(t *testing.T) {
			result := formatKeyDesc(tc.key, tc.explanation)

			// Basic validation - should contain both key and explanation
			if !strings.Contains(result, tc.key) {
				t.Errorf("Result doesn't contain key %q: %s", tc.key, result)
			}
			if tc.explanation != "" && !strings.Contains(result, tc.explanation) {
				t.Errorf("Result doesn't contain explanation %q: %s", tc.explanation, result)
			}

			// Verify color tags are present
			if !strings.Contains(result, "["+Colors.KeyColor+"]") || !strings.Contains(result, "["+Colors.TextDefault+"]") {
				t.Errorf("Result missing color tags: %s", result)
			}
		})
	}
}

// TestNewKeysWithPairs tests the configurable constructor
func TestNewKeysWithPairs(t *testing.T) {
	testPairs := []KeyExplanationPair{
		{"<1>", "first"},
		{"<2>", "second"},
		{"<3>", "third"},
		{"<4>", "fourth"},
		{"<5>", "fifth"},
	}

	keys := NewKeysWithPairs(testPairs)
	if keys == nil {
		t.Fatal("NewKeysWithPairs returned nil")
	}

	// Verify the component was created properly
	if keys.Flex == nil {
		t.Error("Flex component not created")
	}
	if keys.column1 == nil || keys.column2 == nil || keys.column3 == nil {
		t.Error("Column components not created")
	}
}

// TestSetKeyPairs tests the dynamic key pair setting
func TestSetKeyPairs(t *testing.T) {
	keys := NewKeys()

	testPairs := []KeyExplanationPair{
		{"<a>", "action a"},
		{"<b>", "action b"},
		{"<c>", "action c"},
	}

	keys.SetKeyPairs(testPairs)

	// Get the text from columns to verify content was set
	col1Text := keys.column1.GetText(false)
	col2Text := keys.column2.GetText(false)
	col3Text := keys.column3.GetText(false)

	// At least one column should have content
	if col1Text == "" && col2Text == "" && col3Text == "" {
		t.Error("No content set in any column")
	}

	// Combine all column text and verify our pairs are present
	allText := col1Text + col2Text + col3Text
	for _, pair := range testPairs {
		if !strings.Contains(allText, pair.Key) {
			t.Errorf("Key %q not found in columns", pair.Key)
		}
		if !strings.Contains(allText, pair.Explanation) {
			t.Errorf("Explanation %q not found in columns", pair.Explanation)
		}
	}
}

// TestDistributeKeyPairs tests the distribution algorithm
func TestDistributeKeyPairs(t *testing.T) {
	testCases := []struct {
		name         string
		pairs        []KeyExplanationPair
		expectedCols []int // expected items per column
	}{
		{
			name:         "empty pairs",
			pairs:        []KeyExplanationPair{},
			expectedCols: []int{0, 0, 0},
		},
		{
			name: "3 pairs - fills column 1",
			pairs: []KeyExplanationPair{
				{"<1>", "one"},
				{"<2>", "two"},
				{"<3>", "three"},
			},
			expectedCols: []int{3, 0, 0},
		},
		{
			name: "6 pairs - fills column 1",
			pairs: []KeyExplanationPair{
				{"<1>", "one"}, {"<2>", "two"}, {"<3>", "three"},
				{"<4>", "four"}, {"<5>", "five"}, {"<6>", "six"},
			},
			expectedCols: []int{6, 0, 0},
		},
		{
			name: "9 pairs - fills column 1 and 2",
			pairs: []KeyExplanationPair{
				{"<1>", "one"}, {"<2>", "two"}, {"<3>", "three"},
				{"<4>", "four"}, {"<5>", "five"}, {"<6>", "six"},
				{"<7>", "seven"}, {"<8>", "eight"}, {"<9>", "nine"},
			},
			expectedCols: []int{6, 3, 0},
		},
		{
			name: "12 pairs - fills columns 1 and 2",
			pairs: []KeyExplanationPair{
				{"<1>", "one"}, {"<2>", "two"}, {"<3>", "three"},
				{"<4>", "four"}, {"<5>", "five"}, {"<6>", "six"},
				{"<7>", "seven"}, {"<8>", "eight"}, {"<9>", "nine"},
				{"<10>", "ten"}, {"<11>", "eleven"}, {"<12>", "twelve"},
			},
			expectedCols: []int{6, 6, 0},
		},
		{
			name: "18 pairs - fills all columns",
			pairs: func() []KeyExplanationPair {
				var pairs []KeyExplanationPair
				for i := 1; i <= 18; i++ {
					pairs = append(pairs, KeyExplanationPair{
						Key:         fmt.Sprintf("<%d>", i),
						Explanation: fmt.Sprintf("action %d", i),
					})
				}
				return pairs
			}(),
			expectedCols: []int{6, 6, 6},
		},
		{
			name: "20 pairs - truncated to 18, fills all columns",
			pairs: func() []KeyExplanationPair {
				var pairs []KeyExplanationPair
				for i := 1; i <= 20; i++ {
					pairs = append(pairs, KeyExplanationPair{
						Key:         fmt.Sprintf("<%d>", i),
						Explanation: fmt.Sprintf("action %d", i),
					})
				}
				return pairs
			}(),
			expectedCols: []int{6, 6, 6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			col1, col2, col3 := distributeKeyPairs(tc.pairs)

			actualCols := []int{len(col1), len(col2), len(col3)}

			for i, expected := range tc.expectedCols {
				if actualCols[i] != expected {
					t.Errorf("Column %d: expected %d items, got %d", i+1, expected, actualCols[i])
				}
			}

			// Verify total items are preserved (up to 18 max)
			totalActual := len(col1) + len(col2) + len(col3)
			expectedTotal := len(tc.pairs)
			if expectedTotal > 18 {
				expectedTotal = 18 // Algorithm truncates to 18 items max
			}
			if totalActual != expectedTotal {
				t.Errorf("Total items mismatch: expected %d, got %d", expectedTotal, totalActual)
			}
		})
	}
}
