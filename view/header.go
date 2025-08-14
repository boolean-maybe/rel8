package view

import (
	"context"
	"github.com/rivo/tview"
	"rel8/config"
	"rel8/db"
)

// Header wraps a Flex with header-specific functionality
type Header struct {
	*tview.Flex
	leftHeader  *tview.TextView
	keys        *Keys
	rightHeader *tview.TextView
	server      db.DatabaseServer
}

// NewHeader creates a new header with proper configuration
func NewHeader() *Header {
	leftHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	// Default placeholder text - will be updated once server is set
	headerText := ` [` + Colors.HeaderLabel + `]Server: [` + Colors.HeaderValue + `]Not connected[-]
 [` + Colors.HeaderLabel + `]User: [` + Colors.HeaderValue + `]N/A[-]
 [` + Colors.HeaderLabel + `]Database: [` + Colors.HeaderValue + `]N/A[-]
 [` + Colors.HeaderLabel + `]Host: [` + Colors.HeaderValue + `]N/A[-]
 [` + Colors.HeaderLabel + `]Port: [` + Colors.HeaderValue + `]N/A[-]
 [` + Colors.HeaderLabel + `]Max Conn: [` + Colors.HeaderValue + `]N/A[-]
 [` + Colors.HeaderLabel + `]Buffer Pool: [` + Colors.HeaderValue + `]N/A[-]`

	leftHeader.SetText(headerText)
	leftHeader.SetBackgroundColor(Colors.BackgroundDefault)

	rightHeader := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(tview.AlignRight)

	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	rightHeader.SetText(artText)
	rightHeader.SetBackgroundColor(Colors.BackgroundDefault)

	keys := NewKeys()

	headerFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftHeader, 0, 1, false).
		AddItem(keys.Flex, 0, 1, false).
		AddItem(rightHeader, 0, 1, false).
		AddItem(nil, 1, 0, false)

	return &Header{
		Flex:        headerFlex,
		leftHeader:  leftHeader,
		keys:        keys,
		rightHeader: rightHeader,
	}
}

// SetServer sets the database server and updates server info
func (h *Header) SetServer(server db.DatabaseServer) {
	h.server = server
	h.UpdateServerInfo()
}

// UpdateServerInfo fetches and displays current MySQL server information
func (h *Header) UpdateServerInfo() {
	if h.server == nil {
		return
	}

	// Create a short-lived context for the server info query
	ctx, cancel := context.WithTimeout(context.Background(), 2*1000000000) // 2 seconds
	defer cancel()

	// Get server information
	serverInfo := h.server.GetServerInfo(ctx)

	// Extract values with defaults for missing entries
	version := getValueWithDefault(serverInfo, "version", "Unknown")
	user := getValueWithDefault(serverInfo, "user", "Unknown")
	database := getValueWithDefault(serverInfo, "database", "N/A")
	host := getValueWithDefault(serverInfo, "host", "Unknown")
	port := getValueWithDefault(serverInfo, "port", "Unknown")
	maxConn := getValueWithDefault(serverInfo, "max_connections", "Unknown")

	// Check for PostgreSQL shared_buffers or MySQL innodb_buffer_pool_size
	bufferPool := getValueWithDefault(serverInfo, "shared_buffers", "Unknown")
	if bufferPool == "Unknown" {
		bufferPool = getValueWithDefault(serverInfo, "innodb_buffer_pool_size", "Unknown")
	}

	// Determine the buffer label based on the type of buffer setting found
	bufferLabel := "Buffer Pool"
	if _, hasSharedBuffers := serverInfo["shared_buffers"]; hasSharedBuffers {
		bufferLabel = "Shared Buffers"
	}

	headerText := ` [` + Colors.HeaderLabel + `]Server: [` + Colors.HeaderValue + `]` + version + `[-]
 [` + Colors.HeaderLabel + `]User: [` + Colors.HeaderValue + `]` + user + `[-]
 [` + Colors.HeaderLabel + `]Database: [` + Colors.HeaderValue + `]` + database + `[-]
 [` + Colors.HeaderLabel + `]Host: [` + Colors.HeaderValue + `]` + host + `[-]
 [` + Colors.HeaderLabel + `]Port: [` + Colors.HeaderValue + `]` + port + `[-]
 [` + Colors.HeaderLabel + `]Max Conn: [` + Colors.HeaderValue + `]` + maxConn + `[-]
 [` + Colors.HeaderLabel + `]` + bufferLabel + `: [` + Colors.HeaderValue + `]` + bufferPool + `[-]`

	h.leftHeader.SetText(headerText)
}

// Helper function to get a value from a map with a default if key doesn't exist
func getValueWithDefault(m map[string]string, key, defaultValue string) string {
	if value, exists := m[key]; exists {
		return value
	}
	return defaultValue
}

// UpdateKeys updates the keys display in the middle section
func (h *Header) UpdateKeys(text string) {
	h.keys.UpdateKeys(text)
}

// UpdateArt updates the art on the right side of the header
func (h *Header) UpdateArt() {
	artText := config.GetArt()
	artText = tview.TranslateANSI(artText)
	h.rightHeader.SetText(artText)
}

// WrapHeader wraps header with same padding as other components
func WrapHeader(header *Header) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(header.Flex, 0, 1, false).
		AddItem(nil, 0, 0, false) // Right padding
}
