package view

import (
	"rel8/db"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTreeHandleInput(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	tree.SetServer(server)

	// Create a mock Enter key event
	enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)

	// Test handling Enter key
	result := tree.handleInput(enterEvent)

	// Enter key should be consumed (return nil)
	if result != nil {
		t.Error("Expected Enter key to be consumed (return nil)")
	}
}

func TestTreeHandleInputSpaceKey(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	tree.SetServer(server)

	// Create a mock Space key event
	spaceEvent := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)

	// Test handling Space key
	result := tree.handleInput(spaceEvent)

	// Space key should be consumed (return nil)
	if result != nil {
		t.Error("Expected Space key to be consumed (return nil)")
	}
}

func TestTreeHandleInputArrowKey(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	tree.SetServer(server)

	// Create a mock Arrow Down key event
	arrowEvent := tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)

	// Test handling Arrow key
	result := tree.handleInput(arrowEvent)

	// Arrow key should pass through (return the event)
	if result != arrowEvent {
		t.Error("Expected Arrow key to pass through (return the event)")
	}
}

func TestTreeHandleNodeActivation(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	tree.SetServer(server)

	// Get the root node (server node)
	rootNode := tree.treeView.GetRoot()
	if rootNode == nil {
		t.Fatal("Expected root node to exist")
	}

	// Initially the root should be expanded
	initialExpanded := rootNode.IsExpanded()

	// Activate the root node
	tree.HandleNodeActivation(rootNode)

	// Expansion state should toggle
	finalExpanded := rootNode.IsExpanded()
	if initialExpanded == finalExpanded {
		t.Error("Expected root node expansion state to toggle")
	}
}