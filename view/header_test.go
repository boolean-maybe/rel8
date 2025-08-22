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
	assert.NotNil(t, header.serverInfoHeader)
}
