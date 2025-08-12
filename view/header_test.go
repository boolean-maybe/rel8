package view

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"rel8/db"
)

func TestNewHeader(t *testing.T) {
	header := NewHeader()

	assert.NotNil(t, header)
	assert.IsType(t, &Header{}, header)
	assert.NotNil(t, header.Flex)
}

func TestWrapHeader(t *testing.T) {
	header := NewHeader()
	wrappedHeader := WrapHeader(header)
	assert.NotNil(t, wrappedHeader)
	assert.IsType(t, &tview.Flex{}, wrappedHeader)
}

func TestHeaderUpdateServerInfo(t *testing.T) {
	header := NewHeader()

	// Test updating server info without a server set
	// Should not panic
	assert.NotPanics(t, func() {
		header.UpdateServerInfo()
	})

	// We can't easily test the internal text content due to tview's private fields,
	// but we can verify the method doesn't panic
	assert.NotNil(t, header.leftHeader)
}

func TestHeaderSetServer(t *testing.T) {
	header := NewHeader()

	// Create a mock server
	mockServer := &db.MysqlMock{}

	// Set the server
	assert.NotPanics(t, func() {
		header.SetServer(mockServer)
	})

	// Verify server was set
	assert.Equal(t, mockServer, header.server)
}

func TestGetValueWithDefault(t *testing.T) {
	testMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "",
	}

	// Test existing key
	assert.Equal(t, "value1", getValueWithDefault(testMap, "key1", "default"))

	// Test non-existing key
	assert.Equal(t, "default", getValueWithDefault(testMap, "nonexistent", "default"))

	// Test empty value
	assert.Equal(t, "", getValueWithDefault(testMap, "key3", "default"))
}

func TestHeaderUpdateArt(t *testing.T) {
	header := NewHeader()

	// Test updating art
	assert.NotPanics(t, func() {
		header.UpdateArt()
	})

	assert.NotNil(t, header.rightHeader)
}
