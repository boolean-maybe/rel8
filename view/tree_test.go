package view

import (
	"rel8/db"
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	
	if tree == nil {
		t.Error("Expected tree to be created")
	}
	
	if tree.TreeView == nil {
		t.Error("Expected TreeView to be initialized")
	}
	
	if tree.nodes == nil {
		t.Error("Expected nodes map to be initialized")
	}
}

func TestTreeSetServer(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	
	tree.SetServer(server)
	
	if tree.server != server {
		t.Error("Expected server to be set")
	}
	
	// Tree should have a root node after setting server
	if tree.rootNode == nil {
		t.Error("Expected root node to be created after setting server")
	}
	
	if tree.rootNode.Type != "server" {
		t.Errorf("Expected root node type to be 'server', got %s", tree.rootNode.Type)
	}
}

func TestTreePopulateTree(t *testing.T) {
	tree := NewTree()
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	
	tree.SetServer(server)
	
	// Root should have database children
	if len(tree.rootNode.Children) == 0 {
		t.Error("Expected root node to have database children")
	}
	
	// First child should be a database
	firstChild := tree.rootNode.Children[0]
	if firstChild.Type != "database" {
		t.Errorf("Expected first child to be database, got %s", firstChild.Type)
	}
	
	// Database should have category children (Tables, Views, etc.)
	if len(firstChild.Children) == 0 {
		t.Error("Expected database node to have category children")
	}
	
	// Check for expected categories
	categories := make(map[string]bool)
	for _, child := range firstChild.Children {
		if child.Type != "category" {
			t.Errorf("Expected category node, got %s", child.Type)
		}
		categories[child.Name] = true
	}
	
	expectedCategories := []string{"Tables", "Views", "Procedures", "Functions", "Triggers"}
	for _, expected := range expectedCategories {
		if !categories[expected] {
			t.Errorf("Expected category %s not found", expected)
		}
	}
}

func TestTreeGetCurrentDatabase(t *testing.T) {
	tree := NewTree()
	
	// Initially should be empty
	if tree.GetCurrentDatabase() != "" {
		t.Error("Expected current database to be empty initially")
	}
	
	// Set a database
	tree.currentDB = "testdb"
	if tree.GetCurrentDatabase() != "testdb" {
		t.Error("Expected current database to be 'testdb'")
	}
}

func TestWrapTree(t *testing.T) {
	tree := NewTree()
	wrapped := WrapTree(tree)
	
	if wrapped == nil {
		t.Error("Expected wrapped tree to be created")
	}
	
	// Should be a Flex container
	if wrapped.GetItemCount() != 3 {
		t.Error("Expected wrapped tree to have 3 items (padding + tree + padding)")
	}
}