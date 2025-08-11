package view

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
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

func TestHeaderUpdateContext(t *testing.T) {
	header := NewHeader()

	// Test updating context information
	header.UpdateContext(
		"test-context",
		"test-cluster",
		"test-service",
		"v1.0.0",
		"v1.28.0",
		"15%",
		"25%",
	)

	// We can't easily test the internal text content due to tview's private fields,
	// but we can verify the method doesn't panic
	assert.NotNil(t, header.leftHeader)
}

func TestHeaderUpdateArt(t *testing.T) {
	header := NewHeader()

	// Test updating art
	assert.NotPanics(t, func() {
		header.UpdateArt()
	})

	assert.NotNil(t, header.rightHeader)
}
