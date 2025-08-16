package view

import (
	"rel8/db"
	"rel8/model"
	"testing"
)

func TestFindNextNode(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Test: Next node from root should be first database
	nextNode := view.findNextNode(root)
	if nextNode == nil {
		t.Error("Expected to find next node from root")
	}

	// Root should be expanded with database children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) > 0 {
		nextNode = view.findNextNode(root)
		if nextNode != children[0] {
			t.Error("Expected next node from expanded root to be first child")
		}
	}
}

func TestFindPrevNode(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Expand root to get database children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) == 0 {
		t.Skip("No database children found")
	}

	// Test: Previous node from first database should be root
	firstDB := children[0]
	prevNode := view.findPrevNode(firstDB)
	if prevNode != root {
		t.Error("Expected previous node from first database to be root")
	}
}

func TestFindNextSibling(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Expand root to get database children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) < 2 {
		t.Skip("Need at least 2 database children for sibling test")
	}

	// Test: Next sibling of first database should be second database
	firstDB := children[0]
	nextSibling := view.findNextSibling(firstDB)
	if nextSibling != children[1] {
		t.Error("Expected next sibling to be second database")
	}

	// Test: Last database should have no next sibling
	lastDB := children[len(children)-1]
	nextSibling = view.findNextSibling(lastDB)
	if nextSibling != nil {
		t.Error("Expected last database to have no next sibling")
	}
}

func TestFindPrevSibling(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Expand root to get database children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) < 2 {
		t.Skip("Need at least 2 database children for sibling test")
	}

	// Test: Previous sibling of second database should be first database
	secondDB := children[1]
	prevSibling := view.findPrevSibling(secondDB)
	if prevSibling != children[0] {
		t.Error("Expected previous sibling to be first database")
	}

	// Test: First database should have no previous sibling
	firstDB := children[0]
	prevSibling = view.findPrevSibling(firstDB)
	if prevSibling != nil {
		t.Error("Expected first database to have no previous sibling")
	}
}

func TestFindParent(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Test: Root should have no parent
	parent := view.findParent(root)
	if parent != nil {
		t.Error("Expected root to have no parent")
	}

	// Expand root to get database children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) == 0 {
		t.Skip("No database children found")
	}

	// Test: Database parent should be root
	firstDB := children[0]
	parent = view.findParent(firstDB)
	if parent != root {
		t.Error("Expected database parent to be root")
	}
}

func TestFindLastDescendant(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Test: Collapsed node should return itself
	root.SetExpanded(false)
	lastDesc := view.findLastDescendant(root)
	if lastDesc != root {
		t.Error("Expected collapsed node to return itself as last descendant")
	}

	// Test: Expanded node with children
	root.SetExpanded(true)
	children := root.GetChildren()
	if len(children) > 0 {
		// If children are not expanded, last descendant should be last child
		lastChild := children[len(children)-1]
		lastChild.SetExpanded(false)
		lastDesc = view.findLastDescendant(root)
		if lastDesc != lastChild {
			t.Error("Expected last descendant to be last child when children not expanded")
		}
	}
}

func TestTreeNavigationIntegration(t *testing.T) {
	// Create a view with tree
	stateManager := &model.ContextualStateManager{}
	view := NewView(stateManager)
	server := &db.MysqlMock{db.Mysql{DbInstance: nil}}
	view.tree.SetServer(server)

	root := view.tree.treeView.GetRoot()
	if root == nil {
		t.Fatal("Expected root node to exist")
	}

	// Set current node to root
	view.tree.treeView.SetCurrentNode(root)

	// Test navigation sequence: root -> first DB -> back to root
	current := view.tree.treeView.GetCurrentNode()
	if current != root {
		t.Error("Expected current node to be root")
	}

	// Navigate down
	nextNode := view.findNextNode(current)
	if nextNode != nil {
		view.tree.treeView.SetCurrentNode(nextNode)
		
		// Navigate back up
		prevNode := view.findPrevNode(nextNode)
		if prevNode != root {
			t.Error("Expected to navigate back to root")
		}
	}
}