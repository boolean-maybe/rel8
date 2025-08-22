package view

import (
	"context"
	"fmt"
	"log/slog"
	"rel8/db"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TreeNode represents a node in the database tree
type TreeNode struct {
	Type     string // "server", "database", "category", "item"
	Name     string
	Children []*TreeNode
	Parent   *TreeNode
	Data     interface{} // Additional data for the node
}

// Tree wraps a TreeView with database-specific functionality
type Tree struct {
	*tview.TreeView
	server    db.DatabaseServer
	rootNode  *TreeNode
	treeView  *tview.TreeView
	nodes     map[*tview.TreeNode]*TreeNode // Map tview nodes to our nodes
	currentDB string
}

// NewTree creates a new tree view
func NewTree() *Tree {
	treeView := tview.NewTreeView()

	tree := &Tree{
		TreeView: treeView,
		treeView: treeView,
		nodes:    make(map[*tview.TreeNode]*TreeNode),
	}

	// Configure the tree view
	treeView.SetBackgroundColor(Colors.BackgroundDefault)
	treeView.SetBorder(true)
	treeView.SetBorderColor(Colors.BorderDefault)
	treeView.SetBorderPadding(0, 0, 1, 1) // Add internal padding like the grid
	treeView.SetTitle(" Database Explorer ")
	treeView.SetTitleColor(Colors.TitleDefault)
	treeView.SetTitleAlign(tview.AlignCenter)

	// Configure selection highlighting
	treeView.SetGraphicsColor(Colors.SelectionBandBg)

	// Set the selected style to make the selection more visible
	treeView.SetSelectedFunc(tree.onSelectionChanged)

	// Note: Input handling is done at the application level in view.go
	// to allow proper coordination between global shortcuts and tree navigation
	// treeView.SetInputCapture(tree.handleInput) // Removed to fix key handling

	// Set up selection change handler for database tracking
	treeView.SetChangedFunc(tree.onSelectionChanged)

	// Ensure tree is focusable and shows focus
	treeView.SetFocusFunc(func() {
		// Tree gained focus
	})
	treeView.SetBlurFunc(func() {
		// Tree lost focus
	})

	// Note: Input handling is done at the application level in view.go
	// to allow proper coordination between global shortcuts and tree navigation

	return tree
}

// onSelectionChanged is called when the current node selection changes
func (t *Tree) onSelectionChanged(node *tview.TreeNode) {
	if node == nil {
		return
	}

	treeNode, exists := t.nodes[node]
	if !exists {
		return
	}

	// Update current database when a database node is selected
	if treeNode.Type == "database" {
		t.currentDB = treeNode.Name
		slog.Debug("Selection changed to database", "database", t.currentDB)
	}
}

// handleInput handles keyboard navigation and expansion in the tree
func (t *Tree) handleInput(event *tcell.EventKey) *tcell.EventKey {
	currentNode := t.treeView.GetCurrentNode()

	switch event.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		// Navigation handled by tview automatically
	case tcell.KeyEnter:
		// Handle Enter key for expansion/contraction
		if currentNode != nil {
			t.HandleNodeActivation(currentNode)
		}
		return nil // Consume the event
	case tcell.KeyRune:
		switch event.Rune() {
		case ':':
			// Don't consume colon - let it pass through to global handler for command mode
			return event
		case '!':
			// Don't consume exclamation - let it pass through to global handler for SQL mode
			return event
		case 's':
			// Don't consume 's' - let it pass through to global handler for editor mode
			return event
		case ' ': // Space bar also toggles expansion
			if currentNode != nil {
				t.HandleNodeActivation(currentNode)
			}
			return nil // Consume the event
		}
	case tcell.KeyEscape:
		// Don't consume Escape - let it pass through to global handler
		return event
	case tcell.KeyCtrlC:
		// Don't consume Ctrl+C - let it pass through to global handler
		return event
	}

	// Let other keyHelpHeader (arrows, etc.) pass through for normal tree navigation
	return event
}

// HandleNodeActivation handles Enter key or space bar press on a node
func (t *Tree) HandleNodeActivation(node *tview.TreeNode) {
	if node == nil {
		return
	}

	treeNode, exists := t.nodes[node]
	if !exists {
		return
	}

	// Tree node activated

	switch treeNode.Type {
	case "server":
		// Expand/collapse server node (shows/hides databases)
		node.SetExpanded(!node.IsExpanded())

	case "database":
		// Set current database and expand/collapse to show categories
		t.currentDB = treeNode.Name
		node.SetExpanded(!node.IsExpanded())

	case "category":
		// Expand category and load items if not already loaded
		if len(node.GetChildren()) == 0 {
			t.loadCategoryItems(node, treeNode)
		}
		node.SetExpanded(!node.IsExpanded())

	case "item":
		// For items, we don't expand but could trigger detail view
		// This will be handled by the existing state management system
	}
}

// SetServer sets the database server and populates the tree
func (t *Tree) SetServer(server db.DatabaseServer) {
	t.server = server
	t.PopulateTree()
}

// PopulateTree populates the tree with database structure
func (t *Tree) PopulateTree() {
	if t.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get server info for root node
	serverInfo := t.server.GetServerInfo(ctx)
	serverName := serverInfo["version"]
	if serverName == "" {
		serverName = "Database Server"
	}

	// Create root node
	t.rootNode = &TreeNode{
		Type: "server",
		Name: serverName,
	}

	// Create root tree node
	rootTreeNode := tview.NewTreeNode(serverName).
		SetColor(Colors.TreeRootColor).
		SetExpanded(true).
		SetSelectable(true)

	t.nodes[rootTreeNode] = t.rootNode
	t.treeView.SetRoot(rootTreeNode)

	// Populate databases
	t.populateDatabases(ctx, rootTreeNode, t.rootNode)

	// Set initial selection to root node and ensure it's visible
	t.treeView.SetCurrentNode(rootTreeNode)

	// Ensure root is selected and highlighted
	if rootTreeNode != nil {
		rootTreeNode.SetSelectable(true)
	}

	// Tree populated successfully
}

// populateDatabases populates databases under the root
func (t *Tree) populateDatabases(ctx context.Context, parentTreeNode *tview.TreeNode, parentNode *TreeNode) {
	_, databases := t.server.FetchDatabases(ctx)
	if len(databases) == 0 {
		return
	}

	for _, dbData := range databases {
		var dbName string

		// Extract database name based on data type
		switch db := dbData.(type) {
		case db.MysqlDatabase:
			dbName = db.Name
		case db.PostgresDatabase:
			dbName = db.Name
		case map[string]string:
			if name, exists := db["NAME"]; exists {
				dbName = name
			} else if name, exists := db["name"]; exists {
				dbName = name
			}
		default:
			slog.Warn("Unknown database data type", "type", fmt.Sprintf("%T", dbData))
			continue
		}

		if dbName == "" {
			continue
		}

		// Create database node
		dbNode := &TreeNode{
			Type:   "database",
			Name:   dbName,
			Parent: parentNode,
			Data:   dbData,
		}
		parentNode.Children = append(parentNode.Children, dbNode)

		// Create database tree node
		dbTreeNode := tview.NewTreeNode(fmt.Sprintf("📁 %s", dbName)).
			SetColor(Colors.TreeDatabaseColor).
			SetExpanded(false).
			SetSelectable(true)

		t.nodes[dbTreeNode] = dbNode
		parentTreeNode.AddChild(dbTreeNode)

		// Add category nodes under database
		t.addCategoryNodes(dbTreeNode, dbNode, dbName)
	}
}

// addCategoryNodes adds category nodes (Tables, Views, etc.) under a database
func (t *Tree) addCategoryNodes(dbTreeNode *tview.TreeNode, dbNode *TreeNode, dbName string) {
	categories := []struct {
		name  string
		icon  string
		color tcell.Color
	}{
		{"Tables", "📋", Colors.TreeCategoryColor},
		{"Views", "👁", Colors.TreeCategoryColor},
		{"Procedures", "⚙️", Colors.TreeCategoryColor},
		{"Functions", "🔧", Colors.TreeCategoryColor},
		{"Triggers", "⚡", Colors.TreeCategoryColor},
	}

	for _, category := range categories {
		categoryNode := &TreeNode{
			Type:   "category",
			Name:   category.name,
			Parent: dbNode,
			Data:   map[string]string{"database": dbName, "category": category.name},
		}
		dbNode.Children = append(dbNode.Children, categoryNode)

		categoryTreeNode := tview.NewTreeNode(fmt.Sprintf("%s %s", category.icon, category.name)).
			SetColor(category.color).
			SetExpanded(false).
			SetSelectable(true)

		t.nodes[categoryTreeNode] = categoryNode
		dbTreeNode.AddChild(categoryTreeNode)
	}
}

// loadCategoryItems loads items for a category (tables, views, etc.)
func (t *Tree) loadCategoryItems(categoryTreeNode *tview.TreeNode, categoryNode *TreeNode) {
	if categoryNode.Data == nil {
		return
	}

	data, ok := categoryNode.Data.(map[string]string)
	if !ok {
		return
	}

	dbName := data["database"]
	category := data["category"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var items []db.TableData
	var icon string

	switch category {
	case "Tables":
		_, items = t.server.FetchTablesForDatabase(ctx, dbName)
		icon = "📄"
	case "Views":
		_, items = t.server.FetchViewsForDatabase(ctx, dbName)
		icon = "👁‍🗨"
	case "Procedures":
		_, items = t.server.FetchProceduresForDatabase(ctx, dbName)
		icon = "⚙"
	case "Functions":
		_, items = t.server.FetchFunctionsForDatabase(ctx, dbName)
		icon = "🔧"
	case "Triggers":
		_, items = t.server.FetchTriggersForDatabase(ctx, dbName)
		icon = "⚡"
	}

	// Loading category items

	for _, itemData := range items {
		var itemName string

		// Extract item name based on data type
		switch item := itemData.(type) {
		case db.MysqlTable:
			itemName = item.Name
		case db.PostgresTable:
			itemName = item.Name
		case map[string]string:
			if name, exists := item["NAME"]; exists {
				itemName = name
			} else if name, exists := item["name"]; exists {
				itemName = name
			}
		default:
			slog.Warn("Unknown item data type", "type", fmt.Sprintf("%T", itemData))
			continue
		}

		if itemName == "" {
			continue
		}

		// Create item node
		itemNode := &TreeNode{
			Type:   "item",
			Name:   itemName,
			Parent: categoryNode,
			Data:   itemData,
		}
		categoryNode.Children = append(categoryNode.Children, itemNode)

		// Create item tree node
		itemTreeNode := tview.NewTreeNode(fmt.Sprintf("%s %s", icon, itemName)).
			SetColor(Colors.TreeItemColor).
			SetSelectable(true)

		t.nodes[itemTreeNode] = itemNode
		categoryTreeNode.AddChild(itemTreeNode)
	}
}

// GetSelectedNode returns the currently selected tree node
func (t *Tree) GetSelectedNode() *TreeNode {
	currentNode := t.treeView.GetCurrentNode()
	if currentNode == nil {
		return nil
	}

	return t.nodes[currentNode]
}

// GetCurrentDatabase returns the currently selected database
func (t *Tree) GetCurrentDatabase() string {
	return t.currentDB
}

// ExpandAll expands all nodes in the tree
func (t *Tree) ExpandAll() {
	t.expandNode(t.treeView.GetRoot())
}

// expandNode recursively expands a node and its children
func (t *Tree) expandNode(node *tview.TreeNode) {
	if node == nil {
		return
	}

	node.SetExpanded(true)
	for _, child := range node.GetChildren() {
		t.expandNode(child)
	}
}

// CollapseAll collapses all nodes except the root
func (t *Tree) CollapseAll() {
	root := t.treeView.GetRoot()
	if root != nil {
		for _, child := range root.GetChildren() {
			t.collapseNode(child)
		}
	}
}

// collapseNode recursively collapses a node and its children
func (t *Tree) collapseNode(node *tview.TreeNode) {
	if node == nil {
		return
	}

	node.SetExpanded(false)
	for _, child := range node.GetChildren() {
		t.collapseNode(child)
	}
}

// Note: Removed all selection band drawing methods - using tview's built-in selection highlighting

// WrapTree wraps tree with consistent padding matching other components
func WrapTree(tree *Tree) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 0, false). // Left padding
		AddItem(tree.TreeView, 0, 1, true).
		AddItem(nil, 0, 0, false) // Right padding
}
